package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/boatnoah/notedown/internal/users"
)

type RefreshHandler struct {
	userRepo users.Repository
	sessions SessionRepository
	secret   string
}

func NewRefreshHandler(userRepo users.Repository, sessions SessionRepository, jwtSecret string) *RefreshHandler {
	return &RefreshHandler{userRepo: userRepo, sessions: sessions, secret: jwtSecret}
}

func (h *RefreshHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "missing refresh token", http.StatusUnauthorized)
		return
	}

	sum := sha256.Sum256([]byte(cookie.Value))
	hash := hex.EncodeToString(sum[:])
	secure := isSecure(r)

	session, err := h.sessions.GetByTokenHash(r.Context(), hash)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionExpired) {
			http.SetCookie(w, refreshCookie("", -1, secure))
			http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if time.Now().After(session.ExpiresAt) {
		_ = h.sessions.DeleteByTokenHash(r.Context(), hash)
		http.SetCookie(w, refreshCookie("", -1, secure))
		http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), session.UserID)
	if err != nil {
		// User was deleted — treat the session as invalid.
		_ = h.sessions.DeleteByTokenHash(r.Context(), hash)
		http.SetCookie(w, refreshCookie("", -1, secure))
		if errors.Is(err, users.ErrNotFound) {
			http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	newToken, newHash, err := generateRefreshToken()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Delete by token hash (not session ID) to guard against duplicate rows.
	if err := h.sessions.DeleteByTokenHash(r.Context(), hash); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	newSession := &AuthSession{
		UserID:           session.UserID,
		RefreshTokenHash: newHash,
		ExpiresAt:        time.Now().Add(refreshTokenTTL),
	}
	if err := h.sessions.Create(r.Context(), newSession); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	accessToken, err := issueAccessToken(user, h.secret)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, refreshCookie(newToken, int(refreshTokenTTL.Seconds()), secure))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResponse{AccessToken: accessToken})
}
