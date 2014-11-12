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

func unmarshalString(raw *json.RawMessage) string {
	var s string
	err := json.Unmarshal([]byte(*raw), &s)
	if err != nil {
		log.Println("error:", err)
		return ""
	}
	return s
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
	c.toGameChan <- FromClientMsg{MsgExit, c, nil}
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

func (c *Client) handleMsgStart() {
	c.toGameChan <- FromClientMsg{MsgStart, c, nil}
	_ = <-c.ToClientChan
	c.toGameChan <- FromClientMsg{MsgNewTiles, c, nil}
	tileMsg := <-c.ToClientChan
	//TODO: handle MsgError
	tiles := tileMsg.data.([]Tile)
	log.Println("Sending tiles:", tiles)
	c.sendSocketMsg(MsgStart, tiles)
	c.newTiles(tiles)
}

func (c *Client) handleSocketMsg(m *SocketMsg) int {
	if !c.validGame {
		switch m.Type { //For messages not involving a game.
		case MsgStart:
			if m.Type == MsgJoinGame {
				//TODO: handle already in game case
				c.Name = unmarshalString(m.Data)
				newGameChan <- GameRequest{MsgGlobal, c}
				game := <-c.AssignGameChan
				c.toGameChan = game.ToGameChan
				c.validGame = true
				c.sendSocketMsg(MsgOK, nil)
				return 0
			}
		}
	if !c.validGame {
		c.sendSocketMsg(MsgError, nil)
		log.Println("Cannot interact with an invalid game!")
		return 1
	}
	switch m.Type { //For game interaction messages.
	case MsgStart:
		if !c.validGame {
			c.sendSocketMsg(MsgError, nil)
			log.Println("Not a valid game to start!")
			return 1
		}
		c.toGameChan <- FromClientMsg{MsgStart, c, nil}
		//TODO: handle MsgError
		_ = <-c.ToClientChan
		c.handleMsgStart()
	case MsgAddTile:
		c.toGameChan <- FromClientMsg{MsgAddTile, c, nil}
		tileMsg := <-c.ToClientChan
		tile := tileMsg.data.(Tile)
		// TODO: send out of tiles error
		log.Println("Sending tile:", tile)
		c.sendSocketMsg(MsgAddTile, tile)
		c.addTile(tile)
	case MsgVerify:
		score := c.handleSendBoard(m.Data)
		//TODO: uncomment this when code is stable
		//if !score.Valid {
		//c.sendSocketMsg(MsgError, score)
		//} else {
		c.sendSocketMsg(MsgScore, score)
		c.toGameChan <- FromClientMsg{MsgGameOver, c, nil}
		_ = <-c.ToClientChan
		c.toGameChan <- FromClientMsg{MsgOK, c, nil}
		//}
	case MsgSendBoard:
		score := c.handleSendBoard(m.Data)
		c.sendSocketMsg(MsgScore, score)
	case MsgExit:
		c.toGameChan <- FromClientMsg{MsgExit, c, nil}
		return 1
	}
	return 0
}

func (c *Client) handleFromGameMsg(m *FromGameMsg) int {
	switch m.typ {
	case MsgNewGame:
		c.handleMsgStart()
	case MsgGameStatus:
		c.sendSocketMsg(MsgGameStatus, m.data)
	case MsgGameOver:
		c.sendSocketMsg(MsgSendBoard, nil)
		c.toGameChan <- FromClientMsg{MsgOK, c, nil}
	}
	return 0
}

func (c *Client) runClient() {
	defer c.onRunClientExit()
	c.running = true
	c.sendSocketMsg(MsgOK, nil)
	for {
		select {
		case m := <-c.socketChan:
			if c.handleSocketMsg(&m) != 0 {
				return
			}
		case gm := <-c.ToClientChan:
			log.Println("Game told client:", gm.typ)
			if c.handleFromGameMsg(&gm) != 0 {
				return
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
		globalGame.AddPlayerChan <- r.c
		r.c.AssignGameChan <- globalGame
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

func (g *Game) sendGameStatus() {
	var names []string
	for c := range g.clients {
		names = append(names, c.Name)
	}
	status := GameStatus{g.name, names}
	for c := range g.clients {
		c.ToClientChan <- FromGameMsg{MsgGameStatus, g, status}
	}
}

func (g *Game) allClientsTrue() bool {
	result := true
	for _, value := range g.clients {
		result = result && value
	}
	return result
}

func (g *Game) sendToAllClients(t MessageType) {
	log.Println("Sending to all clients.")
	for c := range g.clients {
		c.ToClientChan <- FromGameMsg{t, g, nil}
	}
}

func (g *Game) hearFromAllClients(t MessageType) {
	log.Println("Hearing from all clients.")
	for c := range g.clients {
		g.clients[c] = false
	}
	for !g.allClientsTrue() {
		cm := <-g.ToGameChan
		if cm.typ != t {
			cm.c.ToClientChan <- FromGameMsg{MsgError, g, nil}
		}
		g.clients[cm.c] = true
	}
}

func (g *Game) runGame() {
	for {
		select {
		case c := <-g.AddPlayerChan:
			log.Println("runGame: Adding client to game")
			g.clients[c] = false
			g.sendGameStatus()
		case cm := <-g.ToGameChan:
			log.Println("Game got client message of type:", cm.typ)
			switch cm.typ {
			case MsgStart:
				if g.isRunning { //Game is already running!
					cm.c.ToClientChan <- FromGameMsg{MsgError, g, nil}
					continue
				}
				g.sendToAllClients(MsgNewGame)
				g.hearFromAllClients(MsgStart)
				g.newTiles()
				g.sendToAllClients(MsgStart)
				g.isRunning = true
			case MsgNewTiles:
				if !g.isRunning {
					cm.c.ToClientChan <- FromGameMsg{MsgError, g, nil}
					continue
				}
				//TODO: handle confirm from all games
				m := FromGameMsg{MsgOK, g, g.tiles[:12]}
				cm.c.ToClientChan <- m
			case MsgAddTile:
				if !g.isRunning {
					cm.c.ToClientChan <- FromGameMsg{MsgError, g, nil}
					continue
				}
				if cm.c.TilesServedCount >= len(g.tiles) {
					cm.c.ToClientChan <- FromGameMsg{MsgError, g, nil}
					continue
				}
				m := FromGameMsg{MsgOK, g,
					g.tiles[cm.c.TilesServedCount]}
				cm.c.ToClientChan <- m
			case MsgGameOver:
				//TODO: get scores to determine a winner
				g.sendToAllClients(MsgGameOver)
				g.hearFromAllClients(MsgOK)
				g.isRunning = false
			case MsgExit:
				delete(g.clients, cm.c)
				log.Println("runGame: Removing client from game")
				if len(g.clients) == 0 {
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
