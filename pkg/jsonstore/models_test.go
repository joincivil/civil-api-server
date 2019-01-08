package jsonstore_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

const (
	validTestJSON = `{
		"test": "value",
		"test1": 1000,
		"test2": {
			"test4": 100
		},
		"test5": [
			"list1",
			"list2",
			"list3"
		],
		"test6": [
			{
				"item": 1
			},
			{
				"item": 2
			},
			{
				"item": 3
			},
			{
				"item": 5
			}
		]
	}`

	invalidTestJSON = `{
		"test": "test1",
		"test1": 100,
		"test2": {},
	}`
)

func TestValidateRawJSON(t *testing.T) {
	jsb := &jsonstore.JSONb{}
	jsb.RawJSON = validTestJSON
	err := jsb.ValidateRawJSON()
	if err != nil {
		t.Errorf("Should have validated the JSON for valid JSON: err: %v", err)
	}
}

func TestInvalidValidateRawJSON(t *testing.T) {
	jsb := &jsonstore.JSONb{}
	jsb.RawJSON = invalidTestJSON
	err := jsb.ValidateRawJSON()
	if err == nil {
		t.Errorf("Should have failed on validation for invalid JSON")
	}
}

func TestRawJSONToFields(t *testing.T) {
	jsb := &jsonstore.JSONb{}
	jsb.RawJSON = validTestJSON
	err := jsb.RawJSONToFields()
	if err != nil {
		t.Errorf("Should not have failed while creating the JSONFields: err: %v", err)
	}
	if jsb.JSON == nil {
		t.Errorf("Should have created the JSON fields")
	}
	if len(jsb.JSON) != 5 {
		t.Errorf("Should have 5 fields: len: %v", len(jsb.JSON))
	}
}

func TestHashIDRawJSON(t *testing.T) {
	jsb := &jsonstore.JSONb{}
	jsb.RawJSON = validTestJSON
	err := jsb.HashIDRawJSON()
	if err != nil {
		t.Errorf("Should not have failed while hashing json/id: err: %v", err)
	}
	if jsb.Hash == "" {
		t.Errorf("Should not have a blank hash after hashing")
	}
	if len(jsb.Hash) != 64 {
		t.Errorf("Should have been length 64")
	}
}

func TestInvalidHashIDRawJSON(t *testing.T) {
	jsb := &jsonstore.JSONb{}
	err := jsb.HashIDRawJSON()
	if err == nil {
		t.Errorf("Should have failed while hashing empty raw json: err: %v", err)
	}
	jsb.RawJSON = invalidTestJSON
	err = jsb.HashIDRawJSON()
	if err == nil {
		t.Errorf("Should have failed while hashing the json/id")
	}
}

func TestInvalidRawJSONToFields(t *testing.T) {
	jsb := &jsonstore.JSONb{}
	jsb.RawJSON = invalidTestJSON
	err := jsb.RawJSONToFields()
	if err == nil {
		t.Errorf("Should have failed while creating the JSONFields: err: %v", err)
	}
}
