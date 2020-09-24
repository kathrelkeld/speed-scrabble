package game

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn        WebSocketConn
	socketChan  chan SocketMsg
	tilesServed []Tile
	maxScore    int
	toGameChan  chan FromClientMsg
	validGame   bool
	running     bool

	//Accessible by other routines; Not allowed to change.
	Name             string
	TilesServedCount int
	ToClientChan     chan FromGameMsg
	AssignGameChan   chan *Game
}

func NewClient(conn WebSocketConn) *Client {
	c := Client{}
	c.conn = conn
	c.socketChan = make(chan SocketMsg)
	c.validGame = false
	c.running = false
	c.ToClientChan = make(chan FromGameMsg)
	c.AssignGameChan = make(chan *Game)
	return &c
}

func (c *Client) cleanup() {
	close(c.socketChan)
	close(c.ToClientChan)
	close(c.AssignGameChan)
	c.running = false
}

func (c *Client) newTiles(t []Tile) {
	c.TilesServedCount = len(t)
	c.tilesServed = t
	c.maxScore = 0
	for _, elt := range t {
		c.maxScore += elt.Points
	}
}

func (c *Client) addTile(t Tile) {
	c.tilesServed = append(c.tilesServed, t)
	c.TilesServedCount += 1
	c.maxScore += t.Points
}

func (c *Client) getTilesServedCount() int {
	return c.TilesServedCount
}

func (c *Client) getAllTilesServed() []Tile {
	return c.tilesServed[:c.TilesServedCount]
}

func (c *Client) getMaxScore() int {
	return c.maxScore
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

func (c *Client) ReadSocketMsgs() {
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
				NewGameChan <- GameRequest{MsgGlobal, c}
				game := <-c.AssignGameChan
				c.toGameChan = game.ToGameChan
				c.validGame = true
				c.sendSocketMsg(MsgOK, nil)
				return 0
			}
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

func (c *Client) Run() {
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
