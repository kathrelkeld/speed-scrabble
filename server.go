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
	Data     *json.RawMessage
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

func (c *Client) readMessage() (Message, error) {
	var m Message
	_, p, err := c.conn.ReadMessage()
	if err != nil {
		log.Println("error:", err)
		return m, err
	}
	err = json.Unmarshal(p, &m)
	if err != nil {
		log.Println("error:", err)
		return m, err
	}
	return m, nil
}

func (c *Client) sendMessage(t string, d interface{}) error {
	marshaledData, err := json.Marshal(d)
	if err != nil {
		log.Println("error:", err)
		return err
	}
	raw := json.RawMessage(marshaledData)
	m := Message{Type: t, ClientId: c.id, At: time.Now(), Data: &raw}
	j, err := json.Marshal(m)
	if err != nil {
		log.Println("error:", err)
		return err
	}
	err = c.conn.WriteMessage(websocket.TextMessage, j)
	if err != nil {
		log.Println("error:", err)
		return err
	}
	return nil
}

func (c *Client) runClient() {
	for {
		m, err := c.readMessage()
		if err != nil {
			return
		}
		switch {
		case m.Type == "addTile":
			tile := c.getNextTile()
			if tile.Value == "" {
				//TODO: send out of tiles error
				return
			}
			c.sendMessage("tile", tile)
		case m.Type == "verify":
			var board Board
			err := json.Unmarshal([]byte(*m.Data), &board)
			if err != nil {
				log.Println("error:", err)
				return
			}
			s := board.scoreBoard(globalClient)
			c.sendMessage("score", s)
		}
	}
}

func handleWebsocket(w http.ResponseWriter, req *http.Request) {
	log.Println("Handling new client.")
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("error:", err)
		return
	}
	c := globalClient
	c.conn = conn
	m, err := c.readMessage()
	if err != nil {
		return
	}
	if m.Type != "new" {
		log.Println("Incorrect type for new connection!")
		return
	}
	go c.runClient()
}

func handleNewTiles(w http.ResponseWriter, req *http.Request) {
	globalGame = makeNewGame(globalGame.id + 1)
	globalClient.addToGame(globalGame)
	tiles := globalClient.getInitialTiles()
	log.Println("Sending Tiles:", tiles)
	sendJSON(tiles, w)
}

func main() {
	const addr = "localhost:8080"
	fileserver := http.FileServer(http.Dir("public"))
	redirect := http.RedirectHandler("public/scrabble.html", http.StatusFound)

	http.Handle("/", redirect)
	http.Handle("/public/", http.StripPrefix("/public/", fileserver))
	http.HandleFunc("/tiles", handleNewTiles)
	http.HandleFunc("/connect", handleWebsocket)

	log.Println("Now listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
