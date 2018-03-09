package main

import (
	"crypto/tls"
	"encoding/json"

	"github.com/golang/glog"
	resty "gopkg.in/resty.v1"
)

type OpenShiftAuth struct {
	token          string
	apiHost        string
	relayHost      string
	clientHostname string
}

func (ocp *OpenShiftAuth) login(token string) (*User, error, int) {
	var user = new(User)
	glog.Infoln("Obtaining user from", ocp.apiHost+"/oapi/v1/users/~")
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		Get("https://" + ocp.apiHost + "/oapi/v1/users/~")

	glog.Infoln("Status Code:", resp.StatusCode())

	if err != nil {
		glog.Errorln("Error", err)
	}
	err = json.Unmarshal(resp.Body(), &user)
	if err != nil {
		glog.Errorln("Error", err)
	}
	return user, nil, resp.StatusCode()
}

func (ocp *OpenShiftAuth) getProject() string {

    glog.Infoln("Obtaining prject name from", ocp.apiHost+"/oapi/v1/projects/~")
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(ocp.token).
		Get("https://" + ocp.apiHost + "/oapi/v1/projects/~")

	glog.Infoln("Status Code:", resp.StatusCode())
	glog.Infoln("Body:", string(resp.Body()))

	if err != nil {
		glog.Errorln("Error", err)
	}
	return "rolled"
}
