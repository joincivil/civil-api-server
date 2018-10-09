package auth

import (
	"context"
	"net/http"
)

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
var tokenCtxKey = &contextKey{"token"}

type contextKey struct {
	name string
}

// Token is a decoded JWT auth token
type Token struct {
	Sub     string
	IsAdmin bool
}

// Middleware decodes the `authorization` header jwt token and puts into context
func Middleware(jwt *JwtTokenGenerator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("authorization")

			// Allow unauthenticated users in
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			token, err := validateDecodeToken(jwt, authHeader)
			if err != nil {
				http.Error(w, "Invalid authorization header", http.StatusForbidden)
				return
			}

			// put it in context
			ctx := context.WithValue(r.Context(), tokenCtxKey, token)

			// and call the next with our new context
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func validateDecodeToken(jwt *JwtTokenGenerator, authHeader string) (*Token, error) {

	tokenRune := []rune(authHeader)

	// Start after "Bearer "
	tokenStr := string(tokenRune[7:])

	claims, err := jwt.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}

	token := Token{
		Sub:     claims["sub"].(string),
		IsAdmin: true,
	}
	return &token, nil
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *Token {
	raw, _ := ctx.Value(tokenCtxKey).(*Token)
	return raw
}
