package game

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

// A Client represents a single player.  The client handles websocket interactions
// and keeping state for this player.
type Client struct {
	Name      string
	conn      WebsocketConn
	ga        *GameAssigner
	game      *Game
	servedCnt int
}

// Close is used to request the Client exit gracefully.
func (c *Client) Close() {
	if c.game != nil {
		c.game.toGameChan <- MsgFromClient{msg.Exit, c, nil}
	}
	c.conn.Close()
}

// readSocketMsgs passes all incoming websocket messages to a channel to be handled.
// Must be run as a separate go routine.
// Will return when c.conn is closed through c.Close() or on websocket error.
// The first message byte is always the type, followed by JSON data.
func (c *Client) readSocketMsgs() {
	for {
		_, b, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Error reading socket message; closing client", err)
			if c.game != nil {
				c.Close()
			}
			return
		}
		t := msg.Type(b[0]) // The first byte is the message type.
		b = b[1:]           // Any remaining bytes are JSON encoded data.

		log.Println("Got websocket message of type", t)
		if t == msg.JoinGame {
			c.ga.NewGameChan <- MsgGameRequest{c}
		} else if c.game != nil {
			c.game.toGameChan <- MsgFromClient{t, c, b}
		} else {
			log.Println("Ignoring websocket message of type", t)
		}
	}
}

// sendSocketMsg sends a websocket message of the given type with the given data.
// The data must be marshallable into JSON.
func (c *Client) sendSocketMsg(t msg.Type, d interface{}) {
	b, err := msg.NewSocketData(t, d)
	if err != nil {
		log.Println("Could not sent websocket message of type", t, err)
		return
	}

	err = c.conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Println("Could not write to websocket; closing client", err)
		if c.game != nil {
			c.Close()
		}
		return
	}
	log.Println("Sent websocket message of type", t)
	return
}

// addTile is called when a player requests a new tile.  It either sends a tile or an error.
func (c *Client) addTile() {
	if c.servedCnt >= len(c.game.tiles) {
		c.sendSocketMsg(msg.OutOfTiles, "Out of tiles!")
	} else {
		tile := c.game.tiles[c.servedCnt]
		log.Println("Sending tile:", tile)
		c.sendSocketMsg(msg.AddTile, tile)
		c.servedCnt += 1
	}
}

// ScoreMarshalledBoard takes a JSON board and returns a score for that board.
func (c *Client) ScoreMarshalledBoard(d []byte) *Score {
	var board Board
	err := json.Unmarshal(d, &board)
	if err != nil {
		log.Println("error:", err)
		return nil
	}
	return board.scoreBoard(c.game.tiles[:c.servedCnt])
}

func (c *Client) SendScore(s *Score) {
	// TODO: remove some of this data?
	c.sendSocketMsg(s.msg, s)
}
