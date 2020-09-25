package game

import (
	"log"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

var GlobalGame = NewGame()
var NewGameChan = make(chan MsgGameRequest)

type gameState int

const (
	StateInit gameState = iota
	StateRunning
	StateWaitingRoundReady
	StateWaitingScores
	StateOver
)

type Game struct {
	name            string
	tiles           []Tile
	clients         map[*Client]bool
	state           gameState
	startingTileCnt int

	//Accessible by other routines; Not allowed to change.
	ToGameChan    chan MsgFromClient
	AddPlayerChan chan *Client
}

func NewGame() *Game {
	return &Game{
		tiles:           newTiles(),
		clients:         make(map[*Client]bool),
		ToGameChan:      make(chan MsgFromClient),
		AddPlayerChan:   make(chan *Client),
		startingTileCnt: 12,
	}
}

func (g *Game) Cleanup() {
	close(g.ToGameChan)
	close(g.AddPlayerChan)
}

func (g *Game) newTiles() {
	g.tiles = newTiles()
}

func (g *Game) sendGameStatus() {
	var names []string
	for c := range g.clients {
		names = append(names, c.Name)
	}
	status := msg.GameInfo{g.name, names}
	for c := range g.clients {
		c.ToClientChan <- MsgFromGame{msg.GameStatus, g, status}
	}
}

func (g *Game) allClientsTrue() bool {
	result := true
	for _, value := range g.clients {
		result = result && value
	}
	return result
}

func (g *Game) sendToAllClients(t msg.Type, d interface{}) {
	log.Printf("Sending %s to all clients.\n", t)
	for c := range g.clients {
		c.ToClientChan <- MsgFromGame{t, g, d}
	}
}

func (g *Game) hearFromAllClients(t msg.Type) {
	log.Println("Hearing from all clients.")
	for c := range g.clients {
		g.clients[c] = false
	}
	for !g.allClientsTrue() {
		cm := <-g.ToGameChan
		if cm.Type != t {
			cm.C.ToClientChan <- MsgFromGame{msg.Error, g, nil}
		}
		g.clients[cm.C] = true
	}
}

func (g *Game) resetClientReply() {
	for c := range g.clients {
		g.clients[c] = false
	}
}

func (g *Game) Run() {
	defer g.Cleanup()
	for {
		select {
		case c := <-g.AddPlayerChan:
			log.Println("runGame: Adding client to game")
			g.clients[c] = false
		case cm := <-g.ToGameChan:
			log.Println("Game got client message of type:", cm.Type)
			switch cm.Type {
			case msg.RoundReady:
				// Player is initiating a new game.
				g.newTiles()
				g.resetClientReply()
				g.state = StateWaitingRoundReady
				g.sendToAllClients(msg.RoundReady, nil)
			case msg.Start:
				// Player is ready to start playing.
				if g.state != StateWaitingRoundReady {
					cm.C.ToClientChan <- MsgFromGame{msg.Error, g, "unexpected message"}
					continue
				}
				g.clients[cm.C] = true
				// TODO: add timeout
				if g.allClientsTrue() {
					g.state = StateRunning
					g.sendToAllClients(msg.Start, nil)
				}
			case msg.RoundOver:
				g.resetClientReply()
				g.sendToAllClients(msg.RoundOver, nil)
				g.state = StateWaitingScores
			case msg.Score:
				if g.state != StateWaitingScores {
					cm.C.ToClientChan <- MsgFromGame{msg.Error, g, "unexpected message"}
					continue
				}
				g.clients[cm.C] = true
				// TODO: add timeout
				if g.allClientsTrue() {
					g.state = StateRunning
					//TODO: get scores to determine a winner
					g.sendToAllClients(msg.Score, nil)
					g.state = StateOver
				}
			case msg.Exit:
				delete(g.clients, cm.C)
				log.Println("runGame: Removing client from game")
				if len(g.clients) == 0 {
					g.state = StateOver
					return
				}
			}
		}
	}
}
