package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var globalGame = makeNewGame()
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
	m, err := newSocketMsg(t, d)
	if err != nil {
		log.Println("error:", err)
		return err
	}
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

func (c *Client) readSocketMsgs() {
	for {
		m, err := c.readSocketMsg()
		if err != nil {
			c.socketChan <- SocketMsg{Type: MsgExit}
			return
		}
		c.socketChan <- m
	}
}

func (c *Client) onRunClientExit() {
	c.toGameChan <- FromClientMsg{MsgExit, c.info}
	c.cleanup()
}

func (c *Client) runClient() {
	defer c.onRunClientExit()
	c.running = true
	c.sendSocketMsg(MsgOK, nil)
	for {
		select {
		case m := <-c.socketChan:
			switch m.Type {
			case MsgJoinGame:
				newGameChan <- GameRequest{MsgGlobal, c.info}
				gameInfo := <-c.info.assignGameChan
				c.toGameChan = gameInfo.toGameChan
				c.validGame = true
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
				s := board.scoreBoard(c)
				c.sendSocketMsg(MsgScore, s)
			case MsgExit:
				c.toGameChan <- FromClientMsg{MsgExit, c.info}
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
	c := makeNewClient()
	c.conn = conn
	m, err := c.readSocketMsg()
	if err != nil {
		return
	}
	if m.Type != MsgConnect {
		log.Println("Incorrect type for new connection!")
		return
	}
	go c.runClient()
	go c.readSocketMsgs()
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

func (g *Game) allClientsTrue() bool {
	result := true
	for _, value := range g.clientChans {
		result = result && value
	}
	return result
}

func (g *Game) sendToAllClients(t MessageType) {
	for key := range g.clientChans {
		key <- FromGameMsg{t, g.info}
	}
}

func (g *Game) hearFromAllClients() {
	for key := range g.clientChans {
		g.clientChans[key] = false
	}
	for !g.allClientsTrue() {
		cm := <-g.info.toGameChan
		g.clientChans[cm.cInfo.toClientChan] = true
	}
}

func (g *Game) runGame() {
	for {
		select {
		case clientInfo := <-g.info.addPlayerChan:
			log.Println("runGame: Adding client to game")
			g.clientChans[clientInfo.toClientChan] = false
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
				delete(g.clientChans, cm.cInfo.toClientChan)
				log.Println("runGame: Removing client from game")
			}
		}
	}
}

func cleanup() {
	close(newGameChan)
	globalGame.cleanup()
}

func main() {
	defer cleanup()
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
