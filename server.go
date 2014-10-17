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
var newGameChan = make(chan GameRequest)

type GameRequest struct {
	Type MessageType
	Client *Client
	Chan chan *Game
}

type MessageType string

type WsMessage struct {
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

func (c *Client) readWsMessage() (WsMessage, error) {
	var m WsMessage
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
	log.Println("Got websocket message of type", m.Type)
	return m, nil
}

func (c *Client) sendWsMessage(t string, d interface{}) error {
	marshaledData, err := json.Marshal(d)
	if err != nil {
		log.Println("error:", err)
		return err
	}
	raw := json.RawMessage(marshaledData)
	m := WsMessage{Type: t, ClientId: c.id, At: time.Now(), Data: &raw}
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
	log.Println("Sent websocket message of type", m.Type)
	return nil
}

func (c *Client) readWsMessages(ch chan WsMessage) {
	for {
		m, err := c.readWsMessage()
		if err != nil {
			ch <- WsMessage{Type: "exit"}
			return
		}
		ch <- m
	}
}

func (c *Client) runClient(socketChan chan WsMessage) {
	defer func(){c.gameChan <- MessageType("exit")}()
	responseChan := make(chan *Game)
	c.sendWsMessage("ok", nil)
	for {
		select {
		case m := <- socketChan:
			switch m.Type {
			case "joinGame":
				newGameChan <- GameRequest{"global", c, responseChan}
				c.game = <- responseChan
				c.sendWsMessage("ok", nil)
			case "newTiles":
				c.gameChan <- MessageType("newTiles")
				_ = <- c.gameChan
				tiles := c.getInitialTiles()
				log.Println("Sending tiles:", tiles)
				c.sendWsMessage("tiles", tiles)
			case "addTile":
				tile := c.getNextTile()
				if tile.Value == "" {
					//TODO: send out of tiles error
					return
				}
				c.sendWsMessage("tile", tile)
			case "verify":
				var board Board
				err := json.Unmarshal([]byte(*m.Data), &board)
				if err != nil {
					log.Println("error:", err)
					return
				}
				s := board.scoreBoard(globalClient)
				c.sendWsMessage("score", s)
			case "exit":
				return
			}
		case gm := <- c.gameChan:
			log.Println("Game told client:", gm)
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
	m, err := c.readWsMessage()
	if err != nil {
		return
	}
	if m.Type != "newClient" {
		log.Println("Incorrect type for new connection!")
		return
	}
	socketChan := make(chan WsMessage)
	go c.runClient(socketChan)
	go c.readWsMessages(socketChan)
}

func addPlayerToAGame() {
	for {
		r := <- newGameChan
		log.Println("newGameChan: Adding client to game")
		globalGame.addPlayerChan <- r.Client
		r.Chan <- globalGame
	}
}

type ClientMessage struct {
	typ MessageType
	clientChan chan MessageType
}

func receiveClientMessages(gameChan chan ClientMessage,
		clientChan, endChan chan MessageType) {
	for {
		select {
		case _ = <- endChan:
			log.Println("ClientMessage: Got exit signal! Ending loop!")
			break
		case m := <- clientChan:
			log.Println("Got a client message: ", m)
			gameChan <- ClientMessage{m, clientChan}
		}
	}
}

func (g *Game) runGame() {
	clientChans := make(map[chan MessageType]chan MessageType)
	clientMessageChan := make(chan ClientMessage)
	for {
		select {
		case client := <- g.addPlayerChan:
			log.Println("runGame: Adding client to game")
			endChan := make(chan MessageType)
			clientChans[client.gameChan] = endChan
			go receiveClientMessages(clientMessageChan, client.gameChan, endChan)
			//client.gameChan <- MessageType("ok")
		case cm := <- clientMessageChan:
			log.Println("Game got client message of type:", cm.typ)
			switch cm.typ {
			case "newTiles":
				g.newTiles()
				cm.clientChan <- MessageType("ok")
			case "exit":
				clientChans[cm.clientChan] <- MessageType("end")
				delete(clientChans, cm.clientChan)
				log.Println("runGame: Removing client from game")
			}
		}
	}
}

func main() {
	go addPlayerToAGame()
	go globalGame.runGame()
	const addr = "localhost:8080"
	fileserver := http.FileServer(http.Dir("public"))
	redirect := http.RedirectHandler("public/scrabble.html", http.StatusFound)

	http.Handle("/", redirect)
	http.Handle("/public/", http.StripPrefix("/public/", fileserver))
	http.HandleFunc("/connect", handleWebsocket)

	log.Println("Now listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
