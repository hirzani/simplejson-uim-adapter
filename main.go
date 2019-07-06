package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	uim "grafanauim/uimapi"
	"strings"

	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
)

type TimeQuery struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type IntervalScopeVar struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

type ScopedVar struct {
	__interval    IntervalScopeVar `json:"__interval"`
	__interval_ms IntervalScopeVar `json:"__interval_ms"`
}

type RangeTimeQuery struct {
	From string    `json:"from"`
	To   string    `json:"to"`
	Raw  TimeQuery `json:"raw"`
}

type TargetsQuery struct {
	Target string `json:"target"`
	RefId  string `json:"refId"`
	Type   string `json:"type"`
}

type GrafanaQuery struct {
	Timezone      string         `json:"timezone"`
	PanelId       int            `json:"panelId"`
	DashboardId   int            `json:"dashboardId"`
	Range         RangeTimeQuery `json:"range"`
	RangeRaw      TimeQuery      `json:"rangeRaw"`
	Interval      string         `json:"interval"`
	IntervalMs    int            `json:"intervalMs"`
	Targets       []TargetsQuery `json:"targets"`
	MaxDataPoints int            `json:"maxDataPoints"`
	ScopedVars    ScopedVar      `json:"scopedVars"`
	AdhocFilters  []string       `json:"adhocFilters"`
}

type Application struct {
	AppName string `json:"AppName"`
}

type Hostname struct {
	Host string `json:"Host"`
}

type GrafanaFilter struct {
	Target string
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.POST("/search/metric", func(c *gin.Context) {
		rawdata, err := c.GetRawData()
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		var grafanaFilter GrafanaFilter
		json.Unmarshal([]byte(rawdata), &grafanaFilter)

		grafanaFilterSplit := strings.Split(grafanaFilter.Target, "|")

		hostFilter := grafanaFilterSplit[2]

		var hostnamelist []string

		hostnamelist = append(hostnamelist, hostFilter)

		listqos := uim.GetQOS(hostnamelist)
		c.JSON(http.StatusOK, listqos)
	})

	r.POST("/search/target", func(c *gin.Context) {
		rawdata, err := c.GetRawData()
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		var grafanaFilter GrafanaFilter
		json.Unmarshal([]byte(rawdata), &grafanaFilter)

		grafanaFilterSplit := strings.Split(grafanaFilter.Target, "|")

		hostFilter := grafanaFilterSplit[2]
		metricFilter := grafanaFilterSplit[0]

		var hostnamelist []string
		var metricList []string

		metricList = append(metricList, metricFilter)
		hostnamelist = append(hostnamelist, hostFilter)

		listTarget := uim.GetQosTarget(hostnamelist, metricList)
		listTarget = append(listTarget, "--alltarget--")
		listTarget = append(listTarget, "[source]")

		c.JSON(http.StatusOK, listTarget)
	})

	r.POST("/query", func(c *gin.Context) {
		rawdata, err := c.GetRawData()
		var grafanaQuery GrafanaQuery
		json.Unmarshal([]byte(rawdata), &grafanaQuery)

		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		var uimgraph []uim.Result

		for _, element := range grafanaQuery.Targets {
			var hostnameList []string
			var metricList []string
			var targetList []string

			//targetType := element.Type
			targetQuery := element.Target
			grafanaFilterSplit := strings.Split(targetQuery, "|")

			hostFilter := grafanaFilterSplit[2]
			metricFilter := grafanaFilterSplit[0]
			targetFilter := grafanaFilterSplit[3]
			legendFilter := grafanaFilterSplit[4]

			hostnameList = append(hostnameList, hostFilter)

			if targetFilter == "[source]" {
				targetList = hostnameList
			} else {
				targetList = append(targetList, targetFilter)
			}

			targetList = append(targetList, targetFilter)
			metricList = append(metricList, metricFilter)

			graphResult := uim.GetQosValue(hostnameList, metricList, targetList, grafanaQuery.Range.From+"|"+grafanaQuery.Range.To, legendFilter)
			for _, aGraphResult := range graphResult {
				uimgraph = append(uimgraph, aGraphResult)
			}
		}

		c.JSON(http.StatusOK, uimgraph)
	})

	r.POST("/search/legend", func(c *gin.Context) {
		listLegend := []string{"source", "target", "source+target"}
		c.JSON(http.StatusOK, listLegend)
	})

	return r
}

func main() {
	r := setupRouter()
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	uimIP := viper.Get("uim.ipaddress").(string)
	uimUsername := viper.Get("uim.username").(string)
	uimPassword := viper.Get("uim.password").(string)

	adapterIP := viper.Get("adapter.ipaddress").(string)
	adapterPort := viper.Get("adapter.port").(string)

	uim.SetConnection(uimIP, uimUsername, uimPassword)
	uimConn := uim.GetConnectionInfo()
	fmt.Println("connect to uim api: " + uimConn.APIEndpoint)
	r.Run(adapterIP + ":" + adapterPort)
}
