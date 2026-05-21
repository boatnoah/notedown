package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/boatnoah/notedown/internal/users"
	"github.com/boatnoah/notedown/pkg/types"
)

type RegisterHandler struct {
	svc *users.Service
}

func NewRegisterHandler(svc *users.Service) *RegisterHandler {
	return &RegisterHandler{svc: svc}
}

type registerRequest struct {
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	Username string          `json:"username"`
	Password string          `json:"password"`
	Pfp      types.PfpPreset `json:"pfp"`
}

const maxRequestBytes = 64 * 1024 // 64 KB — generous for a registration payload

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req registerRequest
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
		}
		return
	}
	if err := dec.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := h.svc.Register(r.Context(), users.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
		Pfp:      req.Pfp,
	})
	if err != nil {
		status, msg := registerErrorStatus(err)
		http.Error(w, msg, status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

func registerErrorStatus(err error) (int, string) {
	switch {
	case errors.Is(err, users.ErrMissingFields),
		errors.Is(err, users.ErrInvalidEmail),
		errors.Is(err, users.ErrWeakPassword),
		errors.Is(err, users.ErrInvalidPfp),
		errors.Is(err, users.ErrDuplicateEmail),
		errors.Is(err, users.ErrDuplicateUsername):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
