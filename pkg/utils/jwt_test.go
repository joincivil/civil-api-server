package utils_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/utils"
)

func TestGenerateToken(t *testing.T) {
	generator := utils.NewJwtTokenGenerator([]byte("TestSecret"))
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

func TestRefreshToken(t *testing.T) {
	generator := utils.NewJwtTokenGenerator([]byte("TestSecret"))
	token, err := generator.GenerateRefreshToken("my id")

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

	if claims["aud"].(string) != "refresh" {
		t.Fatalf("audience is not `refresh`")
	}

	if claims["exp"] != nil {
		t.Fatalf("should not have an `exp` claim but it is %v", claims["exp"].(string))
	}
}
