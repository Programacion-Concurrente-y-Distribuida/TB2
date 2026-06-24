package apiserver

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtMiddleware validates the Authorization: Bearer <token> header.
// On success it calls next; on failure it returns 401.
func (s *Server) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "token requerido")
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if err := validateToken(tokenStr, s.cfg.JWTSecret); err != nil {
			writeError(w, http.StatusUnauthorized, "token inválido")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func generateToken(username, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func validateToken(tokenStr, secret string) error {
	_, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	}, jwt.WithExpirationRequired())
	return err
}
