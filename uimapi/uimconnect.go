package uimapi

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
)

type UIMConnection struct {
	IPAddress   string
	Username    string
	Password    string
	APIEndpoint string
}

type Result struct {
	Target     string       `json:"target"`
	Datapoints [][2]float64 `json:"datapoints"`
}

type TableResult struct {
	Column []Column         `json:"columns"`
	Rows   [][3]interface{} `json:"rows"`
	Type   string           `json:"type"`
}

type Column struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type ApiResponse struct {
	Status        string // e.g. "200 OK"
	StatusCode    int    // e.g. 200
	Body          string
	ContentLength int64
}

type Metric struct {
	Origin          string         `json:"origin"`
	Id              string         `json:"id"`
	Type            string         `json:"type"`
	Self            string         `json:"self"`
	Source          string         `json:"source"`
	Target          string         `json:"target"`
	Probe           string         `json:"probe"`
	ComputerSystem  ComputerSystem `json:"for_computer_system"`
	Device          Device         `json:"for_device"`
	Configuration   Configuration  `json:"for_configuration_item"`
	MinSampleValue  float64        `json:"minSampleValue"`
	MaxSampleValue  float64        `json:"maxSampleValue"`
	MeanSampleValue float64        `json:"meanSampleValue"`
	Sample          []SampleItem   `json:"sample"`
}
type ComputerSystem struct {
	Id   string `json:"id"`
	Self string `json:"self"`
	Name string `json:"name"`
	Ip   string `json:"ip"`
}
type Device struct {
	Id   string `json:"id"`
	Self string `json:"self"`
}
type Configuration struct {
	Id          string `json:"id"`
	Self        string `json:"self"`
	Name        string `json:"name"`
	QosName     string `json:"qosName"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
}

type SampleItem struct {
	Time      string  `json:"time"`
	Epochtime float64 `json:"epochtime"`
	Value     float64 `json:"value"`
	Rate      int     `json:"rate"`
}

var uimConnection UIMConnection

func SetConnection(ipaddress string, username string, password string) {
	uimConnection.IPAddress = ipaddress
	uimConnection.Username = username
	uimConnection.Password = password
	uimConnection.APIEndpoint = "https://" + ipaddress + "/uimapi/metrics"
}

func GetConnectionInfo() UIMConnection {
	return uimConnection
}

func GetQOS(hostname []string) []string {
	hostnameString := ""
	for z := 0; z < len(hostname); z++ {
		if z == 0 {
			hostnameString = hostnameString + hostname[z]
		} else {
			hostnameString = hostnameString + "," + hostname[z]
		}

	}

	uimapi := uimConnection.APIEndpoint + "?id_lookup=by_metric_source&id=" + hostnameString + "&period=latest&showSamples=false"

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, _ := http.NewRequest("GET", uimapi, nil)
	req.SetBasicAuth(uimConnection.Username, uimConnection.Password)
	req.Header.Set("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	var apiResponse ApiResponse
	var uim []Metric
	var qos []string
	if res != nil {
		defer res.Body.Close()

		body, _ := ioutil.ReadAll(res.Body)
		apiResponse.Status = res.Status
		apiResponse.StatusCode = res.StatusCode
		apiResponse.ContentLength = res.ContentLength
		apiResponse.Body = string(body)

		json.Unmarshal(body, &uim)
		for i := 0; i < len(uim); i++ {
			qos = append(qos, uim[i].Configuration.QosName)
		}
	} else {
		fmt.Print(err)
	}
	qos = uniqueNonEmptyElementsOf(qos)
	return qos
}

func GetQosTarget(hostname []string, qosName []string) []string {
	hostnameString := convertString(hostname)
	qosNameString := convertString(qosName)

	uimapi := uimConnection.APIEndpoint + "?id_lookup=by_metric_source&id=" + hostnameString + "&metric_type_lookup=by_metric_name&metricFilter=" + qosNameString + "&period=latest&showSamples=true"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, _ := http.NewRequest("GET", uimapi, nil)

	req.SetBasicAuth(uimConnection.Username, uimConnection.Password)
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	var apiResponse ApiResponse

	var uim []Metric
	var target []string

	if res != nil {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		apiResponse.Status = res.Status
		apiResponse.StatusCode = res.StatusCode
		apiResponse.ContentLength = res.ContentLength
		apiResponse.Body = string(body)

		json.Unmarshal(body, &uim)
		for i := 0; i < len(uim); i++ {
			target = append(target, uim[i].Target)
		}
	} else {
		fmt.Print(err)
	}
	target = uniqueNonEmptyElementsOf(target)
	sort.Strings(target)

	return target

}

func GetQosValue(hostname []string, qosName []string, targetname []string, timeRange string, legend string) []Result {
	hostnameString := convertString(hostname)
	qosNameString := convertString(qosName)
	targetNameString := convertString(targetname)

	if targetNameString == "--alltarget--" {
		targetNameString = ""
	}

	var uimapi string

	if timeRange != "" {
		uimapi = uimConnection.APIEndpoint + "?id_lookup=by_metric_source&id=" + hostnameString + "&metric_type_lookup=by_metric_name&metricFilter=" + qosNameString + "&target=" + targetNameString + "&period=" + timeRange + "&showSamples=true"
	} else {
		uimapi = uimConnection.APIEndpoint + "?id_lookup=by_metric_source&id=" + hostnameString + "&metric_type_lookup=by_metric_name&metricFilter=" + qosNameString + "&target=" + targetNameString + "&period=latest&showSamples=true"
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, _ := http.NewRequest("GET", uimapi, nil)

	req.SetBasicAuth(uimConnection.Username, uimConnection.Password)
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	var apiResponse ApiResponse

	var uim []Metric
	var result []Result
	var resultItem Result
	var datapo [2]float64

	if res != nil {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		apiResponse.Status = res.Status
		apiResponse.StatusCode = res.StatusCode
		apiResponse.ContentLength = res.ContentLength
		apiResponse.Body = string(body)

		json.Unmarshal(body, &uim)
		for i := 0; i < len(uim); i++ {
			if len(uim[i].Sample) > 0 {
				resultItem = Result{}
				for j := len(uim[i].Sample) - 1; j >= 0; j-- {
					datapo[0] = uim[i].Sample[j].Value
					datapo[1] = uim[i].Sample[j].Epochtime * 1000
					resultItem.Datapoints = append(resultItem.Datapoints, datapo)
				}
				if legend == "target" {
					resultItem.Target = uim[i].Target
				} else if legend == "source" {
					resultItem.Target = uim[i].Source
				} else {
					resultItem.Target = uim[i].Source + " " + uim[i].Target
				}

				result = append(result, resultItem)
			}
		}
	} else {
		fmt.Print(err)
	}
	return result
}

func uniqueNonEmptyElementsOf(s []string) []string {
	unique := make(map[string]bool, len(s))
	us := make([]string, len(unique))
	for _, elem := range s {
		if len(elem) != 0 {
			if !unique[elem] {
				us = append(us, elem)
				unique[elem] = true
			}
		}
	}
	return us
}

func convertString(s []string) string {
	result := ""
	for z := 0; z < len(s); z++ {
		if z == 0 {
			result = result + s[z]
		} else {
			result = result + "," + s[z]
		}

	}
	return result
}

func FloatToString(input_num float64) string {
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}
