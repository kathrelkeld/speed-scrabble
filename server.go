package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var globalGame = makeNewGame()
var globalClient = makeNewClient()
var newGameChan = make(chan GameRequest)

type MessageType int

const (
	MsgOK MessageType = iota
	MsgError
	MsgExit
	MsgConnect
	MsgJoinGame
	MsgGlobal
	MsgNewTiles
	MsgAddTile
	MsgVerify
	MsgScore
	MsgGameOver
)

var MessageTypeToString = map[MessageType]string{
	MsgOK:       "ok",
	MsgError:    "error",
	MsgExit:     "exit",
	MsgConnect:  "connect",
	MsgJoinGame: "joinGame",
	MsgGlobal:   "global",
	MsgNewTiles: "newTiles",
	MsgAddTile:  "addTile",
	MsgVerify:   "verify",
	MsgScore:    "score",
	MsgGameOver: "gameOver",
}

func (mt *MessageType) MarshalJSON() ([]byte, error) {
	s := MessageTypeToString[*mt]
	if s == "" {
		log.Println("No such message string to marshal!")
		panic("No such message string to marshal!")
	}
	return json.Marshal(s)
}

func (mt *MessageType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	for key, value := range MessageTypeToString {
		if value == s {
			*mt = key
			return nil
		}
	}
	log.Println("No such message string to unmarshal!")
	panic("No such message string to unmarshal!")
}

func (mt MessageType) String() string {
	return fmt.Sprintf("\"%v\"", MessageTypeToString[mt])
}

type GameRequest struct {
	Type  MessageType
	cInfo ClientInfo
}

type SocketMsg struct {
	Type MessageType
	At   time.Time
	Data *json.RawMessage
}

type FromClientMsg struct {
	typ   MessageType
	cInfo ClientInfo
}

type FromGameMsg struct {
	typ   MessageType
	gInfo GameInfo
}

type NewTileMsg struct {
	typ MessageType
	t   Tile
}

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
				_ = <-c.info.toClientChan
				//TODO: fix new tiles
				//tiles := c.getInitialTiles()
				//log.Println("Sending tiles:", tiles)
				//c.sendSocketMsg("tiles", tiles)
			case MsgAddTile:
				//tile := c.getNextTile()
				//if tile.Value == "" {
				// TODO: send out of tiles error
				//return
				//}
				//c.sendSocketMsg("tile", tile)
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
				g.newTiles()
				cm.cInfo.toClientChan <- FromGameMsg{MsgOK, g.info}
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
