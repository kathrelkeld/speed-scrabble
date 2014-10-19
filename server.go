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

func (c *Client) handleSendBoard(raw *json.RawMessage) Score {
	var board Board
	err := json.Unmarshal([]byte(*raw), &board)
	if err != nil {
		log.Println("error:", err)
		return Score{}
	}
	return board.scoreBoard(c)
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
				c.toGameChan <- FromClientMsg{MsgStart, c.info}
				//_ = <- c.info.toClientChan
				//c.toGameChan <- FromClientMsg{MsgStart, c.info}
				//tileMsg := <-c.info.newTilesChan
				//TODO: check for error
				//tiles := tileMsg.tiles
				//log.Println("Sending tiles:", tiles)
				//c.sendSocketMsg(MsgNewTiles, tiles)
				//c.newTiles(tiles)
			case MsgAddTile:
				c.toGameChan <- FromClientMsg{MsgAddTile, c.info}
				tileMsg := <-c.info.newTilesChan
				tile := tileMsg.tiles[0]
				// TODO: send out of tiles error
				log.Println("Sending tile:", tile)
				c.sendSocketMsg(MsgAddTile, tile)
				c.addTile(tile)
			case MsgVerify:
				score := c.handleSendBoard(m.Data)
				//if !score.Valid {
				//c.sendSocketMsg(MsgError, score)
				//} else {
				c.sendSocketMsg(MsgScore, score)
				c.toGameChan <- FromClientMsg{MsgGameOver, c.info}
				//}
			case MsgSendBoard:
				score := c.handleSendBoard(m.Data)
				c.sendSocketMsg(MsgScore, score)
			case MsgExit:
				c.toGameChan <- FromClientMsg{MsgExit, c.info}
				return
			}
		case gm := <-c.info.toClientChan:
			log.Println("Game told client:", gm.typ)
			switch gm.typ {
			case MsgNewGame:
				//TODO: tell playter that game will start
				c.toGameChan <- FromClientMsg{MsgStart, c.info}
			case MsgStart:
				c.toGameChan <- FromClientMsg{MsgNewTiles, c.info}
				tileMsg := <-c.info.newTilesChan
				//TODO: check for error
				tiles := tileMsg.tiles
				log.Println("Sending tiles:", tiles)
				c.sendSocketMsg(MsgNewTiles, tiles)
				c.newTiles(tiles)
			case MsgGameOver:
				c.sendSocketMsg(MsgSendBoard, nil)
				c.toGameChan <- FromClientMsg{MsgOK, c.info}
			}
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

func (g *Game) sendToAllClients(t MessageType, except chan FromGameMsg) {
	for key := range g.clientChans {
		if key != except {
			key <- FromGameMsg{t, g.info}
		}
	}
}

func (g *Game) hearFromAllClients(t MessageType, except chan FromGameMsg) {
	for key := range g.clientChans {
		g.clientChans[key] = false
	}
	g.clientChans[except] = true
	for !g.allClientsTrue() {
		cm := <-g.info.toGameChan
		if cm.typ != t {
			cm.cInfo.toClientChan <- FromGameMsg{MsgError, g.info}
		}
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
			case MsgStart:
				if g.isRunning { //Game is already running!
					cm.cInfo.toClientChan <- FromGameMsg{MsgError, g.info}
					continue
				}
				g.sendToAllClients(MsgNewGame, cm.cInfo.toClientChan)
				g.hearFromAllClients(MsgStart, cm.cInfo.toClientChan)
				g.newTiles()
				g.sendToAllClients(MsgStart, cm.cInfo.toClientChan)
				cm.cInfo.toClientChan <- FromGameMsg{MsgStart, g.info}
				g.isRunning = true
			case MsgNewTiles:
				if !g.isRunning {
					cm.cInfo.toClientChan <- FromGameMsg{MsgError, g.info}
					continue
				}
				//TODO: handle confirm from all games
				m := NewTileMsg{MsgOK, g.tiles[:12], g.info}
				cm.cInfo.newTilesChan <- m
			case MsgAddTile:
				if !g.isRunning {
					cm.cInfo.toClientChan <- FromGameMsg{MsgError, g.info}
					continue
				}
				if cm.cInfo.tilesServedCount >= len(g.tiles) {
					cm.cInfo.toClientChan <- FromGameMsg{MsgError, g.info}
					continue
				}
				m := NewTileMsg{MsgOK, []Tile{g.tiles[cm.cInfo.tilesServedCount]},
					g.info}
				cm.cInfo.newTilesChan <- m
			case MsgGameOver:
				//TODO: get scores to determine a winner
				g.sendToAllClients(MsgGameOver, cm.cInfo.toClientChan)
				g.hearFromAllClients(MsgOK, cm.cInfo.toClientChan)
				g.isRunning = false
			case MsgExit:
				delete(g.clientChans, cm.cInfo.toClientChan)
				log.Println("runGame: Removing client from game")
				if len(g.clientChans) == 0 {
					g.isRunning = false
				}
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
	const addr = ":8080"
	fileserver := http.FileServer(http.Dir("public"))
	redirect := http.RedirectHandler("public/scrabble.html", http.StatusFound)

	http.Handle("/", redirect)
	http.Handle("/public/", http.StripPrefix("/public/", fileserver))
	http.HandleFunc("/connect", handleWebsocket)

	log.Println("Now listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
