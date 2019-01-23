package jsonstore

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	// "database/sql"
	"encoding/json"

	log "github.com/golang/glog"

	"github.com/jmoiron/sqlx"
	// driver for postgresql
	_ "github.com/lib/pq"

	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
	ctime "github.com/joincivil/go-common/pkg/time"
)

const (
	// Could make this configurable later if needed
	maxOpenConns    = 20
	maxIdleConns    = 5
	connMaxLifetime = time.Nanosecond

	defaultJsonbTableName = "jsonb"
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

// RetrieveJsonb retrieves a populated Jsonb object or
// an error.  If no results, returns ErrNoJsonbFound.
func (p *PostgresPersister) RetrieveJsonb(id string, hash string) ([]*JSONb, error) {
	return p.jsonbFromTable(id, hash, defaultJsonbTableName)
}

// SaveJsonb saves a populated Jsonb object or an error
func (p *PostgresPersister) SaveJsonb(jsonb *JSONb) (*JSONb, error) {
	return p.saveJsonbForTable(jsonb, defaultJsonbTableName)
}

// CreateTables creates the tables if they don't exist
func (p *PostgresPersister) CreateTables() error {
	jsonbTableQuery := CreateJsonbTableQuery(defaultJsonbTableName)
	_, err := p.db.Exec(jsonbTableQuery)
	if err != nil {
		return fmt.Errorf("Error creating jsonb table in postgres: %v", err)
	}
	return nil
}

// CreateIndices creates the indices for DB if they don't exist
func (p *PostgresPersister) CreateIndices() error {
	return p.createJsonbIndicesForTable(defaultJsonbTableName)
}

func (p *PostgresPersister) jsonbFromTable(key string, hash string, tableName string) (
	[]*JSONb, error) {
	queryString := retrieveJsonbQuery(tableName, key, hash)
	jsonbs := []*PostgresJSONb{}
	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error preparing query with sqlx: %v", err)
	}
	criteria := map[string]interface{}{
		"key":  key,
		"hash": hash,
	}
	err = nstmt.Select(&jsonbs, criteria)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving jsonb from table: %v", err)
	}
	jsbs := make([]*JSONb, len(jsonbs))
	for index, jsonb := range jsonbs {
		jsb, err := jsonb.PostgresJSONbToJSONb()
		if err != nil {
			log.Errorf("Error converting db to jsonb data: err: %v", err)
			continue
		}
		jsbs[index] = jsb
	}
	if len(jsbs) == 0 {
		return nil, ErrNoJsonbFound
	}
	return jsbs, nil

}

func (p *PostgresPersister) saveJsonbForTable(jsonb *JSONb, tableName string) (*JSONb, error) {
	queryString := cpostgres.InsertIntoDBQueryString(tableName, PostgresJSONb{})
	queryString = p.appendOnConflictQueryClause(tableName, queryString)
	queryString = p.appendReturningQueryClause(queryString)

	postgresJsonb, err := NewPostgresJSONbFromJSONb(jsonb)
	if err != nil {
		return nil, err
	}

	rows, err := p.db.NamedQuery(queryString, postgresJsonb)
	if err != nil {
		return nil, fmt.Errorf("Error saving JSONb to table: err: %v", err)
	}
	defer func() {
		rerr := rows.Close()
		if rerr != nil {
			log.Errorf("Error closing rows: err: %v", rerr)
		}
	}()

	updatedJSONb := &PostgresJSONb{}
	for rows.Next() {
		serr := rows.StructScan(updatedJSONb)
		if serr != nil {
			log.Errorf("Error scanning struct: err: %v", serr)
		}
	}

	jsb, err := updatedJSONb.PostgresJSONbToJSONb()
	if err != nil {
		return nil, fmt.Errorf("Error converting postgresjson to jsonb: err: %v", err)
	}

	return jsb, nil
}

// Appends a custom on conflict statement to an existing insert statement to
// update the raw json if the ID already exists i.e., upsert.
func (p *PostgresPersister) appendOnConflictQueryClause(tableName string, queryString string) string {
	// Remove the ; at the end of the existing query
	splitQueryStrs := strings.Split(queryString, ";")
	queryString = splitQueryStrs[0]

	// Upsert clause to the query when there is an existing extry with the same ID
	queryBuf := bytes.NewBufferString(queryString)
	queryBuf.WriteString(" ON CONFLICT (key)")
	queryBuf.WriteString(" DO UPDATE")
	queryBuf.WriteString(" SET raw_json = EXCLUDED.raw_json, hash = EXCLUDED.hash,")
	// Make sure the created_date stays the same but the last updated is updated
	queryBuf.WriteString(fmt.Sprintf(" created_date = %v.created_date, last_updated_date = EXCLUDED.last_updated_date;", tableName))
	return queryBuf.String()
}

func (p *PostgresPersister) appendReturningQueryClause(queryString string) string {
	// Remove the ; at the end of the existing query
	splitQueryStrs := strings.Split(queryString, ";")
	queryString = splitQueryStrs[0]

	queryBuf := bytes.NewBufferString(queryString)
	queryBuf.WriteString(" RETURNING key, hash, id, created_date, last_updated_date, raw_json;")
	return queryBuf.String()
}

func (p *PostgresPersister) createJsonbIndicesForTable(tableName string) error {
	indexQuery := CreateJsonbTableIndicesString(tableName)
	_, err := p.db.Exec(indexQuery)
	return err
}

// CreateJsonbTableQuery returns the query to create this table
func CreateJsonbTableQuery(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            key TEXT PRIMARY KEY UNIQUE,
			hash TEXT,
			id TEXT,
			namespace TEXT,
            created_date BIGINT,
            last_updated_date BIGINT,
            raw_json JSONB
        );
    `, tableName)
	return queryString
}

// CreateJsonbTableIndicesString returns the query to create indices on the
// JSONb table
func CreateJsonbTableIndicesString(tableName string) string {
	queryString := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS jsonb_hash_idx ON %s (hash);
	`, tableName)
	return queryString
}

func addWhereAnd(buf *bytes.Buffer) {
	if !strings.Contains(buf.String(), "WHERE") {
		buf.WriteString(" WHERE") // nolint: gosec
	} else {
		buf.WriteString(" AND") // nolint: gosec
	}
}

func retrieveJsonbQuery(tableName string, key string, hash string) string {
	queryBuf := bytes.NewBufferString("SELECT ")
	fields, _ := cpostgres.StructFieldsForQuery(PostgresJSONb{}, false, "")
	queryBuf.WriteString(fields)    // nolint: gosec
	queryBuf.WriteString(" FROM ")  // nolint: gosec
	queryBuf.WriteString(tableName) // nolint: gosec
	if key != "" {
		queryBuf.WriteString(" WHERE key = :key") // nolint: gosec
	}
	if hash != "" {
		addWhereAnd(queryBuf)
		queryBuf.WriteString(" hash = :hash") // nolint: gosec
	}
	return queryBuf.String()
}

// NewPostgresJSONbFromJSONb creates a new PostgresJSONb from a JSONb model
func NewPostgresJSONbFromJSONb(jsonb *JSONb) (*PostgresJSONb, error) {
	createdTs := ctime.ToSecsFromEpoch(&jsonb.CreatedDate)
	lastUpdatedTs := ctime.ToSecsFromEpoch(&jsonb.LastUpdatedDate)

	jsonbPayload := &cpostgres.JsonbPayload{}
	err := json.Unmarshal([]byte(jsonb.RawJSON), &jsonbPayload)
	if err != nil {
		return nil, err
	}
	return &PostgresJSONb{
		Hash:            jsonb.Hash,
		ID:              jsonb.ID,
		Key:             jsonb.Key,
		Namespace:       jsonb.Namespace,
		CreatedDate:     createdTs,
		LastUpdatedDate: lastUpdatedTs,
		JSONb:           *jsonbPayload,
	}, nil
}

// PostgresJSONb is the Postgresql model for the JSONb data model.
type PostgresJSONb struct {
	Key string `db:"key"`

	Hash string `db:"hash"`

	ID string `db:"id"`

	Namespace string `db:"namespace"`

	CreatedDate int64 `db:"created_date"`

	LastUpdatedDate int64 `db:"last_updated_date"`

	JSONb cpostgres.JsonbPayload `db:"raw_json"`
}

// JSONbStr returns the JSON string for the JSONb JsonbPayload
func (p *PostgresJSONb) JSONbStr() (string, error) {
	jsonStr, err := json.Marshal(p.JSONb)
	if err != nil {
		return "", err
	}
	return string(jsonStr), nil
}

// PostgresJSONbToJSONb returns this postgres specific struct as a JSONb struct
func (p *PostgresJSONb) PostgresJSONbToJSONb() (*JSONb, error) {
	jsonb := &JSONb{}
	jsonb.Key = p.Key
	jsonb.Hash = p.Hash
	jsonb.ID = p.ID
	jsonb.Namespace = p.Namespace
	jsonb.CreatedDate = ctime.SecsFromEpochToTime(p.CreatedDate)
	jsonb.LastUpdatedDate = ctime.SecsFromEpochToTime(p.LastUpdatedDate)
	jsonStr, err := p.JSONbStr()
	if err != nil {
		return nil, err
	}
	jsonb.RawJSON = jsonStr
	err = jsonb.ValidateRawJSON()
	if err != nil {
		return nil, err
	}
	err = jsonb.RawJSONToFields()
	if err != nil {
		return nil, err
	}
	return jsonb, nil
}
