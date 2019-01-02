package users

import (
	"bytes"
	// "crypto/sha256"
	// "encoding/hex"

	"fmt"
	"time"

	// "database/sql"
	// log "github.com/golang/glog"

	"github.com/jmoiron/sqlx"
	// driver for postgresql
	_ "github.com/lib/pq"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
	ctime "github.com/joincivil/go-common/pkg/time"
)

const (
	// Could make this configurable later if needed
	maxOpenConns            = 20
	maxIdleConns            = 5
	connMaxLifetime         = time.Nanosecond
	defaultKycUserTableName = "civil_user"
	dateUpdatedFieldName    = "DateUpdated"
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

// User retrieves a user based on the given UserCriteria
func (p *PostgresPersister) User(criteria *UserCriteria) (*User, error) {
	return p.userFromTable(criteria, defaultKycUserTableName)
}

// SaveUser saves a new user
func (p *PostgresPersister) SaveUser(user *User) error {
	return p.createKycUserForTable(user, defaultKycUserTableName)
}

// UpdateUser updates an existing user
func (p *PostgresPersister) UpdateUser(user *User, updatedFields []string) error {
	return p.updateKycUserForTable(user, updatedFields, defaultKycUserTableName)
}

// CreateTables creates the tables if they don't exist
func (p *PostgresPersister) CreateTables() error {
	kycUserTableQuery := CreateKycUserTableQuery(defaultKycUserTableName)
	_, err := p.db.Exec(kycUserTableQuery)
	if err != nil {
		return fmt.Errorf("Error creating user table in postgres: %v", err)
	}
	return nil
}

// CreateIndices creates the indices for DB if they don't exist
func (p *PostgresPersister) CreateIndices() error {
	return p.createKycUserIndicesForTable(defaultKycUserTableName)
}

func (p *PostgresPersister) userFromTable(criteria *UserCriteria, tableName string) (*User, error) {
	kycUsers := []*User{}
	queryString := p.kycUserQuery(criteria, tableName)
	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error preparing query with sqlx: %v", err)

	}
	err = nstmt.Select(&kycUsers, criteria)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving user from table: %v", err)
	}
	if len(kycUsers) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}
	return kycUsers[0], err
}

func (p *PostgresPersister) kycUserQuery(criteria *UserCriteria, tableName string) string {
	queryBuf := bytes.NewBufferString("SELECT ")
	fieldNames, _ := cpostgres.StructFieldsForQuery(User{}, false, "")

	queryBuf.WriteString(fieldNames) // nolint: gosec
	queryBuf.WriteString(" FROM ")   // nolint: gosec
	queryBuf.WriteString(tableName)  // nolint: gosec
	queryBuf.WriteString(" r1 ")     // nolint: gosec

	if criteria.UID != "" {
		queryBuf.WriteString(" WHERE r1.uid = :uid") // nolint: gosec
	} else if criteria.Email != "" {
		queryBuf.WriteString(" WHERE r1.email = :email") // nolint: gosec
	} else if criteria.EthAddress != "" {
		queryBuf.WriteString(" WHERE r1.eth_address = :eth_address") // nolint: gosec
	} else if criteria.OnfidoCheckID != "" {
		queryBuf.WriteString(" WHERE r1.onfido_check_id = :onfido_check_id") // nolint: gosec
	}
	return queryBuf.String()
}

func (p *PostgresPersister) createKycUserForTable(user *User, tableName string) error {
	// Ensure a UID is generated. Will fail if a UID already exists, which means
	// the user was already created
	err := user.GenerateUID()
	if err != nil {
		return err
	}
	// Create a default empty payload if none found
	if user.QuizPayload == nil {
		user.QuizPayload = cpostgres.JsonbPayload{}
	}
	ts := ctime.CurrentEpochSecsInInt64()
	user.DateCreated = ts
	user.DateUpdated = ts

	queryString := cpostgres.InsertIntoDBQueryString(tableName, User{})
	_, err = p.db.NamedExec(queryString, user)
	if err != nil {
		return fmt.Errorf("Error saving user to table: err: %v", err)
	}
	return nil
}

func (p *PostgresPersister) updateKycUserForTable(user *User, updatedFields []string,
	tableName string) error {
	ts := ctime.CurrentEpochSecsInInt64()
	user.DateUpdated = ts
	updatedFields = append(updatedFields, dateUpdatedFieldName)

	queryString, err := p.updateKycUserQuery(updatedFields, tableName)
	if err != nil {
		return fmt.Errorf("Error creating query string for update: %v ", err)
	}
	_, err = p.db.NamedExec(queryString, user)
	if err != nil {
		return fmt.Errorf("Error updating fields in db: %v", err)
	}
	return nil
}

func (p *PostgresPersister) updateKycUserQuery(updatedFields []string, tableName string) (string, error) {
	queryString, err := p.updateDBQueryBuffer(updatedFields, tableName, User{})
	if err != nil {
		return "", err
	}
	queryString.WriteString(" WHERE uid = :uid") // nolint: gosec
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
		dbFieldName, err := cpostgres.DbFieldNameFromModelName(dbModelStruct, field)
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

func (p *PostgresPersister) createKycUserIndicesForTable(tableName string) error {
	indexQuery := CreateKycUserTableIndicesString(tableName)
	_, err := p.db.Exec(indexQuery)
	return err
}

// CreateKycUserTableQuery returns the query to create the KYC users
func CreateKycUserTableQuery(tableName string) string {
	// XXX(PN): ideally, uuid should be generated by the DB, but not messing with DB
	// permissions right now.  Going to generate it via application code.
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
			uid uuid PRIMARY KEY,
			email TEXT,
			eth_address TEXT,
			onfido_applicant_id TEXT,
			onfido_check_id TEXT,
			kyc_status TEXT,
			quiz_payload JSONB,
			quiz_status TEXT,
			newsroom_data JSONB,
			date_created INT,
			date_updated INT
        );
    `, tableName)
	return queryString
}

// CreateKycUserTableIndicesString returns the query to create indices on the table
func CreateKycUserTableIndicesString(tableName string) string {
	queryString := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS email_idx ON %s (email);
		CREATE INDEX IF NOT EXISTS eth_address_idx ON %s (eth_address);
	`, tableName, tableName)
	return queryString
}
