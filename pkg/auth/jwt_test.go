package auth_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/auth"
)

func TestGenerateToken(t *testing.T) {
	generator := auth.NewJwtTokenGenerator([]byte("TestSecret"))
	token, err := generator.GenerateToken("my id", 1000)

	if err != nil {
		t.Fatalf("error thrown: %s", err)
	}
	claims, err := generator.ValidateToken(token)
	if err != nil {
		t.Fatalf("error thrown: %s", err)
	}

	if claims["sub"].(string) != "my id" {
		t.Fatalf("could not extract sub claim or is not `my id`")
	}
}
