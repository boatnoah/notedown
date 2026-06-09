package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/boatnoah/notedown/internal/auth"
	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/internal/realtime"
	"github.com/boatnoah/notedown/pkg/types"
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
	JWTSecret       string
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
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Per-IP rate limits guard the auth endpoints against brute force and
	// token hammering. Limits are scoped per route so other endpoints are
	// unaffected.
	loginLimit := newRateLimiter(10)
	registerLimit := newRateLimiter(5)
	refreshLimit := newRateLimiter(30)

	r.Route("/auth", func(r chi.Router) {
		r.With(registerLimit.Middleware).Post("/register", deps.RegisterHandler.ServeHTTP)
		r.With(loginLimit.Middleware).Post("/login", deps.LoginHandler.ServeHTTP)
		r.With(refreshLimit.Middleware).Post("/refresh", deps.RefreshHandler.ServeHTTP)
		r.Post("/logout", deps.LogoutHandler.ServeHTTP)
	})

	r.Route("/documents", func(r chi.Router) {
		r.Use(auth.RequireAuth(deps.JWTSecret))
		r.Post("/", createDocumentHandler(deps.DocumentService))
		r.Get("/", listDocumentsHandler(deps.DocumentService))
		r.Delete("/{id}", deleteDocumentHandler(deps.DocumentService))
		r.Get("/{id}", getDocumentHandler(deps.DocumentService))
		r.Get("/{id}/meta", getDocumentMetaHandler(deps.DocumentService))
		r.Post("/{id}/share", updateShareModeHandler(deps.DocumentService))
	})

	r.Get("/ws", deps.RealtimeHub.HandleWebsocket)

	return r
}

func createDocumentHandler(svc *documents.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := auth.IdentityFromContext(r.Context())
		if !ok {
			http.Error(w, "missing access token", http.StatusUnauthorized)
			return
		}
		doc, err := svc.CreateDocument(r.Context(), identity.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		respondJSON(w, doc)
	}
}

// authorizeDocumentRead loads the document and verifies the caller may view
// it, writing the appropriate error response otherwise.
func authorizeDocumentRead(w http.ResponseWriter, r *http.Request, svc *documents.Service) (*types.Document, bool) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		http.Error(w, "missing access token", http.StatusUnauthorized)
		return nil, false
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return nil, false
	}
	doc, err := svc.GetDocument(r.Context(), id)
	if err != nil {
		if errors.Is(err, documents.ErrNotFound) {
			http.Error(w, "document not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return nil, false
	}
	if !doc.CanRead(identity.UserID) {
		http.Error(w, "you do not have access to this document", http.StatusForbidden)
		return nil, false
	}
	return doc, true
}

func listDocumentsHandler(svc *documents.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := auth.IdentityFromContext(r.Context())
		if !ok {
			http.Error(w, "missing access token", http.StatusUnauthorized)
			return
		}
		docs, err := svc.ListDocuments(r.Context(), identity.UserID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if docs == nil {
			docs = []*types.Document{}
		}
		respondJSON(w, docs)
	}
}

func deleteDocumentHandler(svc *documents.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := auth.IdentityFromContext(r.Context())
		if !ok {
			http.Error(w, "missing access token", http.StatusUnauthorized)
			return
		}
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		if err := svc.DeleteDocument(r.Context(), id, identity.UserID); err != nil {
			switch {
			case errors.Is(err, documents.ErrNotOwner):
				http.Error(w, "forbidden", http.StatusForbidden)
			case errors.Is(err, documents.ErrNotFound):
				http.Error(w, "document not found", http.StatusNotFound)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func getDocumentHandler(svc *documents.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doc, ok := authorizeDocumentRead(w, r, svc)
		if !ok {
			return
		}
		snapshot, err := svc.Snapshot(r.Context(), doc.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		respondJSON(w, snapshot)
	}
}

func getDocumentMetaHandler(svc *documents.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doc, ok := authorizeDocumentRead(w, r, svc)
		if !ok {
			return
		}
		respondJSON(w, doc)
	}
}

func updateShareModeHandler(svc *documents.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := auth.IdentityFromContext(r.Context())
		if !ok {
			http.Error(w, "missing access token", http.StatusUnauthorized)
			return
		}
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		var req struct {
			ShareMode types.ShareMode `json:"shareMode"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		doc, err := svc.SetShareMode(r.Context(), id, identity.UserID, req.ShareMode)
		if err != nil {
			switch {
			case errors.Is(err, documents.ErrInvalidShareMode):
				http.Error(w, "shareMode must be one of: private, read, edit", http.StatusBadRequest)
			case errors.Is(err, documents.ErrNotOwner):
				http.Error(w, "only the document owner can change sharing", http.StatusForbidden)
			case errors.Is(err, documents.ErrNotFound):
				http.Error(w, "document not found", http.StatusNotFound)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
		respondJSON(w, doc)
	}
}

func respondJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
