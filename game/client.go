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
	conn        *websocket.Conn
	socketChan  chan msg.SocketData
	game        *Game
	tilesServed int
	state       gameState

	Name           string
	ToClientChan   chan MsgFromGame
	NewGameChan    chan MsgGameRequest
	AssignGameChan chan *Game
}

// NewClient creates a new Client with the given websocket connection and game assigner.
func NewClient(conn *websocket.Conn, ga *GameAssigner) *Client {
	return &Client{
		conn:           conn,
		socketChan:     make(chan msg.SocketData),
		ToClientChan:   make(chan MsgFromGame),
		AssignGameChan: make(chan *Game),
		NewGameChan:    ga.NewGameChan,
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

// handleSocketMsg is called to process an incomming message from the player.
// Returns 1 if the client needs to exit; else 0.
func (c *Client) handleSocketMsg(m *msg.SocketData) int {
	switch m.Type {
	case msg.Exit:
		// Player asking to close the connection.
		return 1
	case msg.JoinGame:
		// Player asking to join a new game.
		if c.game != nil {
			//TODO: handle already in game case
		} else {
			c.NewGameChan <- MsgGameRequest{c}
		}
	case msg.RoundReady:
		// Player indicating that they want to start a new round.
		if c.state == StateWaitingRoundReady {
			// If game is waiting, tell game player is ready.
			c.state = StateInit
			c.game.ToGameChan <- MsgFromClient{msg.Start, c, nil}
		} else {
			// If game is not waiting, tell game player wants a new round.
			c.state = StateWaitingRoundReady
			c.game.ToGameChan <- MsgFromClient{msg.RoundReady, c, nil}
		}
	case msg.Start:
		// Player confirming that they are ready for a new round.
		if c.state != StateWaitingRoundReady {
			// TODO handle already playing case
		} else {
			// Tell game that player is ready to start playing.
			c.game.ToGameChan <- MsgFromClient{msg.Start, c, nil}
		}
	case msg.AddTile:
		// Player asking for a new tile.
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
		// Player indicating they think they have won.
		if c.game == nil {
			c.sendSocketMsg(msg.Error, "Error: no active game!")
			return 1
		}
		score := ScoreMarshalledBoard(m.Data, c.game.tiles[:c.tilesServed])
		if !score.Win {
			// Send back invalid tiles.
			var invalid []Vec
			for k := range score.Invalid {
				invalid = append(invalid, k)
			}
			c.sendSocketMsg(msg.Invalid, invalid)
		} else {
			// Tell game that round is over.
			c.state = StateWaitingScores
			c.game.ToGameChan <- MsgFromClient{msg.Score, c, &score}
		}
	case msg.SendBoard:
		// Player sending board state after round is over.
		if c.game == nil {
			c.sendSocketMsg(msg.Error, "Error: no active game!")
			return 1
		}
		score := ScoreMarshalledBoard(m.Data, c.game.tiles[:c.tilesServed])
		c.game.ToGameChan <- MsgFromClient{msg.Score, c, &score}
	}
	return 0
}

// handleFromGameMsg is called to process an incomming message from the game.
// Returns 1 if the client needs to exit; else 0.
func (c *Client) handleFromGameMsg(m *MsgFromGame) int {
	switch m.Type {
	case msg.PlayerJoined:
		c.sendSocketMsg(msg.PlayerJoined, nil)
	case msg.RoundReady:
		// Game notifying client that a new round is ready.
		if c.state == StateWaitingRoundReady {
			// Reply ok since player is ready.
			c.state = StateInit
			c.game.ToGameChan <- MsgFromClient{msg.Start, c, nil}
		} else {
			// Ask player if they are ready.
			c.state = StateWaitingRoundReady
			c.sendSocketMsg(msg.RoundReady, nil)
		}
	case msg.Start:
		// Game telling client to start playing.
		c.tilesServed = c.game.startingTileCnt
		tiles := c.game.tiles[:c.tilesServed]
		c.state = StateRunning
		c.sendSocketMsg(msg.Start, tiles)
		log.Println("Sent tiles:", tiles)
	case msg.RoundOver:
		// Game telling client that someone won; asking for scores.
		if c.state != StateWaitingScores {
			// Have not sent score to game; request board from player.
			c.sendSocketMsg(msg.SendBoard, nil)
		}
		c.state = StateOver
	case msg.Score:
		// Game sending scores to client.
		c.sendSocketMsg(msg.Score, m.Data)
	}
	return 0
}

// Run is the main processing loop of a Client.
// Kicked off in its own goroutine by the server.
func (c *Client) Run() {
	go c.readSocketMsgs()
	defer c.cleanup()

	c.state = StateRunning
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
			c.sendSocketMsg(msg.OK, nil)
			c.game.AddPlayerChan <- c
		}
	}
}
