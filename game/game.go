package game

import (
	"log"
)

var GlobalGame = NewGame()
var NewGameChan = make(chan GameRequest)

type Game struct {
	name      string
	tiles     []Tile
	clients   map[*Client]bool
	isRunning bool

	//Accessible by other routines; Not allowed to change.
	ToGameChan    chan FromClientMsg
	AddPlayerChan chan *Client
}

func NewGame() *Game {
	g := Game{}
	g.tiles = newTiles()
	g.isRunning = false
	g.clients = make(map[*Client]bool)
	g.ToGameChan = make(chan FromClientMsg)
	g.AddPlayerChan = make(chan *Client)
	return &g
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
	status := GameStatus{g.name, names}
	for c := range g.clients {
		c.ToClientChan <- FromGameMsg{MsgGameStatus, g, status}
	}
}

func (g *Game) allClientsTrue() bool {
	result := true
	for _, value := range g.clients {
		result = result && value
	}
	return result
}

func (g *Game) sendToAllClients(t MessageType) {
	log.Println("Sending to all clients.")
	for c := range g.clients {
		c.ToClientChan <- FromGameMsg{t, g, nil}
	}
}

func (g *Game) hearFromAllClients(t MessageType) {
	log.Println("Hearing from all clients.")
	for c := range g.clients {
		g.clients[c] = false
	}
	for !g.allClientsTrue() {
		cm := <-g.ToGameChan
		if cm.typ != t {
			cm.c.ToClientChan <- FromGameMsg{MsgError, g, nil}
		}
		g.clients[cm.c] = true
	}
}

func (g *Game) Run() {
	for {
		select {
		case c := <-g.AddPlayerChan:
			log.Println("runGame: Adding client to game")
			g.clients[c] = false
			g.sendGameStatus()
		case cm := <-g.ToGameChan:
			log.Println("Game got client message of type:", cm.typ)
			switch cm.typ {
			case MsgStart:
				if g.isRunning { //Game is already running!
					cm.c.ToClientChan <- FromGameMsg{MsgError, g, nil}
					continue
				}
				g.sendToAllClients(MsgNewGame)
				g.hearFromAllClients(MsgStart)
				g.newTiles()
				g.sendToAllClients(MsgStart)
				g.isRunning = true
			case MsgNewTiles:
				if !g.isRunning {
					cm.c.ToClientChan <- FromGameMsg{MsgError, g, nil}
					continue
				}
				//TODO: handle confirm from all games
				m := FromGameMsg{MsgOK, g, g.tiles[:12]}
				cm.c.ToClientChan <- m
			case MsgAddTile:
				if !g.isRunning {
					cm.c.ToClientChan <- FromGameMsg{MsgError, g, nil}
					continue
				}
				if cm.c.TilesServedCount >= len(g.tiles) {
					cm.c.ToClientChan <- FromGameMsg{MsgError, g, nil}
					continue
				}
				m := FromGameMsg{MsgOK, g,
					g.tiles[cm.c.TilesServedCount]}
				cm.c.ToClientChan <- m
			case MsgGameOver:
				//TODO: get scores to determine a winner
				g.sendToAllClients(MsgGameOver)
				g.hearFromAllClients(MsgOK)
				g.isRunning = false
			case MsgExit:
				delete(g.clients, cm.c)
				log.Println("runGame: Removing client from game")
				if len(g.clients) == 0 {
					g.isRunning = false
				}
			}
		}
	}
}

func AddAPlayerToAGame() {
	for {
		r := <-NewGameChan
		log.Println("NewGameChan: Adding client to game")
		GlobalGame.AddPlayerChan <- r.C
		r.C.AssignGameChan <- GlobalGame
	}
}
