package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/boatnoah/notedown/internal/auth"
	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/internal/realtime"
)

// Dependencies enumerates collaborators needed to wire the HTTP server.
type Dependencies struct {
	RegisterHandler *auth.RegisterHandler
	LoginHandler    *auth.LoginHandler
	RefreshHandler  *auth.RefreshHandler
	LogoutHandler   *auth.LogoutHandler
	DocumentService *documents.Service
	RealtimeHub     *realtime.Hub
	FrontendURL     string
}

// NewRouter builds a chi router with all API endpoints mounted.
func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{deps.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", deps.RegisterHandler.ServeHTTP)
		r.Post("/login", deps.LoginHandler.ServeHTTP)
		r.Post("/refresh", deps.RefreshHandler.ServeHTTP)
		r.Post("/logout", deps.LogoutHandler.ServeHTTP)
	})

	r.Route("/documents", func(r chi.Router) {
		r.Post("/", createDocumentHandler(deps.DocumentService))
		r.Get("/{id}", getDocumentHandler(deps.DocumentService))
	})

	r.Get("/ws", deps.RealtimeHub.HandleWebsocket)

	return r
}

func createDocumentHandler(svc *documents.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerID := r.URL.Query().Get("owner")
		if ownerID == "" {
			ownerID = "anonymous"
		}
		doc, err := svc.CreateDocument(r.Context(), ownerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		respondJSON(w, doc)
	}
}

func getDocumentHandler(svc *documents.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		snapshot, err := svc.Snapshot(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		respondJSON(w, snapshot)
	}
}

func respondJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
