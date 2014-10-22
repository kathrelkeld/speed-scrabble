package main

import (
	"math/rand"
)

type Game struct {
	tiles       []Tile
	clientChans map[chan FromGameMsg]bool
	isRunning   bool
	info        GameInfo
}

type GameInfo struct {
	toGameChan    chan FromClientMsg
	addPlayerChan chan ClientInfo
}

func makeNewGame() *Game {
	g := Game{}
	g.tiles = newTiles()
	g.isRunning = false
	g.clientChans = make(map[chan FromGameMsg]bool)
	g.info = GameInfo{}
	g.info.toGameChan = make(chan FromClientMsg)
	g.info.addPlayerChan = make(chan ClientInfo)
	return &g
}

func (g *Game) cleanup() {
	close(g.info.toGameChan)
	close(g.info.addPlayerChan)
}

func (g *Game) newTiles() {
	g.tiles = newTiles()
}

type Client struct {
	conn        WebSocketConn
	socketChan  chan SocketMsg
	tilesServed []Tile
	maxScore    int
	toGameChan  chan FromClientMsg
	validGame   bool
	running     bool
	info        ClientInfo
}

type ClientInfo struct {
	tilesServedCount int
	toClientChan     chan FromGameMsg
	assignGameChan   chan GameInfo
}

func makeNewClient() *Client {
	c := Client{}
	c.socketChan = make(chan SocketMsg)
	c.validGame = false
	c.running = false
	c.info = ClientInfo{}
	c.info.toClientChan = make(chan FromGameMsg)
	c.info.assignGameChan = make(chan GameInfo)
	return &c
}

func (c *Client) cleanup() {
	close(c.socketChan)
	close(c.info.toClientChan)
	close(c.info.assignGameChan)
	c.running = false
}

func (c *Client) newTiles(t []Tile) {
	c.info.tilesServedCount = len(t)
	c.tilesServed = t
	c.maxScore = 0
	for _, elt := range t {
		c.maxScore += elt.Points
	}
}

func (c *Client) addTile(t Tile) {
	c.tilesServed = append(c.tilesServed, t)
	c.info.tilesServedCount += 1
	c.maxScore += t.Points
}

func (c *Client) getTilesServedCount() int {
	return c.info.tilesServedCount
}

func (c *Client) getAllTilesServed() []Tile {
	return c.tilesServed[:c.info.tilesServedCount]
}

func (c *Client) getMaxScore() int {
	return c.maxScore
}

type Tile struct {
	Value  string
	Points int
}

func (t Tile) String() string {
	return t.Value
}

func newTiles() []Tile {
	var tiles []Tile
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
