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

// A GameAssigner manages game assignments.
// TODO make game assigner keep track of active clients too.
type GameAssigner struct {
	NewGameChan  chan MsgGameRequest
	GameExitChan chan *Game
	games        map[string]*Game
	quit         chan struct{}
}

// StartGameAssigner starts and returns a new GameAssigner.
func StartGameAssigner() *GameAssigner {
	ga := GameAssigner{
		NewGameChan:  make(chan MsgGameRequest),
		GameExitChan: make(chan *Game),
		games:        make(map[string]*Game),
		quit:         make(chan struct{}),
	}
	go ga.Run()
	return &ga
}

// Run accepts game requests from clients, and creates/destroys games.
func (ga *GameAssigner) Run() {
	for {
		select {
		case req := <-ga.NewGameChan:
			if ga.games["global"] == nil {
				ga.games["global"] = StartNewGame(ga.GameExitChan, "global")
			}
			log.Println("GameAssigner assigning client to game")
			req.C.AssignGameChan <- ga.games["global"]
		case game := <-ga.GameExitChan:
			delete(ga.games, game.name)
		case <-ga.quit:
			return
		}
	}
}

// Close gracefully shuts down an active GameAssigner.
// TODO: close any active games or clients
func (ga *GameAssigner) Close() {
	close(ga.quit)
	close(ga.NewGameChan)
	close(ga.GameExitChan)
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
