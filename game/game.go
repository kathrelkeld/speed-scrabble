package game

import (
	"log"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

var GlobalGame = NewGame()
var NewGameChan = make(chan MsgGameRequest)

type Game struct {
	name      string
	tiles     []Tile
	clients   map[*Client]bool
	isRunning bool

	//Accessible by other routines; Not allowed to change.
	ToGameChan    chan MsgFromClient
	AddPlayerChan chan *Client
}

func NewGame() *Game {
	g := Game{}
	g.tiles = newTiles()
	g.isRunning = false
	g.clients = make(map[*Client]bool)
	g.ToGameChan = make(chan MsgFromClient)
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

func (g *Game) sendToAllClients(t msg.Type) {
	log.Println("Sending to all clients.")
	for c := range g.clients {
		c.ToClientChan <- MsgFromGame{t, g, nil}
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

func (g *Game) Run() {
	for {
		select {
		case c := <-g.AddPlayerChan:
			log.Println("runGame: Adding client to game")
			g.clients[c] = false
			g.sendGameStatus()
		case cm := <-g.ToGameChan:
			log.Println("Game got client message of type:", cm.Type)
			switch cm.Type {
			case msg.Start:
				if g.isRunning { //Game is already running!
					cm.C.ToClientChan <- MsgFromGame{msg.Error, g, nil}
					continue
				}
				g.sendToAllClients(msg.NewGame)
				g.hearFromAllClients(msg.Start)
				g.newTiles()
				g.sendToAllClients(msg.Start)
				g.isRunning = true
			case msg.NewTiles:
				if !g.isRunning {
					cm.C.ToClientChan <- MsgFromGame{msg.Error, g, nil}
					continue
				}
				//TODO: handle confirm from all games
				m := MsgFromGame{msg.OK, g, g.tiles[:12]}
				cm.C.ToClientChan <- m
			case msg.AddTile:
				if !g.isRunning {
					cm.C.ToClientChan <- MsgFromGame{msg.Error, g, "Game is not running"}
					continue
				}
				if cm.C.TilesServedCount >= len(g.tiles) {
					cm.C.ToClientChan <- MsgFromGame{msg.Error, g, "No more tiles to serve"}
					continue
				}
				m := MsgFromGame{msg.OK, g,
					g.tiles[cm.C.TilesServedCount]}
				cm.C.ToClientChan <- m
			case msg.GameOver:
				//TODO: get scores to determine a winner
				g.sendToAllClients(msg.GameOver)
				g.hearFromAllClients(msg.OK)
				g.isRunning = false
			case msg.Exit:
				delete(g.clients, cm.C)
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
		//TODO allow for multiple games
		GlobalGame.AddPlayerChan <- r.C
		r.C.AssignGameChan <- GlobalGame
	}
}
