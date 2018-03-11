package main

import (
	"crypto/tls"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/golang/glog"
	resty "gopkg.in/resty.v1"
)

type Pod struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name string `json:"name"`
		Uid  string `json:"uid,omitempty"`
	} `json:"metadata"`
	dice []string
}

type PodList struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Items      []Pod  `json:"items"`
}

func (podList *PodList) GetPodsforJob(jobUUID string, userName string) bool {
	var resource = *OpenshiftApiHost + "/api/v1/namespaces/" + *OpenshiftNamespace + "/pods/"
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	token := Users[userName].token
	//time.Sleep(500 * time.Millisecond)
	resp, err := resty.R().
		SetQueryParams(map[string]string{
			"labelSelector": "controller-uid=" + jobUUID,
		}).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		Get("https://" + resource)
	if err != nil || resp.StatusCode() != 200 {
		glog.Warning("Get Pods - Error in Response -", err, "-", resp.StatusCode())
		return false
	}

	err = json.Unmarshal(resp.Body(), podList)
	if err != nil {
		glog.Warning("Get Pods - Error unmarshalling -", err, "-", resp.StatusCode())
		return false
	}

	return true
}

func (pod *Pod) GetLogs(userName string) bool {
	var resource = *OpenshiftApiHost + "/api/v1/namespaces/" + *OpenshiftNamespace + "/pods/" + pod.Metadata.Name + "/log"
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	token := Users[userName].token
	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		Get("https://" + resource)
	if err != nil || resp.StatusCode() != 200 {
		glog.Warning("GetLogs - Error in Response -", err, "-", resp.StatusCode())
		return false
	}
	re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	final := re_leadclose_whtsp.ReplaceAllString(string(resp.Body()), "")
	final = re_inside_whtsp.ReplaceAllString(final, " ")

	pod.dice = strings.Split(final, " ")

	return true
}
