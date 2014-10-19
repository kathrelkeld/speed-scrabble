package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type MessageType int

const (
	MsgOK MessageType = iota
	MsgError
	MsgExit
	MsgConnect
	MsgJoinGame
	MsgGlobal
	MsgNewGame
	MsgStart
	MsgNewTiles
	MsgAddTile
	MsgSendBoard
	MsgVerify
	MsgScore
	MsgGameOver
)

var MessageTypeToString = map[MessageType]string{
	MsgOK:        "ok",
	MsgError:     "error",
	MsgExit:      "exit",
	MsgConnect:   "connect",
	MsgJoinGame:  "joinGame",
	MsgGlobal:    "global",
	MsgNewGame:   "newGame",
	MsgStart:     "start",
	MsgNewTiles:  "newTiles",
	MsgAddTile:   "addTile",
	MsgSendBoard: "sendBoard",
	MsgVerify:    "verify",
	MsgScore:     "score",
	MsgGameOver:  "gameOver",
}

func (mt MessageType) MarshalJSON() ([]byte, error) {
	s := MessageTypeToString[mt]
	if s == "" {
		panic("No such message string to marshal!")
	}
	j, err := json.Marshal(s)
	return j, err
	//return json.Marshal(s)
}

func (mt *MessageType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	for key, value := range MessageTypeToString {
		if value == s {
			*mt = key
			return nil
		}
	}
	panic("No such message string to unmarshal!")
}

func (mt MessageType) String() string {
	return fmt.Sprintf("\"%v\"", MessageTypeToString[mt])
}

type GameRequest struct {
	Type  MessageType
	cInfo ClientInfo
}

type SocketMsg struct {
	Type MessageType
	At   time.Time
	Data *json.RawMessage
}

func newSocketMsg(t MessageType, d interface{}) (SocketMsg, error) {
	marshaledData, err := json.Marshal(d)
	if err != nil {
		log.Println("error:", err)
		return SocketMsg{}, err
	}
	raw := json.RawMessage(marshaledData)
	m := SocketMsg{Type: t, At: time.Now(), Data: &raw}
	return m, nil
}

type FromClientMsg struct {
	typ   MessageType
	cInfo ClientInfo
}

type FromGameMsg struct {
	typ   MessageType
	gInfo GameInfo
}

type ReqTileMsg struct {
	typ   MessageType
	index int
	cInfo ClientInfo
}

type NewTileMsg struct {
	typ   MessageType
	tiles Tiles
	gInfo GameInfo
}

type WebSocketConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
}
