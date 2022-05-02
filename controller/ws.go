package controller

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"squabble/form"
	"squabble/models"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func ListenGameWS(w http.ResponseWriter, r *http.Request, roomId string, sessionId string) {
	username, err := models.GetUsername(sessionId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	c := &models.Connection{Send: make(chan []byte, 256), Ws: ws}
	s := models.Subscription{Conn: c, Room: roomId, UserId: username}

	models.GetHub().Register <- s
	go s.WritePump()
}

func Start(w http.ResponseWriter, r *http.Request, roomId string) {
	username, err := models.SessionAuth(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
		return
	}

	roomDetails, exist := models.GetHub().Details[roomId]

	// room has already started or starter not the creator
	if !exist || roomDetails.IsStart || roomDetails.CreatorId != username {
		w.WriteHeader(http.StatusBadRequest)
		err = errors.New("room has already started")
		json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
		return
	}

	// add one random word
	models.AddWord(roomId)
	roomDetails = models.GetHub().Details[roomId]

	// set start to true
	roomDetails.IsStart = true
	models.GetHub().Details[roomId] = roomDetails

	// broadcast new all player properties
	models.BroadcastMessage(roomId, "starting-room", []byte{})
}

func Answer(w http.ResponseWriter, r *http.Request, roomId string) {
	username, err := models.SessionAuth(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
		return
	}

	// get word
	var answerRequest form.AnswerRequest
	err = json.NewDecoder(r.Body).Decode(&answerRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
		return
	}

	// change to lower, validate
	answer := strings.ToLower(answerRequest.Word)
	if len(answer) != 5 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check if room and player valid
	details, exist := models.GetHub().Details[roomId]
	if !exist || !details.IsStart || details.Playerproperties[username].Health <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	playerprop, exist := details.Playerproperties[username]
	secretprop := details.SecretPlayerproperties[username]
	if !exist {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//add word list
	if secretprop.WordIdx == details.WordListCount {
		models.AddWord(roomId)
	}

	correctAns := true
	currWord := []rune(details.WordList[secretprop.WordIdx])
	wordMap := make(map[rune]bool)
	for _, char := range currWord {
		wordMap[char] = true
	}

	for pos, char := range answer {
		secretprop.Ans[playerprop.CurrRow][pos] = char
		if currWord[pos] == char {
			playerprop.AnsStatus[playerprop.CurrRow][pos] = "full"
			continue
		}
		if _, ok := wordMap[char]; ok {
			playerprop.AnsStatus[playerprop.CurrRow][pos] = "half"
			correctAns = false
			continue
		}
		playerprop.AnsStatus[playerprop.CurrRow][pos] = "zero"
		correctAns = false
	}

	// return result
	json.NewEncoder(w).Encode(form.AnswerResponseBuilder(playerprop.AnsStatus[playerprop.CurrRow]))

	// set new player properties and secret properties
	if correctAns {
		secretprop.Ans = [6][5]rune{}
		secretprop.WordIdx += 1
		playerprop.CurrRow = 0
		playerprop.AnsStatus = [6][5]string{}
		playerprop.Health += 50
	} else {
		playerprop.Health -= 5
		playerprop.CurrRow += 1
		if playerprop.CurrRow == 6 {
			secretprop.Ans = [6][5]rune{}
			playerprop.AnsStatus = [6][5]string{}
			playerprop.CurrRow = 0
			playerprop.Health -= 50
		}
	}
	details.Playerproperties[username] = playerprop
	details.SecretPlayerproperties[username] = secretprop

	// broadcast new all player properties
	models.BroadcastMessage(roomId, "curr-user-data", models.GetHub().Details[roomId].Playerproperties)
}
