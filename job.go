package main

import (
	"regexp"

	"strconv"

	"github.com/golang/glog"
)

type Job struct {
	numDice  int
	numSides int
	userName string
}

func NewJob(message string, userName string) *Job {
	glog.Infoln(message)
	var validRoll = regexp.MustCompile(`^//roll-dice(\d)-sides(\d)`)
	var parsedRoll = validRoll.FindStringSubmatch(message)
	job := new(Job)
	if parsedRoll == nil {
		job.numSides = 6
		job.numDice = 1
		return job
	}
	job.numDice, _ = strconv.Atoi(parsedRoll[1])
	job.numSides, _ = strconv.Atoi(parsedRoll[2])
	return job

}
