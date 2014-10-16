package main

import (
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"sync"
)

const (
	READY = iota
	RUNNING
	ENDED
)

type Game struct {
	id      int
	state   int
	tiles   Tiles
	players ClientSet
	mu      sync.Mutex
}

func (g *Game) isRunning() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.state == RUNNING
}

func makeNewGame(id int) *Game {
	g := Game{}
	g.id = id
	g.tiles = newTiles()
	g.state = READY
	g.players = make(ClientSet)
	g.mu = sync.Mutex{}
	return &g
}

type Client struct {
	id          int
	conn        *websocket.Conn
	game        *Game
	tilesServed int
	maxScore    int
	mu          sync.Mutex
}

func makeNewClient(id int) *Client {
	c := Client{}
	c.id = id
	c.game = globalGame
	c.mu = sync.Mutex{}
	return &c
}

func (c *Client) addToGame(g *Game) {
	c.mu.Lock()
	c.game = g
	c.mu.Unlock()
	g.mu.Lock()
	g.players.insert(c)
	g.mu.Unlock()
}

func (c *Client) getInitialTiles() Tiles {
	tiles := c.game.tiles[:12]
	score := 0
	for _, elt := range tiles {
		score += elt.Points
	}
	c.mu.Lock()
	c.tilesServed = 12
	c.maxScore = score
	c.mu.Unlock()
	return tiles
}

func (c *Client) getNextTile() Tile {
	if c.tilesServed == len(c.game.tiles) {
		log.Println("No tiles remaining to send!")
		return Tile{Value: "", Points: 0}
	}
	c.mu.Lock()
	tile := c.game.tiles[c.tilesServed]
	c.tilesServed += 1
	c.maxScore += tile.Points
	c.mu.Unlock()
	return tile
}

func (c *Client) getTilesServedCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tilesServed
}

func (c *Client) getAllTilesServed() Tiles {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.game.tiles[:c.tilesServed]
}

func (c *Client) getMaxScore() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.maxScore
}

type ClientSet map[*Client]struct{}

func (s ClientSet) contains(elt *Client) bool {
	_, ok := s[elt]
	return ok
}

func (s ClientSet) insert(elt *Client) {
	s[elt] = struct{}{}
}

type Tile struct {
	Value  string
	Points int
}

func (t Tile) String() string {
	return t.Value
}

type Tiles []Tile

func newTiles() Tiles {
	var tiles Tiles
	for k, v := range freqMap {
		for j := 0; j < v; j++ {
			tile := Tile{Value: k, Points: pointValues[k]}
			tiles = append(tiles, tile)
		}
	}
	for i := range tiles {
		j := rand.Intn(i + 1)
		tiles[i], tiles[j] = tiles[j], tiles[i]
	}
	return tiles
}

var freqMap = map[string]int{
	"A": 13,
	"B": 3,
	"C": 3,
	"D": 6,
	"E": 18,
	"F": 3,
	"G": 4,
	"H": 3,
	"I": 12,
	"J": 2,
	"K": 2,
	"L": 5,
	"M": 3,
	"N": 8,
	"O": 11,
	"P": 3,
	"Q": 2,
	"R": 9,
	"S": 6,
	"T": 9,
	"U": 6,
	"V": 3,
	"W": 3,
	"X": 2,
	"Y": 3,
	"Z": 2,
}

var pointValues = map[string]int{
	"A": 1,
	"B": 3,
	"C": 3,
	"D": 2,
	"E": 1,
	"F": 4,
	"G": 2,
	"H": 4,
	"I": 1,
	"J": 8,
	"K": 5,
	"L": 1,
	"M": 3,
	"N": 1,
	"O": 1,
	"P": 3,
	"Q": 10,
	"R": 1,
	"S": 1,
	"T": 1,
	"U": 1,
	"V": 4,
	"W": 4,
	"X": 8,
	"Y": 4,
	"Z": 10,
}
