package main

import (
	"bufio"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
)

var globalDict = initDictionary("sowpods.txt")

func initDictionary(filename string) map[string]struct{} {
	log.Println("Initializing dictionary from", filename)
	d := make(map[string]struct{})
	f, err := os.Open(filename)
	if err != nil {
		log.Println("err:", err)
		return d
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		word := scanner.Text()
		d[word] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		log.Println("err:", err)
		return d
	}
	log.Println("Finished initializing dictionary")
	return d
}

func verifyWord(w string) bool {
	_, ok := globalDict[w]
	return ok
}

var globalGame = makeNewGame(1)

type Game struct {
	id          int
	tiles       []string
	tilesServed int
	m           sync.Mutex
}

func makeNewGame(id int) *Game {
	return &Game{id: id, tiles: newTiles(), tilesServed: 0, m: sync.Mutex{}}
}

func (g *Game) getInitialTiles() []string {
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
	return g.tiles[g.tilesServed]
}

type Vec struct {
	x int
	y int
}

func (a Vec) add(b Vec) Vec {
	return Vec{a.x + b.x, a.y + b.y}
}

func newTiles() []string {
	var tiles []string
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

type Board [][]string

// Find the tiles connected to the given tile.
func (b Board) findConnectedTiles(c map[Vec]*[]Vec, tile Vec, found *[]Vec) {
	_, ok := c[tile]
	if !ok && tile.x < len(b) && tile.y < len(b[0]) &&
		tile.x >= 0 && tile.y >= 0 && b[tile.x][tile.y] != "" {
		c[tile] = found
		*found = append(*found, tile)
		b.findConnectedTiles(c, tile.add(Vec{1, 0}), found)
		b.findConnectedTiles(c, tile.add(Vec{0, 1}), found)
		b.findConnectedTiles(c, tile.add(Vec{-1, 0}), found)
		b.findConnectedTiles(c, tile.add(Vec{0, -1}), found)
	}
}

// Find the connected components
func (b Board) findComponents() [][]Vec {
	result := [][]Vec{}
	maxX := len(b)
	maxY := len(b[0])
	var i, j int
	c := make(map[Vec]*[]Vec)
	for i = 0; i < maxX; i++ {
		for j = 0; j < maxY; j++ {
			tile := Vec{i, j}
			if _, ok := c[tile]; !ok && b[i][j] != "" {
				var thisCompFound []Vec
				b.findConnectedTiles(c, tile, &thisCompFound)
				result = append(result, thisCompFound)
			}
		}
	}
	log.Println("Found", len(result), "components")
	log.Println(result)
	return result
}

// Verify that the given component list has valid words.
func (b Board) verifyComponent(v []Vec) bool {
	// Component must be 2 or more tiles.
	if len(v) <= 1 {
		return false
	}
	return true
}

func verifyBoard(board [][]string) bool {
	findComponents(board)
	maxX := len(board)
	maxY := len(board[0])
	var i, j, k int
	for i = 0; i < maxX; i++ {
		for j = 0; j < maxY; j++ {
			if board[i][j] != "" {
				if i == 0 || board[i-1][j] == "" {
					if !(i == maxX || board[i+1][j] == "") {
						var word string = board[i][j] + board[i+1][j]
						for k = i + 2; k < maxX; k++ {
							if board[k][j] != "" {
								word += board[k][j]
							} else {
								break
							}
						}
						if !verifyWord(word) {
							return false
						}
					}
				}
				if j == 0 || board[i][j-1] == "" {
					if !(j == maxY || board[i][j+1] == "") {
						var word string = board[i][j] + board[i][j+1]
						for k = j + 2; k < maxY; k++ {
							if board[i][k] != "" {
								word += board[i][k]
							} else {
								break
							}
						}
						if !verifyWord(word) {
							return false
						}
					}
				}
			}
		}
	}
	return true
}

func sendJSON(v interface{}, w http.ResponseWriter) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println("error:", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func handleAddTile(w http.ResponseWriter, req *http.Request) {
	tile := globalGame.getNextTile()
	if tile == "" {
		http.Error(w, "No more tiles!", http.StatusBadRequest)
		return
	}
	log.Println("Sending Tile:", tile)
	sendJSON(tile, w)
}

func handleNewTiles(w http.ResponseWriter, req *http.Request) {
	globalGame = makeNewGame(globalGame.id + 1)
	tiles := globalGame.getInitialTiles()
	log.Println("Sending Tiles:", tiles)
	sendJSON(tiles, w)
}

func handleVerifyTiles(w http.ResponseWriter, req *http.Request) {
	var board []([]string)
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&board)
	if err != nil {
		log.Println("error:", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	sendJSON(verifyBoard(board), w)
}

func main() {
	const addr = "localhost:8080"
	fileserver := http.FileServer(http.Dir("public"))
	redirect := http.RedirectHandler("public/scrabble.html", http.StatusFound)

	http.Handle("/", redirect)
	http.Handle("/public/", http.StripPrefix("/public/", fileserver))
	http.HandleFunc("/tiles", handleNewTiles)
	http.HandleFunc("/add_tile", handleAddTile)
	http.HandleFunc("/verify", handleVerifyTiles)

	log.Println("Now listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
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
