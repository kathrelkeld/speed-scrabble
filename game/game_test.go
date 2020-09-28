package game

import (
	"testing"
	"time"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

type FakeWebsocketConn struct {
	chRead   chan []byte
	chWrite  chan []byte
	lastSent []byte
	lastRead []byte
	t        *testing.T
}

func (conn *FakeWebsocketConn) ReadMessage() (int, []byte, error) {
	m := <-conn.chWrite
	conn.lastSent = m
	return 1, m, nil
}

func (conn *FakeWebsocketConn) WriteMessage(i int, b []byte) error {
	conn.lastRead = b
	conn.chRead <- b
	return nil
}

func (conn *FakeWebsocketConn) Close() error { return nil }

func (conn *FakeWebsocketConn) sendMsg(t msg.Type, d interface{}) {
	m, _ := msg.NewSocketData(t, d)
	conn.chWrite <- m
}

func (conn *FakeWebsocketConn) waitForMsg(t msg.Type) []byte {
	select {
	case m := <-conn.chRead:
		if got := msg.Type(m[0]); got != t {
			conn.t.Errorf("Got message of %v, expected %v", got, t)
		}
		return m[1:]
	case <-time.After(1 * time.Second):
		conn.t.Errorf("Timeout waiting for message of type %v", t)
	}
	return []byte{}
}

func NewFakeWebsocketConn(t *testing.T) *FakeWebsocketConn {
	return &FakeWebsocketConn{
		chRead:  make(chan []byte),
		chWrite: make(chan []byte),
		t:       t,
	}
}

func TestExitMessage(t *testing.T) {
	ga := NewGameAssigner()
	go ga.Run()
	if len(ga.games) != 0 {
		t.Errorf("Game count pre-client: Got %v; Expected 0", len(ga.games))
	}

	conn := NewFakeWebsocketConn(t)
	ga.StartNewClient(conn)
	conn.waitForMsg(msg.PlayerJoined)

	if len(ga.games) != 1 {
		t.Errorf("Game count adding a single client: Got %v; Expected 1", len(ga.games))
	}

	// TODO actual check
}

func TestSynchonousStart(t *testing.T) {
	ga := NewGameAssigner()
	connA := NewFakeWebsocketConn(t)
	ga.StartNewClient(connA)
	connB := NewFakeWebsocketConn(t)
	ga.StartNewClient(connB)

	// TODO actual check
	connA.sendMsg(msg.Start, nil)
	connB.sendMsg(msg.Start, nil)
}
