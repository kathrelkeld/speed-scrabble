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
	m           sync.Mutex
}

func makeNewGame(id int) *Game {
	return &Game{id: id, tiles: newTiles(), tilesServed: 0, m: sync.Mutex{}}
}

func (g *Game) getInitialTiles() Tiles {
	g.m.Lock()
	g.tilesServed = 12
	g.m.Unlock()
	return g.tiles[:12]
}

func (g *Game) getNextTile() string {
	if g.tilesServed == len(g.tiles) {
		log.Println("No tiles remaining to send!")
		return ""
	}
	g.m.Lock()
	defer g.m.Unlock()
	g.tilesServed += 1
	return g.tiles[g.tilesServed - 1]
}

func (g *Game) getTilesServedCount() int {
	g.m.Lock()
	defer g.m.Unlock()
	return g.tilesServed
}

func (g *Game) getAllTilesServed() Tiles {
	g.m.Lock()
	defer g.m.Unlock()
	return g.tiles[:g.tilesServed]
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
