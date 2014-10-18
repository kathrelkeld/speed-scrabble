package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var globalGame = makeNewGame()
var globalClient = makeNewClient()
var newGameChan = make(chan GameRequest)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (c *Client) readSocketMsg() (SocketMsg, error) {
	var m SocketMsg
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

func (c *Client) sendSocketMsg(t MessageType, d interface{}) error {
	marshaledData, err := json.Marshal(d)
	if err != nil {
		log.Println("error:", err)
		return err
	}
	raw := json.RawMessage(marshaledData)
	m := SocketMsg{Type: t, At: time.Now(), Data: &raw}
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

func (c *Client) readSocketMsgs(ch chan SocketMsg) {
	for {
		m, err := c.readSocketMsg()
		if err != nil {
			ch <- SocketMsg{Type: MsgExit}
			return
		}
		ch <- m
	}
}

func (c *Client) runClient(socketChan chan SocketMsg) {
	defer func() { c.toGameChan <- FromClientMsg{MsgExit, c.info} }()
	c.sendSocketMsg(MsgOK, nil)
	for {
		select {
		case m := <-socketChan:
			switch m.Type {
			case MsgJoinGame:
				newGameChan <- GameRequest{MsgGlobal, c.info}
				gameInfo := <-c.info.assignGameChan
				c.toGameChan = gameInfo.toGameChan
				c.sendSocketMsg(MsgOK, nil)
			case MsgNewTiles:
				c.toGameChan <- FromClientMsg{MsgNewTiles, c.info}
				//TODO: check for error
				tileMsg := <-c.info.newTilesChan
				tiles := tileMsg.tiles
				log.Println("Sending tiles:", tiles)
				c.sendSocketMsg(MsgNewTiles, tiles)
				c.newTiles(tiles)
			case MsgAddTile:
				c.toGameChan <- FromClientMsg{MsgAddTile, c.info}
				tileMsg := <-c.info.newTilesChan
				tile := tileMsg.tiles[0]
				// TODO: send out of tiles error
				log.Println("Sending tile:", tile)
				c.sendSocketMsg(MsgAddTile, tile)
				c.addTile(tile)
			case MsgVerify:
				var board Board
				err := json.Unmarshal([]byte(*m.Data), &board)
				if err != nil {
					log.Println("error:", err)
					return
				}
				s := board.scoreBoard(globalClient)
				c.sendSocketMsg(MsgScore, s)
			case MsgExit:
				return
			}
		case gm := <-c.info.toClientChan:
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
	m, err := c.readSocketMsg()
	if err != nil {
		return
	}
	if m.Type != MsgConnect {
		log.Println("Incorrect type for new connection!")
		return
	}
	socketChan := make(chan SocketMsg)
	go c.runClient(socketChan)
	go c.readSocketMsgs(socketChan)
}

func addAPlayerToAGame() {
	for {
		r := <-newGameChan
		log.Println("newGameChan: Adding client to game")
		globalGame.info.addPlayerChan <- r.cInfo
		r.cInfo.assignGameChan <- globalGame.info
	}
}

type ClientMessage struct {
	typ        MessageType
	clientChan chan MessageType
}

func receiveClientMessages(gameChan chan ClientMessage,
	clientChan, endChan chan MessageType) {
	for {
		select {
		case _ = <-endChan:
			log.Println("ClientMessage: Got exit signal! Ending loop!")
			break
		case m := <-clientChan:
			log.Println("Got a client message: ", m)
			gameChan <- ClientMessage{m, clientChan}
		}
	}
}

func (g *Game) runGame() {
	clientChans := make(map[chan FromGameMsg]bool)
	for {
		select {
		case clientInfo := <-g.info.addPlayerChan:
			log.Println("runGame: Adding client to game")
			clientChans[clientInfo.toClientChan] = false
			//client.gameChan <- MessageType("ok")
		case cm := <-g.info.toGameChan:
			log.Println("Game got client message of type:", cm.typ)
			switch cm.typ {
			case MsgNewTiles:
				//TODO: handle confirm from all games
				g.newTiles()
				m := NewTileMsg{MsgOK, g.tiles[:12], g.info}
				cm.cInfo.newTilesChan <- m
			case MsgAddTile:
				m := NewTileMsg{MsgOK, []Tile{g.tiles[cm.cInfo.tilesServedCount]},
					g.info}
				cm.cInfo.newTilesChan <- m
			case MsgExit:
				delete(clientChans, cm.cInfo.toClientChan)
				log.Println("runGame: Removing client from game")
			}
		}
	}
}

func main() {
	go addAPlayerToAGame()
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
