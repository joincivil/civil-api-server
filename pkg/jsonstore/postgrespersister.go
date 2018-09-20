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

	crawlerpg "github.com/joincivil/civil-events-crawler/pkg/persistence/postgres"
	processorutils "github.com/joincivil/civil-events-processor/pkg/utils"
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
func (p *PostgresPersister) SaveJsonb(jsonb *JSONb) error {
	return p.createJsonbForTable(jsonb, defaultJsonbTableName)
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

func (p *PostgresPersister) jsonbFromTable(id string, hash string, tableName string) (
	[]*JSONb, error) {
	queryString := retrieveJsonbQuery(tableName, id, hash)
	jsonbs := []*PostgresJSONb{}
	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error preparing query with sqlx: %v", err)
	}
	criteria := map[string]interface{}{
		"id":   id,
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

func (p *PostgresPersister) createJsonbForTable(jsonb *JSONb, tableName string) error {
	queryString := crawlerpg.InsertIntoDBQueryString(tableName, PostgresJSONb{})
	postgresJsonb, err := NewPostgresJSONbFromJSONb(jsonb)
	if err != nil {
		return err
	}
	_, err = p.db.NamedExec(queryString, postgresJsonb)
	if err != nil {
		return fmt.Errorf("Error saving JSONb to table: err: %v", err)
	}
	return nil
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
            hash TEXT PRIMARY KEY,
            id TEXT,
            created_date BIGINT,
            raw_json JSONB
        );
    `, tableName)
	return queryString
}

// CreateJsonbTableIndicesString returns the query to create indices on the
// JSONb table
func CreateJsonbTableIndicesString(tableName string) string {
	queryString := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS jsonb_id_idx ON %s (id);
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

func retrieveJsonbQuery(tableName string, id string, hash string) string {
	queryBuf := bytes.NewBufferString("SELECT ")
	fields, _ := crawlerpg.StructFieldsForQuery(PostgresJSONb{}, false)
	queryBuf.WriteString(fields)    // nolint: gosec
	queryBuf.WriteString(" FROM ")  // nolint: gosec
	queryBuf.WriteString(tableName) // nolint: gosec
	if id != "" {
		queryBuf.WriteString(" WHERE id = :id") // nolint: gosec
	}
	if hash != "" {
		addWhereAnd(queryBuf)
		queryBuf.WriteString(" hash = :hash") // nolint: gosec
	}
	return queryBuf.String()
}

// NewPostgresJSONbFromJSONb creates a new PostgresJSONb from a JSONb model
func NewPostgresJSONbFromJSONb(jsonb *JSONb) (*PostgresJSONb, error) {
	createdTs := processorutils.TimeToSecsFromEpoch(&jsonb.CreatedDate)
	jsonbPayload := &crawlerpg.JsonbPayload{}
	err := json.Unmarshal([]byte(jsonb.RawJSON), &jsonbPayload)
	if err != nil {
		return nil, err
	}
	return &PostgresJSONb{
		Hash:        jsonb.Hash,
		ID:          jsonb.ID,
		CreatedDate: createdTs,
		JSONb:       *jsonbPayload,
	}, nil
}

// PostgresJSONb is the Postgresql model for the JSONb data model.
type PostgresJSONb struct {
	ID string `db:"id"`

	Hash string `db:"hash"`

	CreatedDate int64 `db:"created_date"`

	JSONb crawlerpg.JsonbPayload `db:"raw_json"`
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
	jsonb.Hash = p.Hash
	jsonb.ID = p.ID
	jsonb.CreatedDate = processorutils.SecsFromEpochToTime(p.CreatedDate)
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
