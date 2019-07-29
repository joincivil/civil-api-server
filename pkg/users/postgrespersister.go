package users

import (
	"bytes"
	// "crypto/sha256"
	// "encoding/hex"

	"fmt"
	"time"

	"github.com/jinzhu/gorm"
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
	maxOpenConns         = 20
	maxIdleConns         = 5
	connMaxLifetime      = time.Nanosecond
	defaultUserTableName = "civil_user"
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

// NewPersisterFromGorm creates a persister using an existing gorm connection
func NewPersisterFromGorm(gormDB *gorm.DB) *PostgresPersister {
	db := sqlx.NewDb(gormDB.DB(), "postgres")

	return &PostgresPersister{db}
}

// Users retrieves a list of users based on the given UserCriteria
func (p *PostgresPersister) Users(criteria *UserCriteria) ([]*User, error) {
	return p.usersFromTable(criteria, defaultUserTableName)
}

// User retrieves a user based on the given UserCriteria
func (p *PostgresPersister) User(criteria *UserCriteria) (*User, error) {
	return p.userFromTable(criteria, defaultUserTableName)
}

// SaveUser saves a new user
func (p *PostgresPersister) SaveUser(user *User) (*User, error) {
	return p.createUserForTable(user, defaultUserTableName)
}

// UpdateUser updates an existing user
func (p *PostgresPersister) UpdateUser(user *User, updatedFields []string) error {
	return p.updateUserForTable(user, updatedFields, defaultUserTableName)
}

// CreateTables creates the tables if they don't exist
func (p *PostgresPersister) CreateTables() error {
	userTableQuery := CreateUserTableQuery(defaultUserTableName)
	_, err := p.db.Exec(userTableQuery)
	if err != nil {
		return fmt.Errorf("Error creating user table in postgres: %v", err)
	}
	return nil
}

// RunMigrations runs the migrations statements to update existing tables.
func (p *PostgresPersister) RunMigrations() error {
	return p.runUserMigrationsForTable(defaultUserTableName)
}

// CreateIndices creates the indices for DB if they don't exist
func (p *PostgresPersister) CreateIndices() error {
	return p.createUserIndicesForTable(defaultUserTableName)
}

func (p *PostgresPersister) usersFromTable(criteria *UserCriteria, tableName string) ([]*User, error) {
	users := []*User{}
	queryString := p.userQuery(criteria, tableName)
	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error preparing query with sqlx: %v", err)

	}
	err = nstmt.Select(&users, criteria)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving user from table: %v", err)
	}
	if len(users) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}
	return users, nil
}

func (p *PostgresPersister) userFromTable(criteria *UserCriteria, tableName string) (*User, error) {
	users, err := p.usersFromTable(criteria, tableName)
	if err != nil {
		return nil, err
	}
	return users[0], err
}

func (p *PostgresPersister) userQuery(criteria *UserCriteria, tableName string) string {
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
	} else if criteria.NewsroomAddr != "" {
		queryBuf.WriteString(" WHERE r1.assoc_nr_addr @> ARRAY[:nr_addr]") // nolint: gosec
	}
	return queryBuf.String()
}

func (p *PostgresPersister) createUserForTable(user *User, tableName string) (*User, error) {
	// Ensure a UID is generated. Will fail if a UID already exists, which means
	// the user was already created
	err := user.GenerateUID()
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("Error saving user to table: err: %v", err)
	}
	return user, nil
}

func (p *PostgresPersister) updateUserForTable(user *User, updatedFields []string,
	tableName string) error {
	ts := ctime.CurrentEpochSecsInInt64()
	user.DateUpdated = ts
	updatedFields = append(updatedFields, dateUpdatedFieldName)

	queryString, err := p.updateUserQuery(updatedFields, tableName)
	if err != nil {
		return fmt.Errorf("Error creating query string for update: %v ", err)
	}
	_, err = p.db.NamedExec(queryString, user)
	if err != nil {
		return fmt.Errorf("Error updating fields in db: %v", err)
	}
	return nil
}

func (p *PostgresPersister) updateUserQuery(updatedFields []string, tableName string) (string, error) {
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

func (p *PostgresPersister) runUserMigrationsForTable(tableName string) error {
	indexQuery := CreateUserTableMigrationQuery(tableName)
	_, err := p.db.Exec(indexQuery)
	return err
}

func (p *PostgresPersister) createUserIndicesForTable(tableName string) error {
	indexQuery := CreateUserTableIndicesString(tableName)
	_, err := p.db.Exec(indexQuery)
	return err
}

// CreateUserTableQuery returns the query to create the  users
func CreateUserTableQuery(tableName string) string {
	// XXX(PN): ideally, uuid should be generated by the DB, but not messing with DB
	// permissions right now.  Going to generate it via application code.
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
			uid uuid PRIMARY KEY,
			email TEXT DEFAULT '',
			eth_address TEXT DEFAULT '',
			quiz_payload JSONB DEFAULT '{}',
			quiz_status TEXT DEFAULT '',
			date_created INT DEFAULT 0,
			date_updated INT DEFAULT 0,
			purchase_txhashes TEXT[] DEFAULT ARRAY[]::TEXT[],
			civilian_whitelist_tx_id TEXT DEFAULT '',
			app_refer TEXT DEFAULT '',
			nr_step INT DEFAULT 0,
			nr_far_step INT DEFAULT 0,
			nr_last_seen INT DEFAULT 0,
			assoc_nr_addr TEXT[] DEFAULT ARRAY[]::TEXT[]
        );
    `, tableName)
	return queryString
}

// CreateUserTableMigrationQuery calls alter table queries to update existing tables if
// they exists in lieu of a migration mechanism for our DB. Ensure we always
// set the default values of any altered columns.
func CreateUserTableMigrationQuery(tableName string) string {
	// ALTER TABLE %s ALTER COLUMN purchase_txhashes TYPE TEXT[] USING ARRAY[purchase_txhashes];
	// ALTER TABLE %s ALTER COLUMN purchase_txhashes SET DEFAULT ARRAY[]::TEXT[];
	// ALTER TABLE %s DROP COLUMN IF EXISTS newsroom_data;
	queryString := fmt.Sprintf(`
		ALTER TABLE %s ADD COLUMN IF NOT EXISTS app_refer TEXT DEFAULT '';
		ALTER TABLE %s ADD COLUMN IF NOT EXISTS nr_step INT DEFAULT 0;
		ALTER TABLE %s ADD COLUMN IF NOT EXISTS nr_far_step INT DEFAULT 0;
		ALTER TABLE %s ADD COLUMN IF NOT EXISTS nr_last_seen INT DEFAULT 0;
		ALTER TABLE %s ADD COLUMN IF NOT EXISTS assoc_nr_addr TEXT[] DEFAULT ARRAY[]::TEXT[];
		ALTER TABLE %s DROP COLUMN IF EXISTS onfido_applicant_id;
		ALTER TABLE %s DROP COLUMN IF EXISTS onfido_check_id;
		ALTER TABLE %s DROP COLUMN IF EXISTS kyc_status;
	`, tableName, tableName, tableName, tableName, tableName, tableName,
		tableName, tableName)
	return queryString
}

// CreateUserTableIndicesString returns the query to create indices on the table
func CreateUserTableIndicesString(tableName string) string {
	queryString := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS email_idx ON %s (email);
		CREATE INDEX IF NOT EXISTS eth_address_idx ON %s (eth_address);
		CREATE INDEX IF NOT EXISTS assoc_nr_addr_idx ON %s USING GIN (assoc_nr_addr);
	`, tableName, tableName, tableName)
	return queryString
}
