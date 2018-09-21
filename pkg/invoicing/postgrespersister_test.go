// +build integration

package invoicing

import (
	"fmt"
	"testing"
	"time"
)

const (
	tpostgresPort   = 5432
	tpostgresDBName = "civil_crawler"
	tpostgresUser   = "docker"
	tpostgresPswd   = "docker"
	tpostgresHost   = "localhost"

	defaultInvoicesTestTableName = "checkbook_invoice_test"
)

var (
	testInvoiceToCreate = &PostgresInvoice{
		Email:    "test@civil.co",
		Name:     "First Last",
		Amount:   100.32,
		StopPoll: false,
	}
	testInvoiceToUpdate = &PostgresInvoice{
		Email:    "test@civil.co",
		Name:     "First Last",
		Amount:   100.32,
		StopPoll: false,
	}
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
	case defaultInvoicesTestTableName:
		queryString = CreateInvoiceTableQuery(tableName)
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
	case defaultInvoicesTestTableName:
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
	err = checkTableExists(defaultInvoicesTableName, persister)
	if err != nil {
		t.Error(err)
	}
}

// TestTableSetup tests to ensure that our DB tables are being setup
func TestIndicesSetup(t *testing.T) {
	persister, err := setupTestTable(defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, defaultInvoicesTestTableName)
	err = persister.createInvoiceIndicesForTable(defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Error creating indices for invoices: %v", err)
	}
}

func TestGetInvoices(t *testing.T) {
	persister, err := setupTestTable(defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, defaultInvoicesTestTableName)
	err = persister.createInvoiceForTable(testInvoiceToCreate, defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Should not have received an error saving invoice: err: %v", err)
	}

	invoices, err := persister.invoicesFromTable("", testInvoiceToCreate.Email, "", "",
		defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Should have received a result from invoices: err: %v", err)
	}
	if len(invoices) != 1 {
		t.Errorf("Should have received seen one item in invoices: len: %v", len(invoices))
	}
	invoice := invoices[0]
	if invoice.Email != testInvoiceToCreate.Email {
		t.Error("Should have gotten the correct results")
	}

	_, err = persister.invoicesFromTable(invoice.InvoiceID, "", "", "",
		defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Should have received a result from invoices: err: %v", err)
	}

	_, err = persister.invoicesFromTable("", "", invoice.InvoiceStatus, "",
		defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Should have received a result from invoices: err: %v", err)
	}
}

func TestSaveInvoice(t *testing.T) {
	persister, err := setupTestTable(defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, defaultInvoicesTestTableName)
	err = persister.createInvoiceForTable(testInvoiceToCreate, defaultInvoicesTestTableName)
	if err != nil {
		t.Error("Should not have received an error saving invoice")
	}

	var numRowsb int
	err = persister.db.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM %v", defaultInvoicesTestTableName),
	).Scan(&numRowsb)
	if numRowsb != 1 {
		t.Errorf("Number of rows in table should be 0 but is: %v", numRowsb)
	}

	invoices, err := persister.invoicesFromTable("", testInvoiceToCreate.Email, "", "",
		defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Should have received a result from invoices: err: %v", err)
	}
	if len(invoices) != 1 {
		t.Errorf("Should have received seen one item in invoices: len: %v", len(invoices))
	}
	invoice := invoices[0]
	if invoice.DateCreated == 0 {
		t.Error("Should set the date created")
	}
	if invoice.DateUpdated == 0 {
		t.Error("Should set the date updated")
	}
}

func TestUpdateInvoice(t *testing.T) {
	persister, err := setupTestTable(defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister, defaultInvoicesTestTableName)
	err = persister.createInvoiceForTable(testInvoiceToUpdate, defaultInvoicesTestTableName)
	if err != nil {
		t.Error("Should not have received an error saving invoice")
	}

	var numRowsb int
	err = persister.db.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM %v", defaultInvoicesTestTableName),
	).Scan(&numRowsb)
	if numRowsb != 1 {
		t.Errorf("Number of rows in table should be 0 but is: %v", numRowsb)
	}

	testInvoiceToUpdate.InvoiceID = "invoiceid1"
	testInvoiceToUpdate.InvoiceNum = "1000"
	testInvoiceToUpdate.InvoiceStatus = InvoiceStatusUnpaid
	updatedFields := []string{"InvoiceID", "InvoiceNum", "InvoiceStatus"}
	time.Sleep(1 * time.Second)
	err = persister.updateInvoiceForTable(testInvoiceToUpdate, updatedFields, defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Should not have received an error updating invoice: err: %v", err)
	}

	invoices, err := persister.invoicesFromTable("", testInvoiceToCreate.Email, "", "",
		defaultInvoicesTestTableName)
	if err != nil {
		t.Errorf("Should have received a result from invoices: err: %v", err)
	}
	invoice := invoices[0]
	if invoice.DateCreated >= invoice.DateUpdated {
		t.Error("Should have updated the date updated on the update")
	}
}
