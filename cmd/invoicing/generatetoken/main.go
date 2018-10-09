package main

// Simple script to generate a referral code.  Used to generate them
// for newsrooms or for one-offs.  Does not store this value anywhere, so need to
// record it elsewhere.

import (
	"fmt"
	"os"

	"github.com/joincivil/civil-api-server/pkg/auth"
)

func main() {
	secret := []byte(os.Getenv("GRAPHQL_JWT_SECRET"))
	jwt := auth.NewJwtTokenGenerator(secret)
	// Generate a new code for a user
	token, _ := jwt.GenerateToken(os.Args[1], 40000)

	fmt.Printf("JWT Token for %v: %v\n", os.Args[1], token)
}
