package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/boatnoah/notedown/internal/users"
)

// dummyHash is used to perform a constant-time bcrypt comparison when the
// requested email does not exist, preventing account enumeration via timing.
var dummyHash []byte

func init() {
	var err error
	dummyHash, err = bcrypt.GenerateFromPassword([]byte("dummy-constant-notedown"), bcrypt.DefaultCost)
	if err != nil {
		panic("auth: failed to initialize dummy bcrypt hash: " + err.Error())
	}
}

type LoginHandler struct {
	userRepo            users.Repository
	sessions            SessionRepository
	secret              string
	onUserAuthenticated func(ctx context.Context, userID string)
}

func NewLoginHandler(userRepo users.Repository, sessions SessionRepository, jwtSecret string) *LoginHandler {
	return &LoginHandler{userRepo: userRepo, sessions: sessions, secret: jwtSecret}
}

// SetOnUserAuthenticated registers a hook invoked asynchronously after each
// successful login. Must be called before ServeHTTP is used.
func (h *LoginHandler) SetOnUserAuthenticated(fn func(ctx context.Context, userID string)) {
	h.onUserAuthenticated = fn
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"accessToken"`
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req loginRequest
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

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	user, hash, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			// Always run bcrypt to prevent timing-based account enumeration.
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := issueAccessToken(user, h.secret)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	refreshToken, refreshHash, err := generateRefreshToken()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	session := &AuthSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        time.Now().Add(refreshTokenTTL),
	}
	if err := h.sessions.Create(r.Context(), session); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, refreshCookie(refreshToken, int(refreshTokenTTL.Seconds()), isSecure(r)))

	if fn := h.onUserAuthenticated; fn != nil {
		baseCtx := context.WithoutCancel(r.Context())
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("onUserAuthenticated panic: %v", rec)
				}
			}()
			ctx, cancel := context.WithTimeout(baseCtx, 30*time.Second)
			defer cancel()
			fn(ctx, user.ID)
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResponse{AccessToken: accessToken})
}
