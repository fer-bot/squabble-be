package controller

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"squabble/form"
	"squabble/models"
	"strings"

	"github.com/gorilla/websocket"
)

var validPath = regexp.MustCompile("^/(room|listen-game-state)/([a-zA-Z0-9]+)")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func ListenGameWS(w http.ResponseWriter, r *http.Request, userid string) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	roomId := m[len(m)-1]

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	c := &models.Connection{Send: make(chan []byte, 256), Ws: ws}
	s := models.Subscription{Conn: c, Room: roomId, UserId: userid}

	models.GetHub().Register <- s
	go s.ReadPump()
}

func Start(w http.ResponseWriter, r *http.Request, userid string) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	roomId := m[len(m)-1]

	roomDetails, exist := models.GetHub().Details[roomId]

	// room has already started or starter not the creator
	if !exist || roomDetails.IsStart || roomDetails.CreatorId != userid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// add one random word
	models.AddWord(roomId)

	// start
	roomDetails.IsStart = true
	models.BroadcastMessage(roomId, "starting-room", []byte{})
}

func Answer(w http.ResponseWriter, r *http.Request, userid string) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	roomId := m[len(m)-1]

	// get word
	var answerRequest form.AnswerRequest
	err := json.NewDecoder(r.Body).Decode(&answerRequest)
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
	if !exist || !details.IsStart || details.Playerproperties[userid].Health <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	playerprop, exist := details.Playerproperties[userid]
	secretprop := details.SecretPlayerproperties[userid]
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

	json.NewEncoder(w).Encode(form.AnswerResponseBuilder(playerprop.AnsStatus[playerprop.CurrRow]))
	if correctAns {
		secretprop.Ans = [6][5]rune{}
		secretprop.WordIdx += 1
		playerprop.CurrRow = 0
		playerprop.AnsStatus = [6][5]string{}
		playerprop.Health += 50
	} else {
		playerprop.Health -= 5
		if playerprop.CurrRow == 6 {
			secretprop.Ans = [6][5]rune{}
			playerprop.AnsStatus = [6][5]string{}
			playerprop.CurrRow = 0
			playerprop.Health -= 50
		}
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(models.GetHub().Details[roomId].Playerproperties)
	models.BroadcastMessage(roomId, "curr-user-data", reqBodyBytes.Bytes())
}
