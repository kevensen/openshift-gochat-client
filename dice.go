package main

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

type Dice struct {
	numDice  int
	numSides int
}

func NewDice(message string) *Dice {
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
	return dice
}

func RollDice(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	userName := r.FormValue("Name")
	user := Users[userName]
	websocketDialer := new(websocket.Dialer)
	headers := make(http.Header)
	cookie, _ := r.Cookie("auth")
	headers.Add("Content-Type", "application/json")
	headers.Add("Cookie", "auth="+cookie.Value)
	headers.Add("Origin", "http://"+r.Host)
	diceMessage := new(message)
	diceMessage.Name = userName

	dice := NewDice(r.FormValue("Message"))
	diceMessage.Message = user.RollDice(dice)

	conn, _, err := websocketDialer.Dial("ws://"+r.Host+"/room", headers)
	if err != nil {
		glog.Warningln(err)
	}
	if err := conn.WriteJSON(diceMessage); err != nil {
		log.Println(err)
		return
	}
}
