package game

import (
	"github.com/kathrelkeld/speed-scrabble/msg"
)

type MsgGameRequest struct {
	Type msg.Type
	C    *Client
}

type MsgFromClient struct {
	Type msg.Type
	C    *Client
	Data interface{}
}

type MsgFromGame struct {
	Type msg.Type
	G    *Game
	data interface{}
}

type WebSocketConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
}
