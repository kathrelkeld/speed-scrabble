package main

import (
	"encoding/json"
	//"log"
	//"reflect"
	"testing"
)

type FakeWebSocketConn struct {
	socketMsgChan chan SocketMsg
	lastSent      SocketMsg
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
	var m SocketMsg
	err := json.Unmarshal(b, &m)
	if err != nil {
		conn.t.Errorf("Error while unmarshalling %v", string(b))
	}
	return err
}

// Call this instead of handling an incoming websocket request
func runTestClient(t *testing.T) (*Client, FakeWebSocketConn) {
	c := makeNewClient()
	conn := FakeWebSocketConn{make(chan SocketMsg), SocketMsg{}, t}
	c.conn = conn
	go c.readSocketMsgs()
	go c.runClient()
	return c, conn
}

func TestExitMessage(t *testing.T) {
	_, conn := runTestClient(t)
	defer close(conn.socketMsgChan)
	m, _ := newSocketMsg(MsgExit, nil)
	conn.socketMsgChan <- m
}

func TestSynchonousStart(t *testing.T) {
	_, connA := runTestClient(t)
	defer close(connA.socketMsgChan)
	_, connB := runTestClient(t)
	defer close(connB.socketMsgChan)
	m, _ := newSocketMsg(MsgJoinGame, nil)
	connA.socketMsgChan <- m
	connB.socketMsgChan <- m
	m, _ = newSocketMsg(MsgStart, nil)
}
