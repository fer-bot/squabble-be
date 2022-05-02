package models

import (
	"bytes"
	"encoding/json"
	"log"
	"squabble/utils"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Connection struct {
	Ws   *websocket.Conn
	Send chan []byte
}
type Subscription struct {
	Conn   *Connection
	Room   string
	UserId string
}
type Message struct {
	Data []byte
	Room string
}
type MessageData struct {
	DataType string
	Data     []byte
}

type PlayerProperties struct {
	Health    int
	AnsStatus [6][5]string
	CurrRow   int
}

type SecretPlayerProperties struct {
	Ans     [6][5]rune
	WordIdx int
}

type SinglePlayerData struct {
	Playerproperty         PlayerProperties
	SecretPlayerproperties SecretPlayerProperties
	CurrWord               string
	IsStart                bool
}

type RoomDetails struct {
	Playerproperties       map[string]PlayerProperties
	SecretPlayerproperties map[string]SecretPlayerProperties
	WordList               []string
	WordListCount          int
	IsStart                bool
	CreatorId              string
	CreatedAt              time.Time
}

type hub struct {
	Rooms      map[string]map[*Connection]bool
	Details    map[string]RoomDetails
	Broadcast  chan Message
	Register   chan Subscription
	Unregister chan Subscription
}

var h = hub{
	Broadcast:  make(chan Message),
	Register:   make(chan Subscription),
	Unregister: make(chan Subscription),
	Rooms:      make(map[string]map[*Connection]bool),
	Details:    make(map[string]RoomDetails),
}

func GetHub() *hub {
	return &h
}

func (h *hub) Run() {
	for {
		select {
		case s := <-h.Register:
			connections := h.Rooms[s.Room]
			if connections == nil {
				connections = make(map[*Connection]bool)
				h.Rooms[s.Room] = connections
				h.Details[s.Room] = RoomDetails{IsStart: false, CreatorId: s.UserId, CreatedAt: time.Now()}
			}
			playerproperty, exist := h.Details[s.Room].Playerproperties[s.UserId]
			if !exist {
				playerproperty = PlayerProperties{Health: 100}
			}
			secretplayerproperty := h.Details[s.Room].SecretPlayerproperties[s.UserId]
			if !exist {
				secretplayerproperty = SecretPlayerProperties{WordIdx: 0}
			}
			h.Rooms[s.Room][s.Conn] = true

			reqBodyBytes := new(bytes.Buffer)
			// send one user's data
			singlePlayerData := SinglePlayerData{Playerproperty: playerproperty, SecretPlayerproperties: secretplayerproperty, IsStart: h.Details[s.Room].IsStart, CurrWord: ""}
			if singlePlayerData.IsStart {
				singlePlayerData.CurrWord = h.Details[s.Room].WordList[secretplayerproperty.WordIdx]
			}
			json.NewEncoder(reqBodyBytes).Encode(singlePlayerData)
			json.NewEncoder(reqBodyBytes).Encode(MessageData{DataType: "single", Data: reqBodyBytes.Bytes()})
			s.Conn.Send <- reqBodyBytes.Bytes()

			// broadcast new users' data
			json.NewEncoder(reqBodyBytes).Encode(h.Details[s.Room].Playerproperties)
			BroadcastMessage(s.Room, "curr-users-data", reqBodyBytes.Bytes())

		case s := <-h.Unregister:
			connections := h.Rooms[s.Room]
			if connections != nil {
				if _, ok := connections[s.Conn]; ok {
					delete(connections, s.Conn)
					close(s.Conn.Send)
					if !h.Details[s.Room].IsStart {
						delete(h.Details[s.Room].Playerproperties, s.UserId)
						delete(h.Details[s.Room].SecretPlayerproperties, s.UserId)
						if h.Details[s.Room].CreatorId == s.UserId {
							delete(h.Rooms, s.Room)
							delete(h.Details, s.Room)
							continue
						}
					}
					if len(connections) == 0 {
						delete(h.Rooms, s.Room)
						delete(h.Details, s.Room)
					}
				}
			}

		case m := <-h.Broadcast:
			connections := h.Rooms[m.Room]
			for c := range connections {
				select {
				case c.Send <- m.Data:
				default:
					close(c.Send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.Rooms, m.Room)
						delete(h.Details, m.Room)
					}
				}
			}
		}
	}
}

func (s Subscription) ReadPump() {
	c := s.Conn
	defer func() {
		h.Unregister <- s
		c.Ws.Close()
	}()
	c.Ws.SetReadLimit(maxMessageSize)
	c.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Ws.SetPongHandler(func(string) error { c.Ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		m := Message{msg, s.Room}
		h.Broadcast <- m
	}
}

func BroadcastMessage(roomId string, messageType string, messageByte []byte) {
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(MessageData{DataType: messageType, Data: messageByte})
	h.Broadcast <- Message{Room: roomId, Data: reqBodyBytes.Bytes()}
}

func AddWord(roomId string) {
	word, err := utils.GetRandomWord(5)
	if err != nil {
		BroadcastMessage(roomId, "error", []byte{})
		return
	}
	roomDetails := h.Details[roomId]
	roomDetails.WordList = append(roomDetails.WordList, word)
	roomDetails.WordListCount += 1
}
