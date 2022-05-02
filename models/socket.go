package models

import (
	"bytes"
	"encoding/json"
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

type MessageData[T any] struct {
	DataType string
	Data     T
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
	Register   chan Subscription
	Unregister chan Subscription
}

var h = hub{
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
				h.Details[s.Room] = RoomDetails{
					IsStart: false, 
					CreatorId: s.UserId, 
					CreatedAt: time.Now(), 
					WordList: make([]string, 0), 
					Playerproperties: make(map[string]PlayerProperties),
					SecretPlayerproperties: make(map[string]SecretPlayerProperties),
				}
			}

			roomDetails := h.Details[s.Room]
			playerproperty, exist := roomDetails.Playerproperties[s.UserId]
			if !exist {
				playerproperty = PlayerProperties{Health: 100}
				roomDetails.Playerproperties[s.UserId] = playerproperty
			}
			secretplayerproperty, exist := roomDetails.SecretPlayerproperties[s.UserId]
			if !exist {
				secretplayerproperty = SecretPlayerProperties{WordIdx: 0}
				roomDetails.SecretPlayerproperties[s.UserId] = secretplayerproperty
			}
			h.Rooms[s.Room][s.Conn] = true
		
			// send one user's data
			singlePlayerData := SinglePlayerData{
				Playerproperty: playerproperty, 
				SecretPlayerproperties: secretplayerproperty, 
				IsStart: roomDetails.IsStart, 
				CurrWord: ""
			}
			if roomDetails.IsStart {
				singlePlayerData.CurrWord = roomDetails.WordList[secretplayerproperty.WordIdx]
			}
			respBodyBytes := new(bytes.Buffer)
			json.NewEncoder(respBodyBytes).Encode(MessageData[SinglePlayerData]{DataType: "single", Data: singlePlayerData})
			s.Conn.Send <- respBodyBytes.Bytes()

			// broadcast new users' data
			BroadcastMessage(s.Room, "curr-users-data", roomDetails.Playerproperties)

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
		}
	}
}

// write writes a message with the given message type and payload.
func (c *Connection) Write(mt int, payload []byte) error {
	c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (s *Subscription) WritePump() {
	c := s.Conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// broadcast to all connections
func BroadcastMessage[T any](roomId string, messageType string, messageData T) {
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(MessageData[T]{DataType: messageType, Data: messageData})

	connections := h.Rooms[roomId]
	for c := range connections {
		select {
		case c.Send <- reqBodyBytes.Bytes():
		default:
			close(c.Send)
			delete(connections, c)
			if len(connections) == 0 {
				delete(h.Rooms, roomId)
				delete(h.Details, roomId)
			}
		}
	}
}

// add new word to the list
func AddWord(roomId string) {
	word, err := utils.GetRandomWord(5)
	if err != nil {
		BroadcastMessage(roomId, "error", []byte{})
		return
	}
	roomDetails := h.Details[roomId]
	roomDetails.WordList = append(roomDetails.WordList, word)
	roomDetails.WordListCount += 1
	h.Details[roomId] = roomDetails
}
