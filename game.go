package main

import (
	"log"
	"math/rand"
	"sync"
)

type Game struct {
	id          int
	tiles       Tiles
	tilesServed int
	maxScore    int
	mu          sync.Mutex
}

func makeNewGame(id int) *Game {
	g := Game{}
	g.id = id
	g.tiles = newTiles()
	g.mu = sync.Mutex{}
	return &g
}

func (g *Game) getInitialTiles() Tiles {
	tiles := g.tiles[:12]
	score := 0
	for _, elt := range tiles {
		score += pointValues[elt]
	}
	g.mu.Lock()
	g.tilesServed = 12
	g.maxScore = score
	g.mu.Unlock()
	return g.tiles[:12]
}

func (g *Game) getNextTile() string {
	if g.tilesServed == len(g.tiles) {
		log.Println("No tiles remaining to send!")
		return ""
	}
	g.mu.Lock()
	g.tilesServed += 1
	tile := g.tiles[g.tilesServed-1]
	g.maxScore += pointValues[tile]
	g.mu.Unlock()
	return tile
}

func (g *Game) getTilesServedCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.tilesServed
}

func (g *Game) getAllTilesServed() Tiles {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.tiles[:g.tilesServed]
}

func (g *Game) getMaxScore() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.maxScore
}

type Tiles []string

func newTiles() Tiles {
	var tiles Tiles
	for k, v := range freqMap {
		for j := 0; j < v; j++ {
			tiles = append(tiles, k)
		}
	}
	for i := range tiles {
		j := rand.Intn(i + 1)
		tiles[i], tiles[j] = tiles[j], tiles[i]
	}
	return tiles
}

type Client struct {
	id   int
	game *Game
	mu   sync.Mutex
}

func makeNewClient(id int) *Client {
	c := Client{}
	c.id = id
	c.game = globalGame
	c.mu = sync.Mutex{}
	return &c
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
