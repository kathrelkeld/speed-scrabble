package game

import (
	"log"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

type MsgGameRequest struct {
	C *Client
}

type MsgFromClient struct {
	Type msg.Type
	C    *Client
	Data interface{}
}

type MsgFromGame struct {
	Type msg.Type
	G    *Game
	Data interface{}
}

type WebSocketConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
}

func AddAPlayerToAGame() {
	for {
		r := <-NewGameChan
		log.Println("NewGameChan: Adding client to game")
		//TODO allow for multiple games
		GlobalGame.AddPlayerChan <- r.C
		r.C.AssignGameChan <- GlobalGame
	}
}
