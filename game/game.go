package game

import (
	"log"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

type Game struct {
	name            string
	tiles           []Tile
	clients         map[*Client]bool
	state           gameState
	startingTileCnt int

	ToGameChan     chan MsgFromClient
	ToAssignerChan chan *Game
	AddPlayerChan  chan *Client
	quit           chan struct{}
}

func StartNewGame(toAssignerChan chan *Game, name string) *Game {
	game := &Game{
		name:            name,
		tiles:           newTiles(),
		clients:         make(map[*Client]bool),
		ToGameChan:      make(chan MsgFromClient),
		AddPlayerChan:   make(chan *Client),
		ToAssignerChan:  toAssignerChan,
		startingTileCnt: 12,
		quit:            make(chan struct{}),
	}
	go game.Run()
	return game
}

// Close game: notify GameAssigner and any clients; close any active channels or go routines.
func (g *Game) Close() {
	g.state = StateOver
	g.ToAssignerChan <- g
	for c := range g.clients {
		c.ToClientChan <- MsgFromGame{msg.Exit, g, nil}
	}
	close(g.ToGameChan)
	close(g.AddPlayerChan)
	close(g.quit)
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
	for {
		select {
		case c := <-g.AddPlayerChan:
			log.Println("runGame: Adding client to game")
			g.clients[c] = false
			c.ToClientChan <- MsgFromGame{msg.PlayerJoined, g, nil}
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
			case msg.Score:
				score := cm.Data.(Score)
				if g.state != StateWaitingScores {
					if !score.Win {
						// TODO: should not have gotten first score from a non-winning hand
					}
					g.resetClientReply()
					g.sendToAllClients(msg.RoundOver, nil)
					g.state = StateWaitingScores
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
					g.Close()
				}
			}
		case <-g.quit:
			return
		}
	}
}
