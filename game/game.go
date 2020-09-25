package game

import (
	"log"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

type Game struct {
	name            string
	tiles           []Tile
	clients         map[*Client]bool
	lastScores      map[*Client]Score
	state           gameState
	startingTileCnt int

	ToGameChan     chan MsgFromClient
	ToAssignerChan chan *Game
	quit           chan struct{}
}

func StartNewGame(toAssignerChan chan *Game, name string) *Game {
	game := &Game{
		name:            name,
		tiles:           newTiles(),
		clients:         make(map[*Client]bool),
		lastScores:      make(map[*Client]Score),
		ToGameChan:      make(chan MsgFromClient),
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
		c.Close()
	}
	close(g.ToGameChan)
	close(g.quit)
}

func (g *Game) AddPlayer(c *Client) {
	log.Println("runGame: Adding client to game")
	g.clients[c] = false
	c.sendSocketMsg(msg.PlayerJoined, nil)
	c.game = g
}

func (g *Game) sendGameInfo() {
	var names []string
	for c := range g.clients {
		names = append(names, c.Name)
	}
	info := msg.GameInfoData{g.name, names}
	for c := range g.clients {
		c.sendSocketMsg(msg.GameInfo, info)
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
		c.sendSocketMsg(t, d)
	}
}

func (g *Game) sendToAllClientsExcept(exc *Client, t msg.Type, d interface{}) {
	log.Printf("Sending %s to all clients.\n", t)
	for c := range g.clients {
		if c != exc {
			c.sendSocketMsg(t, d)
		}
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
		case cm := <-g.ToGameChan:
			log.Println("Game got client message of type:", cm.Type)
			switch cm.Type {
			case msg.RoundReady:
				// Player indicating that they want to start a new round.
				if g.state != StateWaitingRoundReady {
					// If game is not waiting, start a new round and start waiting.
					g.tiles = newTiles()
					g.lastScores = make(map[*Client]Score)
					g.resetClientReply()
					g.state = StateWaitingRoundReady
					g.sendToAllClientsExcept(cm.C, msg.RoundReady, nil)
					// TODO: add timeout
				}
				// Mark this player as ready.
				g.clients[cm.C] = true
				if g.allClientsTrue() {
					g.state = StateRunning
					tiles := g.tiles[:g.startingTileCnt]
					log.Println("Sent tiles:", tiles)
					g.sendToAllClients(msg.Start, tiles)
					for client := range g.clients {
						client.tilesServed = g.startingTileCnt
					}
				}
			case msg.AddTile:
				if g.state != StateRunning {
					cm.C.sendSocketMsg(msg.Error, "Error: no active game!")
				} else {
					cm.C.addTile()
				}
			case msg.Verify:
				if g.state != StateRunning {
					cm.C.sendSocketMsg(msg.Error, "Error: no active game!")
				} else {
					board := cm.Data.([]byte)
					if ok, score := cm.C.checkGameWon(board); ok {
						g.lastScores[cm.C] = score
						g.resetClientReply()
						g.state = StateWaitingScores
						g.clients[cm.C] = true
						g.sendToAllClientsExcept(cm.C, msg.SendBoard, nil)
					}
				}
			case msg.SendBoard:
				if g.state != StateWaitingScores {
					// TODO: should not have gotten this message
					continue
				}
				board := cm.Data.([]byte)
				score := cm.C.ScoreMarshalledBoard(board)
				g.clients[cm.C] = true
				g.lastScores[cm.C] = score
				// TODO: add timeout
				if g.allClientsTrue() {
					//TODO: get scores to determine a winner
					g.sendToAllClients(msg.Score, nil)
					g.state = StateOver
				}
			case msg.Exit:
				cm.C.Close()
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
