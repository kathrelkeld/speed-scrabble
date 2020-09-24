package main

import (
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/game"
)

var onOk []js.Func

func websocketOnOk(f js.Func) {
	onOk = append(onOk, f)
}

func websocketSendEmpty(t game.MessageType) {
	websocketSend([]byte{byte(t)})
}

func websocketSend(b []byte) {
	v := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(v, b)

	fmt.Println("Sending message")
	soc := js.Global().Get("socket")
	if !soc.Truthy() {
		// TODO handle error
	}
	soc.Call("send", v)
}

func websocketGet() js.Func {
	// args = onmessage event
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("Getting message through websocket")
		if len(args) == 0 {
			fmt.Println("empty message received!")
			return nil
		}
		m := []byte(args[0].Get("data").String())
		fmt.Println(m)
		t := game.MessageType(m[0])
		m = m[1:]

		// TODO handle error
		_ = handleSocketMsg(t, m)
		return nil
	})
}

func newSocketWrapper() js.Func {
	onOpen := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("websocket open")
		// TODO: send init here
		m, _ := game.NewSocketMsg(game.MsgJoinGame, "NAME")
		websocketSend(m)
		websocketOnOk(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			websocketSendEmpty(game.MsgStart)
			return nil
		}))
		return nil
	})
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		wsPrefix := "ws://"

		loc := js.Global().Get("window").Get("location")
		if !loc.Truthy() {
			// TODO handle error
		}
		protocol := loc.Get("protocol")
		if protocol.String() == "https:" {
			wsPrefix = "wss://"
		}
		host := loc.Get("host").String()

		fmt.Println(wsPrefix + host + "/connect")
		ws := js.Global().Get("WebSocket").New(wsPrefix + host + "/connect")
		js.Global().Set("socket", ws)
		ws.Call("addEventListener", "message", websocketGet())
		ws.Call("addEventListener", "open", onOpen)
		return nil
	})
}

func main() {
	setUpPage()
	newSocketWrapper().Invoke()
	<-make(chan bool)
}