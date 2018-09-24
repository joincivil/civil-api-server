package invoicing

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	// "strings"
	"time"
	// "database/sql"
	// "encoding/json"
	// log "github.com/golang/glog"

	"github.com/jmoiron/sqlx"
	// driver for postgresql
	_ "github.com/lib/pq"

	processorpg "github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

	crawlerpg "github.com/joincivil/civil-events-crawler/pkg/persistence/postgres"
	crawlerutils "github.com/joincivil/civil-events-crawler/pkg/utils"
)

const (
	// Could make this configurable later if needed
	maxOpenConns             = 20
	maxIdleConns             = 5
	connMaxLifetime          = time.Nanosecond
	defaultInvoicesTableName = "checkbook_invoice"

	dateUpdatedFieldName = "DateUpdated"
)

// NewPostgresPersister creates a new postgres persister instance
func NewPostgresPersister(host string, port int, user string, password string,
	dbname string) (*PostgresPersister, error) {
	pgPersister := &PostgresPersister{}
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		dbname,
	)
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return pgPersister, fmt.Errorf("Error connecting to sqlx: %v", err)
	}
	pgPersister.db = db
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	return pgPersister, nil
}

// PostgresPersister implements the persister for Postgresql
type PostgresPersister struct {
	db *sqlx.DB
}

// Invoices retrieves invoices from Postgres
func (p *PostgresPersister) Invoices(id string, email string, status string,
	checkID string) ([]*PostgresInvoice, error) {
	return p.invoicesFromTable(id, email, status, checkID, defaultInvoicesTableName)
}

// SaveInvoice saves an invoice object or an error.
func (p *PostgresPersister) SaveInvoice(invoice *PostgresInvoice) error {
	return p.createInvoiceForTable(invoice, defaultInvoicesTableName)
}

// UpdateInvoice updates an invoice object or an error.
func (p *PostgresPersister) UpdateInvoice(invoice *PostgresInvoice, updatedFields []string) error {
	return p.updateInvoiceForTable(invoice, updatedFields, defaultInvoicesTableName)
}

// CreateTables creates the tables if they don't exist
func (p *PostgresPersister) CreateTables() error {
	invoiceTableQuery := CreateInvoiceTableQuery(defaultInvoicesTableName)
	_, err := p.db.Exec(invoiceTableQuery)
	if err != nil {
		return fmt.Errorf("Error creating checkbook_invoice table in postgres: %v", err)
	}
	return nil
}

// CreateIndices creates the indices for DB if they don't exist
func (p *PostgresPersister) CreateIndices() error {
	return p.createInvoiceIndicesForTable(defaultInvoicesTableName)
}

func (p *PostgresPersister) invoicesFromTable(id string, email string, status string,
	checkID string, tableName string) ([]*PostgresInvoice, error) {
	invoices := []*PostgresInvoice{}
	queryString := p.invoicesQuery(id, email, status, checkID, tableName)
	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error preparing query with sqlx: %v", err)
	}

	queryParams := map[string]interface{}{
		"id":       id,
		"email":    email,
		"status":   status,
		"check_id": checkID,
	}
	err = nstmt.Select(&invoices, queryParams)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving invoices from table: %v", err)
	}
	return invoices, err
}

func (p *PostgresPersister) invoicesQuery(id string, email string, status string,
	checkID string, tableName string) string {
	queryBuf := bytes.NewBufferString("SELECT ")
	fieldNames, _ := crawlerpg.StructFieldsForQuery(PostgresInvoice{}, false)
	queryBuf.WriteString(fieldNames) // nolint: gosec
	queryBuf.WriteString(" FROM ")   // nolint: gosec
	queryBuf.WriteString(tableName)  // nolint: gosec
	queryBuf.WriteString(" r1 ")     // nolint: gosec

	if email != "" {
		queryBuf.WriteString(" WHERE r1.email = :email") // nolint: gosec
	} else if id != "" {
		queryBuf.WriteString(" WHERE r1.invoice_id = :id") // nolint: gosec
	} else if checkID != "" {
		queryBuf.WriteString(" WHERE r1.check_id = :check_id") // nolint: gosec
	} else if status != "" {
		queryBuf.WriteString(" WHERE r1.invoice_status = :status") // nolint: gosec
	}
	return queryBuf.String()
}

func (p *PostgresPersister) createInvoiceForTable(invoice *PostgresInvoice, tableName string) error {
	ts := crawlerutils.CurrentEpochSecsInInt64()
	invoice.DateCreated = ts
	invoice.DateUpdated = ts
	err := invoice.GenerateHash()
	if err != nil {
		return fmt.Errorf("Error generating hash for invoice: err: %v", err)
	}

	queryString := crawlerpg.InsertIntoDBQueryString(tableName, PostgresInvoice{})
	_, err = p.db.NamedExec(queryString, invoice)
	if err != nil {
		return fmt.Errorf("Error saving invoice to table: err: %v", err)
	}
	return nil
}

func (p *PostgresPersister) updateInvoiceForTable(invoice *PostgresInvoice, updatedFields []string,
	tableName string) error {
	ts := crawlerutils.CurrentEpochSecsInInt64()
	invoice.DateUpdated = ts
	updatedFields = append(updatedFields, dateUpdatedFieldName)

	queryString, err := p.updateInvoiceQuery(updatedFields, tableName)
	if err != nil {
		return fmt.Errorf("Error creating query string for update: %v ", err)
	}
	_, err = p.db.NamedExec(queryString, invoice)
	if err != nil {
		return fmt.Errorf("Error updating fields in db: %v", err)
	}
	return nil
}

func (p *PostgresPersister) updateInvoiceQuery(updatedFields []string, tableName string) (string, error) {
	queryString, err := p.updateDBQueryBuffer(updatedFields, tableName, PostgresInvoice{})
	if err != nil {
		return "", err
	}
	queryString.WriteString(" WHERE hash=:hash;") // nolint: gosec
	return queryString.String(), nil
}

// TODO(PN): Copied from processor code.  All copies should live in a common place,
// perhaps up to the crawler code (ideally a common library)
func (p *PostgresPersister) updateDBQueryBuffer(updatedFields []string, tableName string,
	dbModelStruct interface{}) (bytes.Buffer, error) {
	var queryBuf bytes.Buffer
	queryBuf.WriteString("UPDATE ") // nolint: gosec
	queryBuf.WriteString(tableName) // nolint: gosec
	queryBuf.WriteString(" SET ")   // nolint: gosec
	for idx, field := range updatedFields {
		dbFieldName, err := processorpg.DbFieldNameFromModelName(dbModelStruct, field)
		if err != nil {
			return queryBuf, fmt.Errorf("Error getting %s from %s table DB struct tag: %v", field, tableName, err)
		}
		queryBuf.WriteString(fmt.Sprintf("%s=:%s", dbFieldName, dbFieldName)) // nolint: gosec
		if idx+1 < len(updatedFields) {
			queryBuf.WriteString(", ") // nolint: gosec
		}
	}
	return queryBuf, nil
}

func (p *PostgresPersister) createInvoiceIndicesForTable(tableName string) error {
	indexQuery := CreateInvoiceTableIndicesString(tableName)
	_, err := p.db.Exec(indexQuery)
	return err
}

// CreateInvoiceTableQuery returns the query to create this table
func CreateInvoiceTableQuery(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            hash TEXT PRIMARY KEY,
            email TEXT,
            name TEXT,
            amount DECIMAL,
            invoice_id TEXT UNIQUE,
            invoice_num TEXT,
            invoice_status TEXT,
            check_id TEXT UNIQUE,
            check_status TEXT,
            date_created BIGINT,
            date_updated BIGINT,
            stop_poll bool
        );
    `, tableName)
	return queryString
}

// CreateInvoiceTableIndicesString returns the query to create indices on the
// JSONb table
func CreateInvoiceTableIndicesString(tableName string) string {
	queryString := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS email_idx ON %s (email);
		CREATE INDEX IF NOT EXISTS invoice_status_idx ON %s (invoice_status);
	`, tableName, tableName)
	return queryString
}

// PostgresInvoice represents the invoice data in Postgres
type PostgresInvoice struct {
	Hash          string  `db:"hash"` // key off this
	Email         string  `db:"email"`
	Name          string  `db:"name"`
	Amount        float64 `db:"amount"`
	InvoiceID     string  `db:"invoice_id"`
	InvoiceNum    string  `db:"invoice_num"`
	InvoiceStatus string  `db:"invoice_status"`
	CheckID       string  `db:"check_id"`
	CheckStatus   string  `db:"check_status"`
	DateCreated   int64   `db:"date_created"`
	DateUpdated   int64   `db:"date_updated"`
	StopPoll      bool    `db:"stop_poll"`
}

// GenerateHash sets the Hash field with a hash of email, amount, and date created.
func (p *PostgresInvoice) GenerateHash() error {
	strToHash := fmt.Sprintf("%v|%v|%v", p.Email, p.Amount, p.DateCreated)
	h := sha256.New()
	_, err := h.Write([]byte(strToHash))
	if err != nil {
		return err
	}
	p.Hash = hex.EncodeToString(h.Sum(nil))
	return nil
}
