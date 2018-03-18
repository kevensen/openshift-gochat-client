package main

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	batch_v1 "k8s.io/api/batch/v1"
	core_v1 "k8s.io/api/core/v1"
)

type Dice struct {
	numDice  int
	numSides int
	Job      *batch_v1.Job
	userName string
	JobName  string
}

func NewDice(message string, userName string, registry string, namespace string) *Dice {
	var validRoll = regexp.MustCompile(`^//roll-dice(\d+)-sides(\d+)`)
	var parsedRoll = validRoll.FindStringSubmatch(message)
	dice := new(Dice)
	if parsedRoll == nil {
		dice.numSides = 6
		dice.numDice = 1

	} else {
		dice.numDice, _ = strconv.Atoi(parsedRoll[1])
		dice.numSides, _ = strconv.Atoi(parsedRoll[2])
	}
	dice.userName = userName
	dice.JobName = "dice-" + userName
	var diceCommand []string
	diceCommand = append(diceCommand, "/opt/app-root/fortran-app")
	diceCommand = append(diceCommand, strconv.Itoa(dice.numDice))
	diceCommand = append(diceCommand, strconv.Itoa(dice.numSides))

	diceContainer := &core_v1.Container{
		Name:    dice.JobName,
		Image:   registry + "/" + namespace + "/dice",
		Command: diceCommand,
	}
	var backoff int32
	backoff = 4
	job := new(batch_v1.Job)
	job.ObjectMeta.Name = dice.JobName
	job.Spec.Template.ObjectMeta.Name = dice.JobName
	job.Spec.Template.Spec.Containers = append(job.Spec.Template.Spec.Containers, *diceContainer)
	job.Spec.Template.Spec.RestartPolicy = "Never"
	job.Spec.BackoffLimit = &backoff

	dice.Job = job

	return dice
}

func RollDiceHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	userName := r.FormValue("Name")
	websocketDialer := new(websocket.Dialer)
	headers := make(http.Header)
	cookie, _ := r.Cookie("auth")
	headers.Add("Content-Type", "application/json")
	headers.Add("Cookie", "auth="+cookie.Value)
	headers.Add("Origin", "http://"+r.Host)
	diceMessage := new(message)
	diceMessage.Name = userName

	dice := NewDice(r.FormValue("Message"), userName, "", "")
	if dice.exist() {
		diceMessage.Message = "rolled " + strconv.Itoa(dice.numDice) + " " + strconv.Itoa(dice.numSides) + "-sided dice: " + dice.roll()
	} else {
		diceMessage.Message = "has no dice to roll"
	}

	conn, _, err := websocketDialer.Dial("ws://"+r.Host+"/room", headers)
	if err != nil {
		glog.Warningln(err)
	}
	if err := conn.WriteJSON(diceMessage); err != nil {
		log.Println(err)
		return
	}
}

func (dice *Dice) roll() string {
	/*var job *batch_v1.Job
	glog.Infoln("Dice - roll - Looking for job", dice.JobName)
	_, err := APIClientSet.BatchV1().Jobs(*OpenshiftNamespace).Get(dice.JobName, meta_v1.GetOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		glog.Warningln("Dice - roll - Error getting job -", err.Error())
		return "has dice to roll but has trouble rolling dice."
	}

	err = APIClientSet.BatchV1().Jobs(*OpenshiftNamespace).Delete(dice.JobName, &meta_v1.DeleteOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		glog.Warningln("Dice - roll - Error deleting job -", err)
		return "has dice to roll but has trouble rolling dice."
	}
	job, err = APIClientSet.BatchV1().Jobs(*OpenshiftNamespace).Create(dice.Job)
	for err != nil {
		glog.Warningln(err.Error())
		err = nil
		job, err = APIClientSet.BatchV1().Jobs(*OpenshiftNamespace).Create(dice.Job)
		time.Sleep(500 * time.Millisecond)
	}

	listOptions := meta_v1.ListOptions{}
	listOptions.LabelSelector = "controller-uid=" + string(job.ObjectMeta.UID)

	pods, err := APIClientSet.CoreV1().Pods(*OpenshiftNamespace).List(listOptions)
	podName := pods.Items[0].ObjectMeta.Name

	req := APIClientSet.CoreV1().Pods(*OpenshiftNamespace).GetLogs(podName, &core_v1.PodLogOptions{})

	result := req.Do()
	for result.Error() != nil {

		time.Sleep(500 * time.Millisecond)
		result = req.Do()

	}

	body, err := result.Raw()

	re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	final := re_leadclose_whtsp.ReplaceAllString(string(body), "")
	final = re_inside_whtsp.ReplaceAllString(final, " ")

	return final*/
	return "5"

}

/* func (dice *Dice) exist() bool {
	var resource = *OpenshiftApiHost + "/oapi/v1/namespaces/" + *OpenshiftNamespace + "/imagestreams/dice"
	glog.Infoln("Checking for dice at", resource)
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(Users[dice.userName].token).
		Get("https://" + resource)
	if err != nil || resp.StatusCode() != 200 {
		return false
	}

	return true
} */

func (dice *Dice) exist() bool {
	//var resource = *OpenshiftApiHost + "/oapi/v1/namespaces/" + *OpenshiftNamespace + "/imagestreams/dice"

	return true
}
