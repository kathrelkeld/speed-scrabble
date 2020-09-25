package game

import (
	"log"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

var CloseGameChan = make(chan MsgFromGame)

type gameState int

const (
	StateInit gameState = iota
	StateRunning
	StateWaitingRoundReady
	StateWaitingScores
	StateOver
)

type GameAssigner struct {
	NewGameChan  chan MsgGameRequest
	GameExitChan chan *Game
	games        map[string]*Game
	exitChan     chan struct{}
}

func StartGameAssigner() *GameAssigner {
	ga := GameAssigner{
		NewGameChan:  make(chan MsgGameRequest),
		GameExitChan: make(chan *Game),
		games:        make(map[string]*Game),
		exitChan:     make(chan struct{}),
	}
	go ga.Run()
	return &ga
}

// Run is the main function of this package, to be called by the server.
// Accept game requests from clients, and create/destroy games.
func (ga *GameAssigner) Run() {
	GlobalGame := StartNewGame(ga.GameExitChan, "global")
	ga.games["global"] = GlobalGame
	for {
		select {
		case req := <-ga.NewGameChan:
			log.Println("GameAssigner assigning client to game")
			req.C.AssignGameChan <- GlobalGame
		case game := <-ga.GameExitChan:
			delete(ga.games, game.name)
		case <-ga.exitChan:
			return
		}
	}
}

func (ga *GameAssigner) Close() {
	ga.exitChan <- struct{}{}
	close(ga.NewGameChan)
	close(ga.GameExitChan)
	close(ga.exitChan)
}

// A MsgGameRequest is sent from a Client to ask to create or join a new Game.
type MsgGameRequest struct {
	// TODO: allow client to defined desired parameters.
	C *Client
}

// A MsgFromClient is sent from a Client to a Game.
type MsgFromClient struct {
	Type msg.Type
	C    *Client
	Data interface{}
}

// A MsgFromGame is sent from a Game to a Client or Run().
type MsgFromGame struct {
	Type msg.Type
	G    *Game
	Data interface{}
}
