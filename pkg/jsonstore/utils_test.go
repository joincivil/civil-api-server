package jsonstore_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

func TestNamespacePlusIDHashKey(t *testing.T) {
	namespace := "peter@civil.co"
	theID := "thisisatestID"
	key1, err := jsonstore.NamespacePlusIDHashKey(namespace, theID)
	if err != nil {
		t.Fatalf("Should have not returned an error while creating key: err: %v", err)
	}
	if key1 == "" {
		t.Errorf("Should have not have returned an empty string: err: %v", err)
	}

	namespace = "pete@civil.co"
	theID = "thisisatestID"
	key2, err := jsonstore.NamespacePlusIDHashKey(namespace, theID)
	if err != nil {
		t.Fatalf("Should have not returned an error while creating key: err: %v", err)
	}
	if key2 == "" {
		t.Errorf("Should have not have returned an empty string: err: %v", err)
	}

	if key1 == key2 {
		t.Errorf("Keys should not have been identical")
	}

}
