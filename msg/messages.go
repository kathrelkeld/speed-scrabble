package msg

import (
	"encoding/json"
	"fmt"
	"log"
)

type Type byte

const (
	// Error means something went wrong in a survivable way.
	// Data: string of any additional info.
	// Exit means the sender is exiting, either due to request or error.
	// Data: string of any additional info.
	Exit Type = iota
	Error
	JoinGame
	GameInfo
	RoundReady
	Start
	AddTile
	SendBoard
	Verify
	Score
	Invalid
	OutOfTiles
	PlayerJoined
	Result
)

var TypeToString = map[Type]string{
	Exit:         "exit",
	Error:        "error",
	JoinGame:     "joinGame",
	GameInfo:     "gameInfo",
	RoundReady:   "roundReady",
	Start:        "start",
	AddTile:      "addTile",
	SendBoard:    "sendBoard",
	Verify:       "verify",
	Score:        "score",
	Invalid:      "invalid",
	OutOfTiles:   "outOfTiles",
	PlayerJoined: "playerJoined",
	Result:       "result",
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

type GameInfoData struct {
	GameName    string
	PlayerNames []string
}
