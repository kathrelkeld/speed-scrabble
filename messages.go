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
	MsgNewTiles
	MsgAddTile
	MsgVerify
	MsgScore
	MsgGameOver
)

var MessageTypeToString = map[MessageType]string{
	MsgOK:       "ok",
	MsgError:    "error",
	MsgExit:     "exit",
	MsgConnect:  "connect",
	MsgJoinGame: "joinGame",
	MsgGlobal:   "global",
	MsgNewTiles: "newTiles",
	MsgAddTile:  "addTile",
	MsgVerify:   "verify",
	MsgScore:    "score",
	MsgGameOver: "gameOver",
}

func (mt *MessageType) MarshalJSON() ([]byte, error) {
	s := MessageTypeToString[*mt]
	if s == "" {
		log.Println("No such message string to marshal!")
		panic("No such message string to marshal!")
	}
	return json.Marshal(s)
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
	log.Println("No such message string to unmarshal!")
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

type FromClientMsg struct {
	typ   MessageType
	cInfo ClientInfo
}

type FromGameMsg struct {
	typ   MessageType
	gInfo GameInfo
}

type NewTileMsg struct {
	typ MessageType
	t   Tile
}
