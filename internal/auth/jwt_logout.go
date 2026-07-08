package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid or expired token")

var jwtSecret []byte

func SetSecret(secret string) {
	jwtSecret = []byte(secret)
}

type Claims struct {
	AdminID string `json:"admin_id"`
	Role    string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for the given admin ID.
// Each token gets a unique jti (JWT ID) so it can be individually
// revoked later via the revoked_tokens table.
func GenerateToken(adminID int) (string, error) {
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := Claims{
		AdminID: strconv.Itoa(adminID),
		Role:    "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken validates a JWT string and returns its claims.
// Note: this only checks the signature and expiry — it does NOT check
// the revoked_tokens blacklist. Callers (middleware) must check that separately.
func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}