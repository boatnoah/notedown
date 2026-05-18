package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	"github.com/boatnoah/notedown/internal/config"
	"github.com/boatnoah/notedown/internal/documents"
)

// Handler wires auth providers into HTTP handlers.
type Handler struct {
	cfg  config.Config
	docs *documents.Service
}

func NewHandler(cfg config.Config, docs *documents.Service) *Handler {
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = false
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		goth.UseProviders(google.New(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.AuthCallbackURL))
	} else {
		log.Println("google oauth credentials missing; login disabled")
	}

	return &Handler{cfg: cfg, docs: docs}
}

// BeginAuth routes to provider-specific authentication flows.
func (h *Handler) BeginAuth(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		http.Error(w, "missing provider", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	q.Set("provider", provider)
	r.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(w, r)
}

// Callback completes OAuth and provisions a collaborative room for the user.
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		http.Error(w, "missing provider", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	q.Set("provider", provider)
	r.URL.RawQuery = q.Encode()

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Printf("oauth callback error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ownerID := user.UserID
	if ownerID == "" {
		ownerID = user.Email
	}
	if ownerID == "" {
		ownerID = uuid.NewString()
	}

	doc, err := h.docs.CreateDocument(r.Context(), ownerID)
	if err != nil {
		log.Printf("create document failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirect := fmt.Sprintf("%s/editor?room=%s", h.cfg.FrontendURL, doc.ID)
	http.Redirect(w, r, redirect, http.StatusFound)
}
