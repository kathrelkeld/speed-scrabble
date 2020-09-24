package game

import (
	"encoding/json"
	"fmt"
	"log"
)

type MessageType byte

const (
	MsgOK MessageType = iota
	MsgError
	MsgExit
	MsgConnect
	MsgJoinGame
	MsgGameStatus
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
	Data []byte
}

func NewSocketMsg(t MessageType, d interface{}) ([]byte, error) {
	b, err := json.Marshal(d)
	if err != nil {
		log.Println("error marshalling message:", err)
		// TODO handle error
		return nil, nil
	}

	b = append([]byte{byte(t)}, b...)

	return b, nil
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
