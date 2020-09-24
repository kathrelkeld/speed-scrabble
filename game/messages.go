package game

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type MessageType byte

const (
	MsgOK MessageType = iota
	MsgError
	MsgExit
	MsgConnect
	MsgJoinGame
	MsgGameStatus
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
	MsgOK:         "ok",
	MsgError:      "error",
	MsgExit:       "exit",
	MsgConnect:    "connect",
	MsgJoinGame:   "joinGame",
	MsgGameStatus: "gameStatus",
	MsgGlobal:     "global",
	MsgNewGame:    "newGame",
	MsgStart:      "start",
	MsgNewTiles:   "newTiles",
	MsgAddTile:    "addTile",
	MsgSendBoard:  "sendBoard",
	MsgVerify:     "verify",
	MsgScore:      "score",
	MsgGameOver:   "gameOver",
}

func (mt MessageType) String() string {
	return fmt.Sprintf("\"%v\"", MessageTypeToString[mt])
}

type GameRequest struct {
	Type MessageType
	C    *Client
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
	typ  MessageType
	c    *Client
	data interface{}
}

type FromGameMsg struct {
	typ  MessageType
	g    *Game
	data interface{}
}

type GameStatus struct {
	GameName    string
	PlayerNames []string
}

type WebSocketConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
}
