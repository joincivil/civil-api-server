package discourse

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	ceth "github.com/joincivil/go-common/pkg/eth"
	cpersist "github.com/joincivil/go-common/pkg/persistence"
	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
	ctime "github.com/joincivil/go-common/pkg/time"
)

const (
	ldmTableBaseName = "listing_discourse_map"

	// Could make this configurable later if needed
	maxOpenConns    = 5
	maxIdleConns    = 2
	connMaxLifetime = time.Nanosecond
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
		return pgPersister, fmt.Errorf("error connecting to sqlx: %v", err)
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

// RetrieveListingMap retrieves a populated  ListingMap struct
func (p *PostgresPersister) RetrieveListingMap(listingAddress string) (
	*ListingMap, error) {
	return p.listingMapFromTable(listingAddress, ldmTableBaseName)
}

// SaveListingMap saves a populated ListingMap
func (p *PostgresPersister) SaveListingMap(ldm *ListingMap) error {
	return p.saveListingMapForTable(ldm, ldmTableBaseName)
}

// CreateTables creates the tables if they don't exist
func (p *PostgresPersister) CreateTables() error {
	ldmTableQuery := CreateListingMapTableQuery(ldmTableBaseName)
	_, err := p.db.Exec(ldmTableQuery)
	if err != nil {
		return fmt.Errorf("error creating listing discourse map table: %v", err)
	}
	return nil
}

// RunMigrations runs the migration statements to update existing tables
// func (p *PostgresPersister) RunMigrations() error {
// 	return p.runJsonbMigrationsForTable(ListingMapTableBaseName)
// }

// CreateIndices creates the indices for DB if they don't exist
// func (p *PostgresPersister) CreateIndices() error {
// 	return p.createJsonbIndicesForTable(ListingMapTableBaseName)
// }

func (p *PostgresPersister) listingMapFromTable(listing string,
	tableName string) (*ListingMap, error) {
	queryStr := p.retrieveListingMapQuery(tableName)

	ldms := []*ListingMap{}

	addr := ceth.NormalizeEthAddress(listing)
	criteria := struct {
		ListingAddress string `db:"listing_address"`
	}{
		ListingAddress: addr,
	}

	nstmt, err := p.db.PrepareNamed(queryStr)
	if err != nil {
		return nil, fmt.Errorf("error preparing query with sqlx: %v", err)
	}

	err = nstmt.Select(&ldms, criteria)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user from table: %v", err)
	}

	if len(ldms) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}

	return ldms[0], nil
}

func (p *PostgresPersister) retrieveListingMapQuery(tableName string) string {
	queryBuf := bytes.NewBufferString("SELECT ")
	fieldNames, _ := cpostgres.StructFieldsForQuery(ListingMap{}, false, "")

	queryBuf.WriteString(fieldNames)                                     // nolint: gosec
	queryBuf.WriteString(" FROM ")                                       // nolint: gosec
	queryBuf.WriteString(tableName)                                      // nolint: gosec
	queryBuf.WriteString(" r1 ")                                         // nolint: gosec
	queryBuf.WriteString(" WHERE r1.listing_address = :listing_address") // nolint: gosec

	return queryBuf.String()
}

func (p *PostgresPersister) saveListingMapForTable(ldm *ListingMap,
	tableName string) error {
	queryString := cpostgres.InsertIntoDBQueryString(tableName, ListingMap{})
	queryString = p.appendOnConflictQueryClause(tableName, queryString)
	queryString = p.appendReturningQueryClause(queryString)

	ldm.ListingAddress = ceth.NormalizeEthAddress(ldm.ListingAddress)

	ts := ctime.CurrentEpochSecsInInt64()
	if ldm.CreatedTs == 0 {
		ldm.CreatedTs = ts
	}
	ldm.UpdatedTs = ts

	_, err := p.db.NamedExec(queryString, ldm)
	if err != nil {
		return fmt.Errorf("error saving ldm to table: err: %v", err)
	}
	return nil
}

func (p *PostgresPersister) appendOnConflictQueryClause(tableName string, queryString string) string {
	// Remove the ; at the end of the existing query
	splitQueryStrs := strings.Split(queryString, ";")
	queryString = splitQueryStrs[0]

	// Upsert clause to the query when there is an existing extry with the same ID
	queryBuf := bytes.NewBufferString(queryString)
	queryBuf.WriteString(" ON CONFLICT (listing_address)")
	queryBuf.WriteString(" DO UPDATE")
	queryBuf.WriteString(" SET topic_id = EXCLUDED.topic_id,")

	// Make sure the created_date stays the same but the last updated is updated
	queryBuf.WriteString(" created_ts = ")
	queryBuf.WriteString(tableName)
	queryBuf.WriteString(".created_ts, updated_ts = EXCLUDED.updated_ts;")

	return queryBuf.String()
}

func (p *PostgresPersister) appendReturningQueryClause(queryString string) string {
	// Remove the ; at the end of the existing query
	splitQueryStrs := strings.Split(queryString, ";")
	queryString = splitQueryStrs[0]

	queryBuf := bytes.NewBufferString(queryString)
	queryBuf.WriteString(" RETURNING listing_address, topic_id, created_ts, updated_ts")
	return queryBuf.String()
}

// CreateListingMapTableQuery returns the query to create the table
// for the model
func CreateListingMapTableQuery(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            listing_address TEXT PRIMARY KEY,
			topic_id INT,
			created_ts INT,
			updated_ts INT
        );
    `, tableName)
	return queryString
}
