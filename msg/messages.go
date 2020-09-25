package msg

import (
	"encoding/json"
	"fmt"
	"log"
)

type Type byte

const (
	// OK is used to acknowlege a previous message.
	// Data: empty.
	OK Type = iota
	// Error means something went wrong in a survivable way.
	// Data: string of any additional info.
	Error
	// Exit means the sender is exiting, either due to request or error.
	// Data: string of any additional info.
	Exit
	// StartRequest is us
	StartRequest
	JoinGame
	GameStatus
	GameReady
	NewGame
	Start
	NewTiles
	AddTile
	SendBoard
	Verify
	Score
	Notify
	Invalid
	OutOfTiles
	GameOver
)

var TypeToString = map[Type]string{
	OK:         "ok",
	Error:      "error",
	Exit:       "exit",
	JoinGame:   "joinGame",
	GameStatus: "gameStatus",
	GameReady:  "gameReady",
	NewGame:    "newGame",
	Start:      "start",
	NewTiles:   "newTiles",
	AddTile:    "addTile",
	SendBoard:  "sendBoard",
	Verify:     "verify",
	Score:      "score",
	Notify:     "notify",
	Invalid:    "invalid",
	OutOfTiles: "outOfTiles",
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
