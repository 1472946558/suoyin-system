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

	log.Printf("gold recycle mock API listening on :%s (mode=memory, stage=week1)", server.Config.Port)
	if err := http.ListenAndServe(":"+server.Config.Port, server.Router()); err != nil {
		log.Fatal(err)
	}
}
