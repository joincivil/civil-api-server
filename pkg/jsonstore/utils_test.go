package jsonstore_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

func TestTokenPlusIDHashKey(t *testing.T) {
	// Make sure the hashes are always the same for the same value
	token := &auth.Token{
		Sub:     "peter@civil.co",
		IsAdmin: false,
	}
	theID := "thisisatestID"

	key, err := jsonstore.TokenPlusIDHashKey(token, theID)
	if err != nil {
		t.Fatalf("Should have not returned an error while creating key: err: %v", err)
	}
	if key == "" {
		t.Errorf("Should have not have returned an empty string: err: %v", err)
	}

	token = &auth.Token{
		Sub:     "pete@joincivil.com",
		IsAdmin: false,
	}
	theID = "thisisatestID1000"

	key2, err := jsonstore.TokenPlusIDHashKey(token, theID)
	if err != nil {
		t.Fatalf("Should have not returned an error while creating key: err: %v", err)
	}
	if key2 == "" {
		t.Errorf("Should have not have returned an empty string: err: %v", err)
	}

	if key == key2 {
		t.Errorf("Keys should not have been identical")
	}
}
