package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
)

type LogoutHandler struct {
	sessions SessionRepository
}

func NewLogoutHandler(sessions SessionRepository) *LogoutHandler {
	return &LogoutHandler{sessions: sessions}
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sum := sha256.Sum256([]byte(cookie.Value))
	hash := hex.EncodeToString(sum[:])

	if err := h.sessions.DeleteByTokenHash(r.Context(), hash); err != nil && !errors.Is(err, ErrSessionNotFound) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, refreshCookie("", -1, isSecure(r)))
	w.WriteHeader(http.StatusNoContent)
}
