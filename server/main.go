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

var ga *game.GameAssigner

func newConnection(w http.ResponseWriter, req *http.Request) {
	log.Println("Handling new client.")
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("error making connection:", err)
		return
	}
	game.StartNewClient(conn, ga)
}

func main() {
	ga = game.StartGameAssigner()
	defer ga.Close()

	const port = ":8888"
	fileserver := http.FileServer(http.Dir("public"))
	redirect := http.RedirectHandler("public/game.html", http.StatusFound)

	http.Handle("/", redirect)
	http.Handle("/public/", http.StripPrefix("/public/", fileserver))
	http.HandleFunc("/connect", newConnection)

	log.Println("Now listening on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
