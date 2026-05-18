package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/boatnoah/notedown/internal/auth"
	"github.com/boatnoah/notedown/internal/config"
	"github.com/boatnoah/notedown/internal/crdt"
	"github.com/boatnoah/notedown/internal/db"
	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/internal/realtime"
	"github.com/boatnoah/notedown/internal/server"
	"github.com/boatnoah/notedown/internal/storage/memory"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close()

	docRepo := memory.NewDocumentRepository()
	opRepo := memory.NewOperationRepository()
	sessionRepo := memory.NewSessionRepository()
	manager := crdt.NewManager()

	docService := documents.NewService(documents.Deps{
		Documents:  docRepo,
		Operations: opRepo,
		Sessions:   sessionRepo,
		Manager:    manager,
	})

	realtimeHub := realtime.NewHub(docService)
	authHandler := auth.NewHandler(cfg, docService)

	router := server.NewRouter(server.Dependencies{
		AuthHandler:     authHandler,
		DocumentService: docService,
		RealtimeHub:     realtimeHub,
	})

	log.Printf("listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, router))
}
