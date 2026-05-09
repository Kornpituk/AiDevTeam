package main

import (
	"log"
	"net/http"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/config"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/server"
)

func main() {
	cfg := config.Load()
	srv := server.NewServer(cfg)

	log.Printf("Starting server on %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, srv.Router()); err != nil {
		log.Fatal(err)
	}
}