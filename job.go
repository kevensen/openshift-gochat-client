package main

import (
	"crypto/tls"
	"strings"
	"time"

	"encoding/json"
	"strconv"

	"github.com/golang/glog"
	resty "gopkg.in/resty.v1"
)

type JobList struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Items      []Job  `json:"items"`
}

type Job struct {
	userName   string
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name string `json:"name"`
		Uid  string `json:"uid,omitempty"`
	} `json:"metadata"`
	Spec struct {
		Template struct {
			Spec struct {
				Containers    []container `json:"containers"`
				RestartPolicy string      `json:"restartPolicy"`
			} `json:"spec"`
		} `json:"template"`
		BackoffLimit int `json:"backOffLimit"`
	} `json:"spec"`
	Status struct {
		Succeeded int `json:"succeeded,omitempty"`
	} `json:"status,omitempty"`
}

type container struct {
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	Command []string `json:"command"`
}

func NewJob(numDice int, numSides int, userName string) *Job {
	job := new(Job)
	job.userName = userName
	job.ApiVersion = "batch/v1"
	job.Kind = "Job"
	job.Metadata.Name = "dice-" + userName

	diceContainer := new(container)
	diceContainer.Name = "dice-" + userName
	diceContainer.Image = *OpenshiftRegistry + "/" + *OpenshiftNamespace + "/dice"
	diceContainer.Command = append(diceContainer.Command, "/opt/dice")
	diceContainer.Command = append(diceContainer.Command, strconv.Itoa(numDice))
	diceContainer.Command = append(diceContainer.Command, strconv.Itoa(numSides))

	job.Spec.Template.Spec.Containers = append(job.Spec.Template.Spec.Containers, *diceContainer)
	job.Spec.Template.Spec.RestartPolicy = "Never"
	job.Spec.BackoffLimit = 4

	return job
}

func (job *Job) Roll() string {
	for job.exists() {
		glog.Infoln("Job exists.  Attempting delete")
		job.delete()
		time.Sleep(200 * time.Millisecond)
	}
	if !job.create() {

		return "has dice to roll but has trouble rolling dice."
	}
	var podList = new(PodList)
	for len(podList.Items) == 0 {
		podList.GetPodsforJob(job.Metadata.Uid, job.userName)
	}

	for !job.completed() {
		time.Sleep(1000 * time.Millisecond)
	}

	podList.Items[0].GetLogs(job.userName)

	return "rolled " + string(job.Spec.Template.Spec.Containers[0].Command[1]) +
		", " + string(job.Spec.Template.Spec.Containers[0].Command[2]) + " sided dice - " +
		strings.Join(podList.Items[0].dice, ", ")
}

func (job *Job) create() bool {
	var resource = *OpenshiftApiHost + "/apis/batch/v1/namespaces/" + *OpenshiftNamespace + "/jobs"
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	token := Users[job.userName].token

	openshiftJob, err := json.Marshal(job)
	if err != nil {
		glog.Warning("Create Job - Error Marshalling -", err)
		return false
	}
	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		SetBody(openshiftJob).
		Post("https://" + resource)
	if err != nil || resp.StatusCode() > 299 {
		glog.Warning("Create Job - Error in Response -", err, "-", resp.StatusCode())
		return false
	}
	err = json.Unmarshal(resp.Body(), &job)
	if err != nil {
		glog.Warning("Create Job - Unmarshalling -", err)
		return false
	}

	return true
}

func (job *Job) completed() bool {
	var resource = *OpenshiftApiHost + "/apis/batch/v1/namespaces/" + *OpenshiftNamespace + "/jobs"
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	token := Users[job.userName].token

	openshiftJob, err := json.Marshal(job)
	if err != nil {
		glog.Warning("Create Job - Error Marshalling -", err)
		return false
	}
	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		SetBody(openshiftJob).
		Get("https://" + resource)
	if err != nil || resp.StatusCode() > 299 {
		glog.Warning("Get Job - Error in Response -", err, "-", resp.StatusCode())
		return false
	}
	jobList := new(JobList)
	err = json.Unmarshal(resp.Body(), &jobList)
	if err != nil {
		glog.Warning("Get Job - Unmarshalling -", err)
		return false
	}
	if jobList.Items[0].Status.Succeeded == 0 {
		return false
	}

	return true
}

func (job *Job) delete() bool {
	var resource = *OpenshiftApiHost + "/apis/batch/v1/namespaces/" + *OpenshiftNamespace + "/jobs/dice-" + job.userName
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	token := Users[job.userName].token

	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		Delete("https://" + resource)
	if err != nil || (resp.StatusCode() != 200 && resp.StatusCode() != 404) {
		glog.Warning("Delete Job - Error in Response -", err, "-", resp.StatusCode())
		return false
	}

	return true
}

func (job *Job) exists() bool {
	var resource = *OpenshiftApiHost + "/apis/batch/v1/namespaces/" + *OpenshiftNamespace + "/jobs/dice-" + job.userName
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	token := Users[job.userName].token
	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		Get("https://" + resource)
	if err != nil || (resp.StatusCode() != 200 && resp.StatusCode() != 404) {
		glog.Warning("Job Exists - Error in Response -", err, "-", resp.StatusCode())
		return false
	} else if resp.StatusCode() == 404 {
		glog.Infoln("Job Exists - Job not found.  Returning False")
		return false
	}

	return true
}
