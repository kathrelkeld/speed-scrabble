package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/kathrelkeld/speed-scrabble/game"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func newConnection(w http.ResponseWriter, req *http.Request) {
	log.Println("Handling new client.")
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("error making connection:", err)
		return
	}
	c := game.NewClient(conn)
	go c.Run()
	go c.ReadSocketMsgs()
}

func cleanup() {
	close(game.NewGameChan)
	game.GlobalGame.Cleanup()
}

func main() {
	defer cleanup()
	go game.GlobalGame.Run()
	go game.AddAPlayerToAGame()

	const port = ":8888"
	fileserver := http.FileServer(http.Dir("public"))
	redirect := http.RedirectHandler("public/game.html", http.StatusFound)

	http.Handle("/", redirect)
	http.Handle("/public/", http.StripPrefix("/public/", fileserver))
	http.HandleFunc("/connect", newConnection)

	log.Println("Now listening on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
