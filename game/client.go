package game

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

type Client struct {
	conn        WebSocketConn
	socketChan  chan msg.SocketData
	tilesServed []Tile
	maxScore    int
	toGameChan  chan MsgFromClient
	validGame   bool
	running     bool

	//Accessible by other routines; Not allowed to change.
	Name             string
	TilesServedCount int
	ToClientChan     chan MsgFromGame
	AssignGameChan   chan *Game
}

func NewClient(conn WebSocketConn) *Client {
	c := Client{}
	c.conn = conn
	c.socketChan = make(chan msg.SocketData)
	c.validGame = false
	c.running = false
	c.ToClientChan = make(chan MsgFromGame)
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
	typ        msg.Type
	clientChan chan msg.Type
}

func receiveClientMessages(gameChan chan ClientMessage,
	clientChan, endChan chan msg.Type) {
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

func unmarshalString(b []byte) string {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		log.Println("error:", err)
		return ""
	}
	return s
}

func (c *Client) ReadSocketMsg() (msg.SocketData, error) {
	// TODO check error
	_, b, err := c.conn.ReadMessage()
	if err != nil {
		log.Println("Bad message", err)
		return msg.SocketData{}, err
	}
	t := msg.Type(b[0])
	b = b[1:]

	return msg.SocketData{t, b}, nil
}

func (c *Client) ReadSocketMsgs() {
	for {
		m, err := c.ReadSocketMsg()
		if err != nil {
			log.Println("Error reading socket message", err)
			return
		}
		log.Println("Got websocket message of type", m.Type, m.Data)
		c.socketChan <- m
	}
}

func (c *Client) sendSocketMsg(t msg.Type, d interface{}) error {
	b, err := msg.NewSocketData(t, d)
	if err != nil {
		// TODO handle error
		return err
	}

	err = c.conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		return err
	}
	log.Println("Sent websocket message of type", t)
	return nil
}

func (c *Client) onRunClientExit() {
	c.toGameChan <- MsgFromClient{msg.Exit, c, nil}
	c.cleanup()
}

func (c *Client) handleSendBoard(b []byte) Score {
	var board Board
	err := json.Unmarshal(b, &board)
	if err != nil {
		log.Println("error:", err)
		return Score{}
	}
	return board.scoreBoard(c)
}

func (c *Client) handleMsgStart() {
	log.Println("Starting game")
	c.toGameChan <- MsgFromClient{msg.Start, c, nil}
	_ = <-c.ToClientChan
	c.toGameChan <- MsgFromClient{msg.NewTiles, c, nil}
	tileMsg := <-c.ToClientChan
	//TODO: handle MsgError
	tiles := tileMsg.data.([]Tile)
	log.Println("Sending tiles:", tiles)
	c.sendSocketMsg(msg.Start, tiles)
	c.newTiles(tiles)
}

func (c *Client) handleSocketMsg(m *msg.SocketData) int {
	if !c.validGame {
		switch m.Type { //For messages not involving a game.
		case msg.JoinGame:
			//TODO: handle already in game case
			c.Name = unmarshalString(m.Data)
			NewGameChan <- MsgGameRequest{msg.OK, c}
			game := <-c.AssignGameChan
			c.validGame = true
			c.toGameChan = game.ToGameChan
			c.sendSocketMsg(msg.OK, nil)
			return 0
		}

		c.sendSocketMsg(msg.Error, nil)
		log.Println("Cannot interact with an invalid game!")
		return 1
	}
	switch m.Type { //For game interaction messages.
	case msg.Start:
		c.toGameChan <- MsgFromClient{msg.Start, c, nil}
		//TODO: handle msg.Error
		_ = <-c.ToClientChan
		c.handleMsgStart()
	case msg.AddTile:
		c.toGameChan <- MsgFromClient{msg.AddTile, c, nil}
		tileMsg := <-c.ToClientChan
		if tileMsg.Type == msg.Error {
			// TODO: send out of tiles error
			log.Println(tileMsg.data)
		} else {
			tile := tileMsg.data.(Tile)
			log.Println("Sending tile:", tile)
			c.sendSocketMsg(msg.AddTile, tile)
			c.addTile(tile)
		}
	case msg.Verify:
		score := c.handleSendBoard(m.Data)
		if !score.Valid {
			var invalid []Vec
			for k := range(score.Invalid) {
				invalid = append(invalid, k)
			}
			c.sendSocketMsg(msg.Invalid, invalid)
		} else {
			c.sendSocketMsg(msg.Score, score)
			c.toGameChan <- MsgFromClient{msg.GameOver, c, nil}
			_ = <-c.ToClientChan
			c.toGameChan <- MsgFromClient{msg.OK, c, nil}
		}
	case msg.SendBoard:
		score := c.handleSendBoard(m.Data)
		c.sendSocketMsg(msg.Score, score)
	case msg.Exit:
		c.toGameChan <- MsgFromClient{msg.Exit, c, nil}
		return 1
	}
	return 0
}

func (c *Client) handleFromGameMsg(m *MsgFromGame) int {
	switch m.Type {
	case msg.NewGame:
		c.handleMsgStart()
		c.validGame = true
	case msg.GameStatus:
		c.sendSocketMsg(msg.GameStatus, m.data)
	case msg.GameOver:
		c.sendSocketMsg(msg.SendBoard, nil)
		c.toGameChan <- MsgFromClient{msg.OK, c, nil}
		c.validGame = false
	}
	return 0
}

func (c *Client) Run() {
	defer c.onRunClientExit()
	c.running = true
	c.sendSocketMsg(msg.OK, nil)
	for {
		select {
		case m := <-c.socketChan:
			if c.handleSocketMsg(&m) != 0 {
				log.Println("Exiting Client")
				return
			}
		case gm := <-c.ToClientChan:
			if c.handleFromGameMsg(&gm) != 0 {
				log.Println("Exiting Client")
				return
			}
		}
	}
}
