package jsonstore_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/testutils"

	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

const (
	testJSONStr = `{
		"testkey1": "testval1",
		"testkey2": 1000,
		"testkey3": "testval3",
		"testkey4": {
			"nested1": "nestedval1"
		},
		"testkey5": [
			"listval1",
			"listval2",
			"listval3"
		],
		"testkey6": [
			{
				"listnested1": 1000,
				"listnested1": "listval1"
			}
		]
	}`
	testInvalidJSONStr = `{
		"testkey1": "testval1",
		"testkey2": 1000
		"testkey3": "testval3",
		"testkey4": {
			"nested1": "nestedval1"
		},
	}`
)

func TestSaveRetrieveJSONb(t *testing.T) {
	persister := &testutils.InMemoryJSONbPersister{
		Store: map[string]*jsonstore.JSONb{},
	}
	jsonbService := jsonstore.NewJsonbService(persister)

	testID := "testid"
	namespace := "somenamespaceid"

	jsonb, err := jsonbService.SaveRawJSONb(testID, namespace,
		jsonstore.NoSaltValue, testJSONStr)
	if err != nil {
		t.Errorf("Should have saved to the JSONb store: err: %v", err)
	}

	if jsonb.ID != testID {
		t.Errorf("Should have returned jsonb with correct testID")
	}
	if jsonb.RawJSON != testJSONStr {
		t.Errorf("Should have returned the exact raw JSON string")
	}

	jsonbs, err := jsonbService.RetrieveJSONb(testID, namespace, jsonstore.NoSaltValue)
	if err != nil {
		t.Errorf("Should have retrieved from the JSONb store: err: %v", err)
	}
	if len(jsonbs) != 1 {
		t.Errorf("Should have retrieved just 1 item from the JSONb store")
	}

	retrievedJson := jsonbs[0]

	if retrievedJson.Hash != jsonb.Hash {
		t.Errorf("Should have seen the same item hashes")
	}
	if retrievedJson.CreatedDate != jsonb.CreatedDate {
		t.Errorf("Should have seen the same item creation dates")
	}
	if retrievedJson.Key != jsonb.Key {
		t.Errorf("Should have seen the same key")
	}
	if retrievedJson.ID != jsonb.ID {
		t.Errorf("Should have seen the same item id")
	}
	if retrievedJson.RawJSON != jsonb.RawJSON {
		t.Errorf("Should have seen the same item raw json")
	}
	if retrievedJson.Namespace != jsonb.Namespace {
		t.Errorf("Should have been the correct namespace")
	}
	if retrievedJson.Namespace != namespace {
		t.Errorf("Should have been the correct namespace")
	}

}

func TestInvalidJSONStr(t *testing.T) {
	persister := &testutils.InMemoryJSONbPersister{
		Store: map[string]*jsonstore.JSONb{},
	}
	jsonbService := jsonstore.NewJsonbService(persister)

	testID := "testid"
	namespace := "somenamespaceid"

	_, err := jsonbService.SaveRawJSONb(testID, namespace, jsonstore.NoSaltValue,
		testInvalidJSONStr)
	if err == nil {
		t.Errorf("Should have received error with invalid JSON: err: %v", err)
	}

	_, err = jsonbService.SaveRawJSONb(testID, namespace, jsonstore.NoSaltValue, "")
	if err == nil {
		t.Errorf("Should have received error with empty JSON: err: %v", err)
	}
}
