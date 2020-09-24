package main

import (
	"github.com/kathrelkeld/speed-scrabble/game"
)

func handleSocketMsg(t game.MessageType, data []byte) int {
	switch t {
	case game.MsgOK:
	}
	return 0
}
