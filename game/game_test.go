package game

import (
	"encoding/json"
	"testing"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

type FakeWebSocketConn struct {
	socketMsgChan chan msg.Socket
	lastSent      msg.Socket
	t             *testing.T
}

func (conn FakeWebSocketConn) ReadMessage() (int, []byte, error) {
	m := <-conn.socketMsgChan
	j, err := json.Marshal(m)
	if err != nil {
		conn.t.Errorf("Error while marshalling %v", m)
	}
	return 1, j, err
}

func (conn FakeWebSocketConn) WriteMessage(i int, b []byte) error {
	var m msg.Socket
	err := json.Unmarshal(b, &m)
	if err != nil {
		conn.t.Errorf("Error while unmarshalling %v", string(b))
	}
	return err
}

// Call this instead of handling an incoming websocket request
func runTestClient(t *testing.T) (*Client, FakeWebSocketConn) {
	c := makeNewClient()
	conn := FakeWebSocketConn{make(chan msg.Socket), msg.Socket{}, t}
	c.conn = conn
	go c.readSocketMsgs()
	go c.runClient()
	return c, conn
}

func TestExitMessage(t *testing.T) {
	_, conn := runTestClient(t)
	defer close(conn.socketMsgChan)
	m, _ := msg.NewSocket(msg.Exit, nil)
	conn.socketMsgChan <- m
}

func TestSynchonousStart(t *testing.T) {
	_, connA := runTestClient(t)
	defer close(connA.socketMsgChan)
	_, connB := runTestClient(t)
	defer close(connB.socketMsgChan)
	m, _ := msg.NewSocket(msg.JoinGame, nil)
	connA.socketMsgChan <- m
	connB.socketMsgChan <- m
	m, _ = msg.NewSocket(msg.Start, nil)
}
