package msg

import (
	"encoding/json"
	"fmt"
	"log"
)

type Type byte

const (
	OK Type = iota
	Error
	Exit
	Connect
	JoinGame
	GameStatus
	NewGame
	Start
	NewTiles
	AddTile
	SendBoard
	Verify
	Score
	Notify
	Invalid
	GameOver
)

var TypeToString = map[Type]string{
	OK:         "ok",
	Error:      "error",
	Exit:       "exit",
	Connect:    "connect",
	JoinGame:   "joinGame",
	GameStatus: "gameStatus",
	NewGame:    "newGame",
	Start:      "start",
	NewTiles:   "newTiles",
	AddTile:    "addTile",
	SendBoard:  "sendBoard",
	Verify:     "verify",
	Score:      "score",
	Notify:     "notify",
	Invalid:    "invalid",
	GameOver:   "gameOver",
}

func (mt Type) String() string {
	return fmt.Sprintf("\"%v\"", TypeToString[mt])
}

type SocketData struct {
	Type Type
	Data []byte
}

func NewSocketData(t Type, d interface{}) ([]byte, error) {
	b, err := json.Marshal(d)
	if err != nil {
		log.Println("error marshalling message:", err)
		// TODO handle error
		return nil, nil
	}

	b = append([]byte{byte(t)}, b...)

	return b, nil
}

type GameInfo struct {
	GameName    string
	PlayerNames []string
}
