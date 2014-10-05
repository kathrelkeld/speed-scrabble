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

var globalDict map[string]bool = initDictionary("sowpods.txt")

func initDictionary(filename string) map[string]bool {
	log.Println("Initializing dictionary from", filename)
	d := make(map[string]bool)
	f, err := os.Open(filename)
	if err != nil {
		log.Println("err:", err)
		return d
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		word := scanner.Text()
		d[word] = true
	}
	if err := scanner.Err(); err != nil {
		log.Println("err:", err)
		return d
	}
	log.Println("Finished initializing dictionary")
	return d
}

func verifyWord(w string) bool {
	d := globalDict
	_, found := d[w]
	return found
}

var globalGame *Game = makeNewGame(1)

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
	result := g.tiles[g.tilesServed]
	g.tilesServed += 1
	g.m.Unlock()
	return result
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

func verifyBoard(board [][]string) bool {
	maxX := len(board)
	maxY := len(board[0])
	var words []string
	var i, j, k int
	for i = 0; i < maxX; i++ {
		for j = 0; j < maxY; j++ {
			if board[i][j] != "" {
				if i == 0 || board[i-1][j] == "" {
					if !(i == maxX || board[i+1][j] == "") {
						var word string = board[i][j] + board[i+1][j]
						for k = i + 2; k < maxX-(i+2); k++ {
							if board[k][j] != "" {
								word += board[k][j]
							} else {
								break
							}
						}
						words = append(words, word)
					}
				}
				if j == 0 || board[i][j-1] == "" {
					if !(j == maxY || board[i][j+1] == "") {
						var word string = board[i][j] + board[i][j+1]
						for k = j + 2; k < maxY-(j+2); k++ {
							if board[i][k] != "" {
								word += board[i][k]
							} else {
								break
							}
						}
						words = append(words, word)
					}
				}
			}
		}
	}
	log.Println("Found words:", words)
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
