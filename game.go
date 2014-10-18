package main

import (
	"github.com/gorilla/websocket"
	"math/rand"
)

type Game struct {
	tiles Tiles
	info  GameInfo
}

type GameInfo struct {
	addPlayerChan chan ClientInfo
	toGameChan    chan FromClientMsg
}

func makeNewGame() *Game {
	g := Game{}
	g.tiles = newTiles()
	g.info = GameInfo{}
	g.info.addPlayerChan = make(chan ClientInfo)
	g.info.toGameChan = make(chan FromClientMsg)
	return &g
}

func (g *Game) newTiles() {
	g.tiles = newTiles()
}

type Client struct {
	conn             *websocket.Conn
	tilesServedCount int
	tilesServed      Tiles
	maxScore         int
	toGameChan       chan FromClientMsg
	info             ClientInfo
}

type ClientInfo struct {
	toClientChan   chan FromGameMsg
	newTileChan    chan NewTileMsg
	assignGameChan chan GameInfo
}

func makeNewClient() *Client {
	c := Client{}
	c.info = ClientInfo{}
	c.info.toClientChan = make(chan FromGameMsg)
	c.info.newTileChan = make(chan NewTileMsg)
	c.info.assignGameChan = make(chan GameInfo)
	c.newGame()
	return &c
}

func (c *Client) newGame() {
	c.tilesServedCount = 0
	c.tilesServed = Tiles{}
	c.maxScore = 0
}

func (c *Client) addTile(t Tile) {
	c.tilesServed = append(c.tilesServed, t)
	c.tilesServedCount += 1
	c.maxScore += t.Points
}

//func (c *Client) getInitialTiles() Tiles {
//tiles := c.game.tiles[:12]
//score := 0
//for _, elt := range tiles {
//score += elt.Points
//}
//c.mu.Lock()
//c.tilesServed = 12
//c.maxScore = score
//c.mu.Unlock()
//return tiles
//}

//func (c *Client) getNextTile() Tile {
//if c.tilesServed == len(c.game.tiles) {
//log.Println("No tiles remaining to send!")
//return Tile{Value: "", Points: 0}
//}
//c.mu.Lock()
//tile := c.game.tiles[c.tilesServed]
//c.tilesServed += 1
//c.maxScore += tile.Points
//c.mu.Unlock()
//return tile
//}

func (c *Client) getTilesServedCount() int {
	return c.tilesServedCount
}

func (c *Client) getAllTilesServed() Tiles {
	return c.tilesServed[:c.tilesServedCount]
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
