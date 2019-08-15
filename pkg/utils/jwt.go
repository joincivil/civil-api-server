package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/golang/glog"
)

// JwtTokenGenerator is responsible for generating JWT authentication tokens.
type JwtTokenGenerator struct {
	Secret []byte
}

// NewJwtTokenGenerator creates a new instance of JwtTokenGenerator
func NewJwtTokenGenerator(secret []byte) *JwtTokenGenerator {
	return &JwtTokenGenerator{
		Secret: secret,
	}
}

// GenerateToken generates a new token for the desired user and signs it with a secret
// this is only valid until `ExpiresAt`, at which point `ValidateToken` will begin to throw an error
func (gen *JwtTokenGenerator) GenerateToken(sub string, expires int) (string, error) {

	var claims = &jwt.StandardClaims{
		Subject:   sub,
		ExpiresAt: time.Now().Unix() + int64(expires),
		Issuer:    "Civil",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(gen.Secret)

	return tokenString, err
}

// GenerateRefreshToken generates a new refresh token for the desired user and signs it with a secret
// this token does not have an expiration, and is used to generate a new expiring token
func (gen *JwtTokenGenerator) GenerateRefreshToken(id string) (string, error) {
	var claims = &jwt.StandardClaims{
		Subject:  id,
		IssuedAt: time.Now().Unix(),
		Issuer:   "Civil",
		Audience: "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(gen.Secret)

	return tokenString, err
}

// ValidateToken confirms that the token passed in matches the expected signature and has not expired
// returns a list of JWT "claims" embedded in the token (issuer, expiration, user, etc)
func (gen *JwtTokenGenerator) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return gen.Secret, nil
	})

	if err != nil {
		return jwt.MapClaims{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	log.Errorf("Error validating token: err: %v", err)
	return nil, err

}

// RefreshToken uses a refresh token to genrate an new expiring token
func (gen *JwtTokenGenerator) RefreshToken(refresh string, expires int) (string, error) {
	claims, err := gen.ValidateToken(refresh)

	if err != nil {
		return "", err
	}

	return gen.GenerateToken(claims["sub"].(string), expires)
}
