package main

// Simple script to generate a JWT token
// make sure `GRAPHQL_JWT_SECRET` environment variable is set
// example usage: GRAPHQL_JWT_SECRET=civiliscool ./main someaddress@gmail.com

import (
	"fmt"
	"log"
	"os"

	"github.com/joincivil/civil-api-server/pkg/auth"
)

func main() {
	secret := []byte(os.Getenv("GRAPHQL_JWT_SECRET"))
	jwt := auth.NewJwtTokenGenerator(secret)
	// Generate a new code for a user
	token, err := jwt.GenerateToken(os.Args[1], 40000)
	if err != nil {
		log.Fatalf("Error generating token: err: %v", err)
	}

	fmt.Printf("JWT Token for %v: %v\n", os.Args[1], token)
}
