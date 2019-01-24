package auth

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/golang/glog"

	jwtgo "github.com/dgrijalva/jwt-go"
)

const (
	middlewareInvalidTokenCode = 500
	middlewareExpiredTokenCode = 501
)

const (
	middlewareInvalidTokenMsg = "Invalid authorization header"
	middlewareExpiredTokenMsg = "Authorization has expired"
)

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
var tokenCtxKey = &contextKey{"token"}

type contextKey struct {
	name string
}

// Middleware decodes the `authorization` header jwt token and puts into context
// The authorization header must be of format
// "Authorization: Bearer <JWT token>"
func Middleware(jwt *JwtTokenGenerator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("authorization")

			// Allow unauthenticated users in
			if authHeader == "" {
				log.Infof("No authorization header")
				next.ServeHTTP(w, r)
				return
			}

			token, err := validateDecodeToken(jwt, authHeader)
			if err != nil {
				code, msg := parseValidationErrorToCodeMsg(err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				respBody := fmt.Sprintf("{\"errCode\": \"%v\", \"err\": \"%v\"}", code, msg)
				_, _ = w.Write([]byte(respBody)) // nolint: gosec
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

func parseValidationErrorToCodeMsg(err error) (int, string) {
	// Default to invalid token
	code := middlewareInvalidTokenCode
	msg := middlewareInvalidTokenMsg

	ve, ok := err.(*jwtgo.ValidationError)

	// If the token is valid, but expired
	// If there is a validation error that isn't an expiration, log it
	if ok && ve.Errors&jwtgo.ValidationErrorExpired != 0 {
		code = middlewareExpiredTokenCode
		msg = middlewareExpiredTokenMsg

	} else {
		log.Infof("token validation error: err: %v", err)
	}
	return code, msg
}

func validateDecodeToken(jwt *JwtTokenGenerator, authHeader string) (*Token, error) {

	tokenRune := []rune(authHeader)

	if len(tokenRune) < 7 {
		return nil, fmt.Errorf("Token length is invalid")
	}

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
