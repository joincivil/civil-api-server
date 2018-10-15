// +build integration

package users

import (
	"fmt"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	tpostgresPort   = 5432
	tpostgresDBName = "civil_crawler"
	tpostgresUser   = "docker"
	tpostgresPswd   = "docker"
	tpostgresHost   = "localhost"

	defaultKycUserTestTableName = "civil_user_test"
)

func setupDBConnection() (*PostgresPersister, error) {
	return NewPostgresPersister(tpostgresHost, tpostgresPort, tpostgresUser,
		tpostgresPswd, tpostgresDBName)
}

func setupTestTable(tableName string) (*PostgresPersister, error) {
	persister, err := setupDBConnection()
	if err != nil {
		return persister, fmt.Errorf("Error connecting to DB: %v", err)
	}
	var queryString string
	switch tableName {
	case defaultKycUserTestTableName:
		queryString = CreateKycUserTableQuery(tableName)
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
	case defaultKycUserTestTableName:
		_, err = persister.db.Query(fmt.Sprintf("DROP TABLE %v;", tableName))
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
	err = checkTableExists(defaultKycUserTableName, persister)
	if err != nil {
		t.Error(err)
	}
}

// TestTableSetup tests to ensure that our DB tables are being setup
func TestIndicesSetup(t *testing.T) {
	persister, err := setupTestTable(defaultKycUserTestTableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, defaultKycUserTestTableName)
	err = persister.createKycUserIndicesForTable(defaultKycUserTestTableName)
	if err != nil {
		t.Errorf("Error creating indices for invoices: %v", err)
	}
}

func TestGetKycUser(t *testing.T) {
	persister, err := setupTestTable(defaultKycUserTestTableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, defaultKycUserTestTableName)

	testKycUserToCreate := &User{
		Email:      "test@civil.co",
		EthAddress: "testEthAddress1",
	}

	err = persister.createKycUserForTable(testKycUserToCreate, defaultKycUserTestTableName)
	if err != nil {
		t.Fatalf("Should not have received an error saving user: err: %v", err)
	}

	criteria := &UserCriteria{
		Email: testKycUserToCreate.Email,
	}
	user, err := persister.userFromTable(criteria, defaultKycUserTestTableName)
	if err != nil {
		t.Fatalf("Should have received a result from users: err: %v", err)
	}

	if user.Email != testKycUserToCreate.Email {
		t.Error("Should have gotten the correct entry")
	}

	if user.UID == "" {
		t.Error("Should have set and stored UID")
	}

	// Validate the UUID stored
	_, err = uuid.FromString(user.UID)
	if err != nil {
		t.Errorf("Should have gotten back a valid UUID for UID: err: %v", err)
	}
}

func TestUpdateKycUser(t *testing.T) {
	persister, err := setupTestTable(defaultKycUserTestTableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, defaultKycUserTestTableName)

	testKycUserToCreate := &User{
		Email:      "test@civil.co",
		EthAddress: "testEthAddress1",
	}

	err = persister.createKycUserForTable(testKycUserToCreate, defaultKycUserTestTableName)
	if err != nil {
		t.Fatalf("Should not have received an error saving user: err: %v", err)
	}

	criteria := &UserCriteria{
		Email: testKycUserToCreate.Email,
	}
	user, err := persister.userFromTable(criteria, defaultKycUserTestTableName)
	if err != nil {
		t.Fatalf("Should have received a result from users: err: %v", err)
	}

	updatedFields := []string{}
	lastUpdate := user.DateUpdated
	user.QuizStatus = "updatedstatus"
	updatedFields = append(updatedFields, "QuizStatus")
	user.KycStatus = "done"
	updatedFields = append(updatedFields, "KycStatus")

	// To test the date updated
	time.Sleep(1 * time.Second)

	err = persister.updateKycUserForTable(user, updatedFields, defaultKycUserTestTableName)
	if err != nil {
		t.Fatalf("Should not have received an error updating user: err: %v", err)
	}

	user, err = persister.userFromTable(criteria, defaultKycUserTestTableName)
	if err != nil {
		t.Fatalf("Should have received a result from users: err: %v", err)
	}

	if user.QuizStatus != "updatedstatus" {
		t.Error("Should have updated the quiz status")
	}
	if user.KycStatus != "done" {
		t.Error("Should have updated the kyc status")
	}
	if user.DateUpdated <= lastUpdate {
		t.Error("Should have updated date_updated")
	}
}
