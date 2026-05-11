package main

import (
	"log"
	"net/http"

	"gold-recycle-miniapp/backend/internal/app"
)

func main() {
	server, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Printf("gold recycle API listening on %s (mode=%s)", server.Config.ListenAddr(), server.Config.Mode)
	if err := http.ListenAndServe(server.Config.ListenAddr(), server.Router()); err != nil {
		log.Fatal(err)
	}
}
