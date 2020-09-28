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

type Server struct {
	ga       *game.GameAssigner
	serveMux *http.ServeMux
}

func NewServer() *Server {
	ga := game.NewGameAssigner()
	serveMux := http.NewServeMux()

	s := &Server{
		ga:       ga,
		serveMux: serveMux,
	}
	fileserver := http.FileServer(http.Dir("public"))
	redirect := http.RedirectHandler("public/game.html", http.StatusFound)
	serveMux.Handle("/", redirect)
	serveMux.Handle("/public/", http.StripPrefix("/public/", fileserver))
	serveMux.HandleFunc("/connect", s.newConnection)

	return s
}

func (s *Server) newConnection(w http.ResponseWriter, req *http.Request) {
	log.Println("Handling new client.")
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("error making connection:", err)
		return
	}
	s.ga.StartNewClient(conn)
}

func main() {
	server := NewServer()
	go server.ga.Run()
	game.InitDictionary()

	const addr = ":8888"
	s := &http.Server{
		Addr:    addr,
		Handler: server.serveMux,
	}
	log.Println("Now listening on", addr)
	log.Fatal(s.ListenAndServe())
}
