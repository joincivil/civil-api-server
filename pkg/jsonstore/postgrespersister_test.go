// +build integration

package jsonstore

import (
	"fmt"
	"testing"

	"github.com/joincivil/civil-api-server/pkg/auth"
)

const (
	postgresPort   = 5432
	postgresDBName = "civil_crawler"
	postgresUser   = "docker"
	postgresPswd   = "docker"
	postgresHost   = "localhost"
)

var (
	testToken = &auth.Token{
		Sub:     "peter@civil.co",
		IsAdmin: false,
	}
	testID         = "thisisavalidid"
	validTestJSONb = &JSONb{
		Key: "",
		ID:  testID,
		RawJSON: `{
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
		}`,
	}
)

func setupDBConnection() (*PostgresPersister, error) {
	return NewPostgresPersister(postgresHost, postgresPort, postgresUser, postgresPswd, postgresDBName)
}

func setupTestTable(tableName string) (*PostgresPersister, error) {
	persister, err := setupDBConnection()
	if err != nil {
		return persister, fmt.Errorf("Error connecting to DB: %v", err)
	}
	var queryString string
	switch tableName {
	case "jsonb_test":
		queryString = CreateJsonbTableQuery(tableName)
	}
	_, err = persister.db.Query(queryString)
	if err != nil {
		return persister, fmt.Errorf("Couldn't create test table %s: %v", tableName, err)
	}
	return persister, nil
}

func deleteTestTable(persister *PostgresPersister, tableName string) error {
	var err error
	switch tableName {
	case "jsonb_test":
		_, err = persister.db.Query("DROP TABLE jsonb_test;")
	}
	if err != nil {
		return fmt.Errorf("Couldn't delete test table %s: %v", tableName, err)
	}
	return nil
}

func checkTableExists(tableName string, persister *PostgresPersister) error {
	var exists bool
	queryString := fmt.Sprintf(`SELECT EXISTS ( SELECT 1
        FROM   information_schema.tables 
        WHERE  table_schema = 'public'
        AND    table_name = '%s'
        );`, tableName)
	err := persister.db.QueryRow(queryString).Scan(&exists)
	if err != nil {
		return fmt.Errorf("Couldn't get %s table", tableName)
	}
	if !exists {
		return fmt.Errorf("%s table does not exist", tableName)
	}
	return nil
}

// TestDBConnection tests that we can connect to DB
func TestDBConnection(t *testing.T) {
	persister, err := setupDBConnection()
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	var result int
	err = persister.db.QueryRow("SELECT 1;").Scan(&result)
	if err != nil {
		t.Errorf("Error querying DB: %v", err)
	}
	if result != 1 {
		t.Errorf("Wrong result from DB")
	}
}

// TestTableSetup tests to ensure that our DB tables are being setup
func TestTableSetup(t *testing.T) {
	// run function to create tables, and test table exists
	persister, err := setupDBConnection()
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	err = persister.CreateTables()
	if err != nil {
		t.Errorf("Error creating tables: %v", err)
	}
	err = checkTableExists("jsonb", persister)
	if err != nil {
		t.Error(err)
	}
}

// TestTableSetup tests to ensure that our DB tables are being setup
func TestIndicesSetup(t *testing.T) {
	tableName := "jsonb_test"
	// create fake listing in listing_test
	persister, err := setupTestTable(tableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, tableName)
	err = persister.createJsonbIndicesForTable(tableName)
	if err != nil {
		t.Errorf("Error creating indices for jsonb: %v", err)
	}
}

func TestSaveJsonb(t *testing.T) {
	tableName := "jsonb_test"
	// create fake listing in listing_test
	persister, err := setupTestTable(tableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, tableName)

	newJsonb := validTestJSONb
	newJsonb.Key, err = NamespacePlusIDHashKey(testToken.Sub, testID)
	if err != nil {
		t.Error("Should not have failed creating the key")
	}
	newJsonb.HashIDRawJSON()
	newJsonb.RawJSONToFields()

	_, err = persister.saveJsonbForTable(newJsonb, tableName)
	if err != nil {
		t.Errorf("Should not have received an error saving jsonb: err: %v", err)
	}

	// check that jsonb is there
	var numRowsb int
	err = persister.db.QueryRow(`SELECT COUNT(*) FROM jsonb_test`).Scan(&numRowsb)
	if numRowsb != 1 {
		t.Errorf("Number of rows in table should be 0 but is: %v", numRowsb)
	}
}

func TestRetrieveJsonb(t *testing.T) {
	tableName := "jsonb_test"
	// create fake listing in listing_test
	persister, err := setupTestTable(tableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, tableName)

	_, err = persister.jsonbFromTable("thisisavalidid", "", tableName)
	if err == nil {
		t.Error("Should have failed since nothing in table")
	}
	if err != ErrNoJsonbFound {
		t.Error("Should have failed with no jsonb found, not normal error")
	}

	newJsonb := validTestJSONb
	newJsonb.Key, err = NamespacePlusIDHashKey(testToken.Sub, testID)
	if err != nil {
		t.Error("Should not have failed creating the key")
	}
	newJsonb.HashIDRawJSON()
	newJsonb.RawJSONToFields()
	_, err = persister.saveJsonbForTable(newJsonb, tableName)
	if err != nil {
		t.Errorf("Should not have received an error saving jsonb: err: %v", err)
	}

	retrievedJsonb, err := persister.jsonbFromTable(newJsonb.Key, "", tableName)
	if err != nil {
		t.Error("Should have not failed since there should be value in table")
	}
	if len(retrievedJsonb) != 1 {
		t.Error("Should have only one result in table")
	}
	firstResult := retrievedJsonb[0]
	if newJsonb.ID != firstResult.ID {
		t.Error("Should have matching IDs")
	}
	if newJsonb.Hash != firstResult.Hash {
		t.Error("Should have matching Hashes")
	}

	retrievedJsonb, err = persister.jsonbFromTable("", newJsonb.Hash, tableName)
	if err != nil {
		t.Error("Should have not failed since there should be value in table")
	}
	if len(retrievedJsonb) != 1 {
		t.Error("Should have only one result in table")
	}
	firstResult = retrievedJsonb[0]
	if newJsonb.ID != firstResult.ID {
		t.Error("Should have matching IDs")
	}
	if newJsonb.Hash != firstResult.Hash {
		t.Error("Should have matching Hashes")
	}

	retrievedJsonb, err = persister.jsonbFromTable(newJsonb.Key, newJsonb.Hash, tableName)
	if err != nil {
		t.Error("Should have not failed since there should be value in table")
	}
	if len(retrievedJsonb) != 1 {
		t.Error("Should have only one result in table")
	}
	firstResult = retrievedJsonb[0]
	if newJsonb.ID != firstResult.ID {
		t.Error("Should have matching IDs")
	}
	if newJsonb.Hash != firstResult.Hash {
		t.Error("Should have matching Hashes")
	}
}
