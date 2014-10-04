package main

import (
	"log"
	"net/http"
)

func main() {
	const addr = "localhost:8080"
	fileserver := http.FileServer(http.Dir("public"))
	redirect := http.RedirectHandler("public/scrabble.html", http.StatusFound)

	http.Handle("/", redirect)
	http.Handle("/public/", http.StripPrefix("/public/", fileserver))

	log.Println("Now listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
