package game

import (
	"log"

	"github.com/gorilla/websocket"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

// A Client represents a single player.  The client interacts with the player's page
// (via websocket), with a single game (via channels), and with the overall game state
// (via channels, e.g. starting new games).
type Client struct {
	conn        WebSocketConn
	socketChan  chan msg.SocketData
	game        *Game
	tilesServed int
	state       gameState
	lastScore   *Score

	Name           string
	ToClientChan   chan MsgFromGame
	AssignGameChan chan *Game
}

// NewClient creates a new Client with the given websocket connection.  No game is assigned.
func NewClient(conn WebSocketConn) *Client {
	return &Client{
		conn:           conn,
		socketChan:     make(chan msg.SocketData),
		ToClientChan:   make(chan MsgFromGame),
		AssignGameChan: make(chan *Game),
	}
}

func (c *Client) cleanup() {
	close(c.socketChan)
	close(c.ToClientChan)
	close(c.AssignGameChan)
	c.state = StateOver
	if c.game != nil {
		c.game.ToGameChan <- MsgFromClient{msg.Exit, c, nil}
	}
}

// readSocketMsgs passes all incoming websocket messages to a channel to be handled.
// Must be run as a separate go routine; started by Run().
// The first message byte is always the type, followed by JSON data.
func (c *Client) readSocketMsgs() {
	for {
		_, b, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Error reading socket message; closing client", err)
			c.socketChan <- msg.SocketData{msg.Exit, nil}
			return
		}
		t := msg.Type(b[0]) // The first byte is the message type.
		b = b[1:]           // Any remaining bytes are JSON encoded data.

		log.Println("Got websocket message of type", t)
		c.socketChan <- msg.SocketData{t, b}
	}
}

// sendSocketMsg sends a websocket message of the given type with the given data.
// The data must be marshallable into JSON.
func (c *Client) sendSocketMsg(t msg.Type, d interface{}) error {
	b, err := msg.NewSocketData(t, d)
	if err != nil {
		log.Println("Could not sent websocket message of type", t, err)
		return err
	}

	err = c.conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Println("Could not write to websocket; closing client", err)
		c.socketChan <- msg.SocketData{msg.Exit, nil}
		return err
	}
	log.Println("Sent websocket message of type", t)
	return nil
}

func (c *Client) handleSocketMsg(m *msg.SocketData) int {
	switch m.Type {
	case msg.Exit:
		return 1
	case msg.JoinGame:
		// Player asking to join a new game.
		if c.game != nil {
			//TODO: handle already in game case
		} else {
			NewGameChan <- MsgGameRequest{c}
		}
	case msg.GameReady:
		// Player indicating that they want to start a new round.
		if c.state == StateWaitingGameReady {
			// Tell game player is ready.
			c.state = StateInit
			c.game.ToGameChan <- MsgFromClient{msg.Start, c, nil}
		} else {
			// Tell game player wants a new game.
			c.state = StateWaitingGameReady
			c.game.ToGameChan <- MsgFromClient{msg.GameReady, c, nil}
		}
	case msg.Start:
		// Player confirming that they are ready for a new round.
		if c.state != StateWaitingGameReady {
			// TODO handle already playing case
		} else {
			// Tell game that player is ready to start playing.
			c.game.ToGameChan <- MsgFromClient{msg.Start, c, nil}
		}
	case msg.AddTile:
		if c.game == nil {
			c.sendSocketMsg(msg.Error, "Error: no active game!")
			return 1
		}
		if c.tilesServed >= len(c.game.tiles) {
			c.sendSocketMsg(msg.OutOfTiles, "Out of tiles!")
		} else {
			tile := c.game.tiles[c.tilesServed]
			log.Println("Sending tile:", tile)
			c.sendSocketMsg(msg.AddTile, tile)
			c.tilesServed += 1
		}
	case msg.Verify:
		if c.game == nil {
			c.sendSocketMsg(msg.Error, "Error: no active game!")
			return 1
		}
		score := ScoreMarshalledBoard(m.Data, c.game.tiles[:c.tilesServed])
		if !score.Win {
			var invalid []Vec
			for k := range score.Invalid {
				invalid = append(invalid, k)
			}
			c.sendSocketMsg(msg.Invalid, invalid)
		} else {
			c.state = StateWaitingScores
			c.lastScore = &score
			c.game.ToGameChan <- MsgFromClient{msg.GameOver, c, nil}
		}
	case msg.SendBoard:
		if c.game == nil {
			c.sendSocketMsg(msg.Error, "Error: no active game!")
			return 1
		}
		score := ScoreMarshalledBoard(m.Data, c.game.tiles[:c.tilesServed])
		c.game.ToGameChan <- MsgFromClient{msg.Score, c, &score}
	}
	return 0
}

func (c *Client) handleFromGameMsg(m *MsgFromGame) int {
	switch m.Type {
	case msg.GameReady:
		// Game notifying client that a new game is ready.
		if c.state == StateWaitingGameReady {
			// Reply ok since player is ready.
			c.state = StateInit
			c.game.ToGameChan <- MsgFromClient{msg.Start, c, nil}
		} else {
			// Ask player if they are ready.
			c.state = StateWaitingGameReady
			c.sendSocketMsg(msg.GameReady, nil)
		}
	case msg.Start:
		// Game telling client to start playing.
		c.tilesServed = c.game.startingTileCnt
		tiles := c.game.tiles[:c.tilesServed]
		log.Println("Sending tiles:", tiles)
		c.state = StateRunning
		c.sendSocketMsg(msg.Start, tiles)
	case msg.GameOver:
		// Game telling client that someone won; asking for scores.
		if c.state == StateWaitingScores {
			// Client already has score.  Send it.
			c.game.ToGameChan <- MsgFromClient{msg.Score, c, c.lastScore}
		} else {
			c.sendSocketMsg(msg.SendBoard, nil)
		}
		c.state = StateOver
	case msg.Score:
		// Game sending scores to client.
		c.sendSocketMsg(msg.Score, m.Data)
	}
	return 0
}

func (c *Client) Run() {
	go c.readSocketMsgs() // Pass websocket messages to socketChan.
	defer c.cleanup()

	c.state = StateRunning
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
		case game := <-c.AssignGameChan:
			c.game = game
			c.game.ToGameChan = game.ToGameChan
			c.sendSocketMsg(msg.OK, nil)
		}
	}
}
