package jsonstore_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

func TestNamespaceIDSaltHashKey(t *testing.T) {
	namespace := "namespace1"
	theID := "thisisatestID"
	key1, err := jsonstore.NamespaceIDSaltHashKey(namespace, theID, jsonstore.NoSaltValue)
	if err != nil {
		t.Fatalf("Should have not returned an error while creating key: err: %v", err)
	}
	if key1 == "" {
		t.Errorf("Should have not have returned an empty string: err: %v", err)
	}

	namespace = "namespace2"
	theID = "thisisatestID"
	key2, err := jsonstore.NamespaceIDSaltHashKey(namespace, theID, jsonstore.NoSaltValue)
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
func TestNamespaceIDSaltHashKeyWithSalt(t *testing.T) {
	salt := "thisisasalt"

	namespace := "namespace1"
	theID := "thisisatestID"
	key1, err := jsonstore.NamespaceIDSaltHashKey(namespace, theID, salt)
	if err != nil {
		t.Fatalf("Should have not returned an error while creating key: err: %v", err)
	}
	if key1 == "" {
		t.Errorf("Should have not have returned an empty string: err: %v", err)
	}

	namespace = "namespace1"
	theID = "thisisatestID"
	key3, err := jsonstore.NamespaceIDSaltHashKey(namespace, theID, "differentsalt")
	if err != nil {
		t.Fatalf("Should have not returned an error while creating key: err: %v", err)
	}
	if key3 == "" {
		t.Errorf("Should have not have returned an empty string: err: %v", err)
	}

	if key1 == key3 {
		t.Errorf("Keys should not have been identical with a salt")
	}

	namespace = "namespace2"
	theID = "thisisatestID"
	key2, err := jsonstore.NamespaceIDSaltHashKey(namespace, theID, salt)
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
