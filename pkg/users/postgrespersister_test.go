// +build integration

package users

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/joincivil/civil-api-server/pkg/testutils"
	uuid "github.com/satori/go.uuid"
)

const (
	defaultUserTestTableName = "civil_user_test"
)

func setupDBConnection() (*PostgresPersister, error) {
	creds := testutils.GetTestDBCreds()
	return NewPostgresPersister(creds.Host, creds.Port, creds.User, creds.Password, creds.Dbname)
}

func setupTestTable(tableName string) (*PostgresPersister, error) {
	persister, err := setupDBConnection()
	if err != nil {
		return persister, fmt.Errorf("Error connecting to DB: %v", err)
	}
	var queryString string
	switch tableName {
	case defaultUserTestTableName:
		queryString = CreateUserTableQuery(tableName)
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
	case defaultUserTestTableName:
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
	err = persister.RunMigrations()
	if err != nil {
		t.Errorf("Error running migrations: %v", err)
	}
	err = checkTableExists(defaultUserTableName, persister)
	if err != nil {
		t.Error(err)
	}
}

// TestTableSetup tests to ensure that our DB tables are being setup
func TestIndicesSetup(t *testing.T) {
	persister, err := setupTestTable(defaultUserTestTableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, defaultUserTestTableName)
	err = persister.createUserIndicesForTable(defaultUserTestTableName)
	if err != nil {
		t.Errorf("Error creating indices for invoices: %v", err)
	}
}

func TestGetUser(t *testing.T) {
	persister, err := setupTestTable(defaultUserTestTableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, defaultUserTestTableName)

	testUserToCreate := &User{
		Email:      "test@civil.co",
		EthAddress: "testEthAddress1",
		PurchaseTxHashes: []string{
			"0xf8732e94ee5e1da45489566ee29d0954bf51cca77f22bdacde01a1b1a09a61a9",
			"0x6f73ffd316cfe4b7e9db1ca50bd9ea316cffc60c87d240eee9d9693e26d0381b",
		},
		AssocNewsoomAddr: []string{
			"0x01DEae0e3f07bbF8D979e751bc160bf11Ec58511",
			"0x16af58Ce3e58e230C92d969752a3b89a92064d3E",
		},
	}

	_, err = persister.createUserForTable(testUserToCreate, defaultUserTestTableName)
	if err != nil {
		t.Fatalf("Should not have received an error saving user: err: %v", err)
	}

	// Test the get by email query
	criteria := &UserCriteria{
		Email: testUserToCreate.Email,
	}
	user, err := persister.userFromTable(criteria, defaultUserTestTableName)
	if err != nil {
		t.Fatalf("Should have received a result from users: err: %v", err)
	}

	if user.Email != testUserToCreate.Email {
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

	// Test the onfido check id query
	criteria = &UserCriteria{
		Email: testUserToCreate.Email,
	}
	user, err = persister.userFromTable(criteria, defaultUserTestTableName)
	if err != nil {
		t.Fatalf("Should have received a result from users given check id: err: %v", err)
	}

	if user.Email != testUserToCreate.Email {
		t.Error("Should have gotten the correct email")
	}

	if user.EthAddress != testUserToCreate.EthAddress {
		t.Error("Should have gotten the correct eth address")
	}

	if len(user.PurchaseTxHashes) != 2 {
		t.Error("Should have gotten two results in purchase txhash")
	} else {
		if strings.ToLower(user.PurchaseTxHashes[0]) != strings.ToLower("0xf8732e94ee5e1da45489566ee29d0954bf51cca77f22bdacde01a1b1a09a61a9") {
			t.Error("Should have gotten correct hash")
		}
		if strings.ToLower(user.PurchaseTxHashes[1]) != strings.ToLower("0x6f73ffd316cfe4b7e9db1ca50bd9ea316cffc60c87d240eee9d9693e26d0381b") {
			t.Error("Should have gotten correct hash")
		}
	}

	if len(user.AssocNewsoomAddr) != 2 {
		t.Error("Should have gotten two results in associated newsroom addresses")
	} else {
		if strings.ToLower(user.AssocNewsoomAddr[0]) != strings.ToLower("0x01DEae0e3f07bbF8D979e751bc160bf11Ec58511") {
			t.Error("Should have gotten correct address")
		}
		if strings.ToLower(user.AssocNewsoomAddr[1]) != strings.ToLower("0x16af58Ce3e58e230C92d969752a3b89a92064d3E") {
			t.Error("Should have gotten correct address")
		}
	}

}

func TestUpdateUser(t *testing.T) {
	persister, err := setupTestTable(defaultUserTestTableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, defaultUserTestTableName)

	testUserToCreate := &User{
		Email:      "test@civil.co",
		EthAddress: "testEthAddress1",
	}

	_, err = persister.createUserForTable(testUserToCreate, defaultUserTestTableName)
	if err != nil {
		t.Fatalf("Should not have received an error saving user: err: %v", err)
	}

	criteria := &UserCriteria{
		Email: testUserToCreate.Email,
	}
	user, err := persister.userFromTable(criteria, defaultUserTestTableName)
	if err != nil {
		t.Fatalf("Should have received a result from users: err: %v", err)
	}

	updatedFields := []string{}
	lastUpdate := user.DateUpdated
	user.QuizStatus = "updatedstatus"
	updatedFields = append(updatedFields, "QuizStatus")

	// To test the date updated
	time.Sleep(1 * time.Second)

	err = persister.updateUserForTable(user, updatedFields, defaultUserTestTableName)
	if err != nil {
		t.Fatalf("Should not have received an error updating user: err: %v", err)
	}

	user, err = persister.userFromTable(criteria, defaultUserTestTableName)
	if err != nil {
		t.Fatalf("Should have received a result from users: err: %v", err)
	}

	if user.QuizStatus != "updatedstatus" {
		t.Error("Should have updated the quiz status")
	}
	if user.DateUpdated <= lastUpdate {
		t.Error("Should have updated date_updated")
	}
}
