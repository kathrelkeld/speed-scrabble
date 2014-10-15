package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var globalGame = makeNewGame(1)
var globalClient = makeNewClient(1)

type Message struct {
	Type     string
	ClientId int
	At       time.Time
	Data     json.RawMessage
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWebsocket(w http.ResponseWriter, req *http.Request) {
	log.Println("Handling new client.")
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("error:", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("error:", err)
			return
		}
		var m Message
		err = json.Unmarshal(p, &m)
		if err != nil {
			log.Println("error:", err)
			return
		}
		log.Println("Got message:", m.Type)
	}
}

func handleAddTile(w http.ResponseWriter, req *http.Request) {
	tile := globalClient.getNextTile()
	if tile.Value == "" {
		http.Error(w, "No more tiles!", http.StatusBadRequest)
		return
	}
	log.Println("Sending Tile:", tile)
	sendJSON(tile, w)
}

func handleNewTiles(w http.ResponseWriter, req *http.Request) {
	globalGame = makeNewGame(globalGame.id + 1)
	globalClient.addToGame(globalGame)
	tiles := globalClient.getInitialTiles()
	log.Println("Sending Tiles:", tiles)
	sendJSON(tiles, w)
}

func handleVerifyTiles(w http.ResponseWriter, req *http.Request) {
	var board Board
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&board)
	if err != nil {
		log.Println("error:", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	s := board.scoreBoard(globalClient)
	log.Println("Board score:", s)
	sendJSON(s, w)
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
	http.HandleFunc("/connect", handleWebsocket)

	log.Println("Now listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
