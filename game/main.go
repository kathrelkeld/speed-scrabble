package game

import (
	"log"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

// A GameAssigner manages game assignments.
// TODO make game assigner keep track of active clients too.
type GameAssigner struct {
	// Used by Clients to send in game requests.
	NewGameChan chan MsgGameRequest
	// Used by Games to indicate they're closing.
	GameExitChan chan *Game
	// Map of id -> running games.
	games map[string]*Game
	// Used to cleanly exit server.
	quit chan struct{}
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
			ga.games["global"].AddPlayer(req.C)
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
