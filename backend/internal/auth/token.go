package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/boatnoah/notedown/pkg/types"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
)

type Claims struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Pfp      string `json:"pfp"`
	jwt.RegisteredClaims
}

func issueAccessToken(user *types.User, secret string) (string, error) {
	claims := Claims{
		Name:     user.Name,
		Username: user.Username,
		Pfp:      string(user.Pfp),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// generateRefreshToken returns a cryptographically random token and its SHA-256 hash.
func generateRefreshToken() (token string, hash string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return
	}
	token = hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	hash = hex.EncodeToString(sum[:])
	return
}
