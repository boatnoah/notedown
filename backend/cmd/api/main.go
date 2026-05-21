package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/boatnoah/notedown/internal/auth"
	"github.com/boatnoah/notedown/internal/config"
	"github.com/boatnoah/notedown/internal/db"
	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/internal/ot"
	"github.com/boatnoah/notedown/internal/realtime"
	"github.com/boatnoah/notedown/internal/server"
	"github.com/boatnoah/notedown/internal/storage/memory"
	"github.com/boatnoah/notedown/internal/storage/postgres"
	"github.com/boatnoah/notedown/internal/users"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func() { _ = database.Close() }()

	docRepo := postgres.NewDocumentRepository(database)
	opRepo := postgres.NewOperationRepository(database)
	sessionRepo := memory.NewSessionRepository() // ephemeral WS presence sessions, not persisted
	authSessionRepo := postgres.NewAuthSessionRepository(database)
	userRepo := postgres.NewUserRepository(database)
	manager := ot.NewManager()

	docService := documents.NewService(documents.Deps{
		Documents:  docRepo,
		Operations: opRepo,
		Sessions:   sessionRepo,
		Manager:    manager,
	})
	userService := users.NewService(userRepo)

	realtimeHub := realtime.NewHub(docService)
	authHandler := auth.NewHandler(cfg, docService)
	registerHandler := auth.NewRegisterHandler(userService)
	loginHandler := auth.NewLoginHandler(userRepo, authSessionRepo, cfg.JWTSecret)
	refreshHandler := auth.NewRefreshHandler(userRepo, authSessionRepo, cfg.JWTSecret)
	logoutHandler := auth.NewLogoutHandler(authSessionRepo)

	router := server.NewRouter(server.Dependencies{
		AuthHandler:     authHandler,
		RegisterHandler: registerHandler,
		LoginHandler:    loginHandler,
		RefreshHandler:  refreshHandler,
		LogoutHandler:   logoutHandler,
		DocumentService: docService,
		RealtimeHub:     realtimeHub,
	})

	log.Printf("listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, router))
}
