package main

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Dice struct {
	numDice    int
	numSides   int
	Job        *batchv1.Job
	userName   string
	JobName    string
	restConfig *rest.Config
}

func NewDice(message string, userName string) *Dice {
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

	diceContainer := &corev1.Container{
		Name:    dice.JobName,
		Image:   *openshiftRegistry + "/" + *openshiftNamespace + "/dice",
		Command: diceCommand,
	}
	var backoff int32
	backoff = 4
	job := new(batchv1.Job)
	job.ObjectMeta.Name = dice.JobName
	job.Spec.Template.ObjectMeta.Name = dice.JobName
	job.Spec.Template.Spec.Containers = append(job.Spec.Template.Spec.Containers, *diceContainer)
	job.Spec.Template.Spec.RestartPolicy = "Never"
	job.Spec.BackoffLimit = &backoff

	dice.Job = job

	dice.restConfig = &rest.Config{
		Host:            *openshiftApiHost,
		BearerToken:     UserTokens[userName],
		TLSClientConfig: rest.TLSClientConfig{Insecure: *allowInsecure},
	}

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

	dice := NewDice(r.FormValue("Message"), userName)
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
	var job *batchv1.Job
	glog.Infoln("Dice - roll - Looking for job", dice.JobName)
	kubeClient, err := kubernetes.NewForConfig(dice.restConfig)
	if err != nil {
		glog.Errorln("Dice - roll - Error creating API client -", err.Error())
		return "has dice to roll but has trouble rolling dice."
	}

	_, err = kubeClient.BatchV1().Jobs(*openshiftNamespace).Get(dice.JobName, metav1.GetOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		glog.Warningln("Dice - roll - Error getting job -", err.Error())
		return "has dice to roll but has trouble rolling dice."
	}

	err = kubeClient.BatchV1().Jobs(*openshiftNamespace).Delete(dice.JobName, &metav1.DeleteOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		glog.Warningln("Dice - roll - Error deleting job -", err)
		return "has dice to roll but has trouble rolling dice."
	}
	job, err = kubeClient.BatchV1().Jobs(*openshiftNamespace).Create(dice.Job)
	for err != nil {
		glog.Warningln(err.Error())
		err = nil
		job, err = kubeClient.BatchV1().Jobs(*openshiftNamespace).Create(dice.Job)
		time.Sleep(500 * time.Millisecond)
	}

	listOptions := metav1.ListOptions{}
	listOptions.LabelSelector = "controller-uid=" + string(job.ObjectMeta.UID)

	pods, err := kubeClient.CoreV1().Pods(*openshiftNamespace).List(listOptions)
	if err != nil {
		glog.Warningln("Dice - roll - Error getting pods -", err)
		return "has dice to roll but has trouble rolling dice."
	}

	for len(pods.Items) == 0 {
		pods, err = kubeClient.CoreV1().Pods(*openshiftNamespace).List(listOptions)
		if err != nil {
			glog.Warningln("Dice - roll - Error getting pods -", err)
			return "has dice to roll but has trouble rolling dice."
		}
		time.Sleep(500 * time.Millisecond)
	}
	podName := pods.Items[0].ObjectMeta.Name

	req := kubeClient.CoreV1().Pods(*openshiftNamespace).GetLogs(podName, &corev1.PodLogOptions{})

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

	return final

}

func (dice *Dice) exist() bool {

	imageV1Client, err := imagev1.NewForConfig(dice.restConfig)
	if err != nil {
		glog.Errorln("Could not connect to OpenShift API:", err)
		return false
	}

	_, err = imageV1Client.ImageStreams(*openshiftNamespace).Get("dice", metav1.GetOptions{})
	if err != nil {
		glog.Errorln("Could not get ImageStream:", err)
		return false

	}

	return true
}
