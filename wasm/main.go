package main

import (
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

func (mgr *GameManager) websocketSendEmpty(t msg.Type) {
	mgr.websocketSend([]byte{byte(t)})
}

func (mgr *GameManager) websocketSend(b []byte) {
	v := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(v, b)

	fmt.Println("Sending message")
	if !mgr.socket.Truthy() {
		// TODO handle error
	}
	mgr.socket.Call("send", v)
}

func (mgr *GameManager) websocketGet() js.Func {
	// args = onmessage event
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("Getting message through websocket")
		if len(args) == 0 {
			fmt.Println("empty message received!")
			return nil
		}
		m := []byte(args[0].Get("data").String())
		fmt.Println(m)
		t := msg.Type(m[0])
		m = m[1:]

		// TODO handle error
		_ = mgr.handleSocketMsg(t, m)
		return nil
	})
}

func (mgr *GameManager) newSocketWrapper() js.Func {
	onOpen := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("websocket open")
		return nil
	})
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		wsPrefix := "ws://"
		loc := js.Global().Get("window").Get("location")
		protocol := loc.Get("protocol")
		if protocol.String() == "https:" {
			wsPrefix = "wss://"
		}
		host := loc.Get("host").String()

		ws := js.Global().Get("WebSocket").New(wsPrefix + host + "/connect")
		mgr.socket = ws
		ws.Call("addEventListener", "message", mgr.websocketGet())
		ws.Call("addEventListener", "open", onOpen)
		return nil
	})
}

func main() {
	mgr := NewGameManager(Vec{16, 16}, 16)
	mgr.setUpPage()
	mgr.listens = NewListeners(mgr)
	mgr.newSocketWrapper().Invoke()
	<-make(chan bool)
}
