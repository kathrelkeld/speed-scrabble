package game

import (
	"log"
)

// The GameAssigner manages game assignments.
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

// NewGameAssigner returns a new GameAssigner.
func NewGameAssigner() *GameAssigner {
	return &GameAssigner{
		NewGameChan:  make(chan MsgGameRequest),
		GameExitChan: make(chan *Game),
		games:        make(map[string]*Game),
		quit:         make(chan struct{}),
	}
}

// Run accepts game requests from clients, and creates/destroys games.
func (ga *GameAssigner) Run() {
	for {
		select {
		case req := <-ga.NewGameChan:
			if ga.games["global"] == nil {
				ga.games["global"] = ga.StartNewGame("global")
			}
			log.Println("GameAssigner assigning client to game")
			ga.games["global"].AddPlayer(req.C)
		case game := <-ga.GameExitChan:
			delete(ga.games, game.Name)
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

// StartNewClient creates a new Client with the given websocket connection.
func (ga *GameAssigner) StartNewClient(conn WebsocketConn) *Client {
	c := &Client{
		conn: conn,
		ga:   ga,
	}
	go c.readSocketMsgs()
	return c
}

// StartNewGame is used by the GameAssigner to make a new game.
func (ga *GameAssigner) StartNewGame(name string) *Game {
	game := &Game{
		Name:            name,
		tiles:           newTiles(),
		clients:         make(map[*Client]bool),
		lastScores:      make(map[*Client]*Score),
		toGameChan:      make(chan MsgFromClient),
		startingTileCnt: 12,
		ga:              ga,
		quit:            make(chan struct{}),
	}
	go game.Run()
	return game
}

// A MsgGameRequest is sent from a Client to ask to create or join a new Game.
type MsgGameRequest struct {
	// TODO: allow client to defined desired parameters.
	C *Client
}

type WebsocketConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	Close() error
}
