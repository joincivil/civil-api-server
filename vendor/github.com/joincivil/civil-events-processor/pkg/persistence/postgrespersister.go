// Package persistence contains components to interact with the DB
package persistence // import "github.com/joincivil/civil-events-processor/pkg/persistence"

import (
	"bytes"
	"database/sql"
	"fmt"
	// log "github.com/golang/glog"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	// driver for postgresql
	_ "github.com/lib/pq"

	crawlerutils "github.com/joincivil/civil-events-crawler/pkg/utils"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/persistence/postgres"
)

const (
	listingTableName  = "listing"
	contRevTableName  = "content_revision"
	govEventTableName = "governance_event"
	cronTableName     = "cron"

	lastUpdatedDateDBModelName = "LastUpdatedDateTs"

	// Could make this configurable later if needed
	maxOpenConns    = 20
	maxIdleConns    = 5
	connMaxLifetime = time.Nanosecond
)

// NewPostgresPersister creates a new postgres persister
func NewPostgresPersister(host string, port int, user string, password string, dbname string) (*PostgresPersister, error) {
	pgPersister := &PostgresPersister{}
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
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

// PostgresPersister holds the DB connection and persistence
type PostgresPersister struct {
	db *sqlx.DB
}

// ListingsByCriteria returns a slice of Listings by ListingCriteria
func (p *PostgresPersister) ListingsByCriteria(criteria *model.ListingCriteria) ([]*model.Listing, error) {
	return p.listingsByCriteriaFromTable(criteria, listingTableName)
}

// ListingsByAddresses returns a slice of Listings in order based on addresses
// NOTE(IS): If one of these listings is not found, empty *model.Listing will be returned in the list
func (p *PostgresPersister) ListingsByAddresses(addresses []common.Address) ([]*model.Listing, error) {
	return p.listingsByAddressesFromTableInOrder(addresses, listingTableName)
}

// ListingByAddress retrieves listings based on addresses
func (p *PostgresPersister) ListingByAddress(address common.Address) (*model.Listing, error) {
	return p.listingByAddressFromTable(address, listingTableName)
}

// CreateListing creates a new listing
func (p *PostgresPersister) CreateListing(listing *model.Listing) error {
	return p.createListingForTable(listing, listingTableName)
}

// UpdateListing updates fields on an existing listing
func (p *PostgresPersister) UpdateListing(listing *model.Listing, updatedFields []string) error {
	return p.updateListingInTable(listing, updatedFields, listingTableName)
}

// DeleteListing removes a listing
func (p *PostgresPersister) DeleteListing(listing *model.Listing) error {
	return p.deleteListingFromTable(listing, listingTableName)
}

// CreateContentRevision creates a new content revision
func (p *PostgresPersister) CreateContentRevision(revision *model.ContentRevision) error {
	return p.createContentRevisionForTable(revision, contRevTableName)
}

// ContentRevision retrieves a specific content revision for newsroom content
func (p *PostgresPersister) ContentRevision(address common.Address, contentID *big.Int, revisionID *big.Int) (*model.ContentRevision, error) {
	return p.contentRevisionFromTable(address, contentID, revisionID, contRevTableName)
}

// ContentRevisionsByCriteria returns a list of ContentRevision by ContentRevisionCriteria
func (p *PostgresPersister) ContentRevisionsByCriteria(criteria *model.ContentRevisionCriteria) (
	[]*model.ContentRevision, error) {
	return p.contentRevisionsByCriteriaFromTable(criteria, contRevTableName)
}

// ContentRevisions retrieves the revisions for content on a listing
func (p *PostgresPersister) ContentRevisions(address common.Address, contentID *big.Int) ([]*model.ContentRevision, error) {
	return p.contentRevisionsFromTable(address, contentID, contRevTableName)
}

// UpdateContentRevision updates fields on an existing content revision
func (p *PostgresPersister) UpdateContentRevision(revision *model.ContentRevision, updatedFields []string) error {
	return p.updateContentRevisionInTable(revision, updatedFields, contRevTableName)
}

// DeleteContentRevision removes a content revision
func (p *PostgresPersister) DeleteContentRevision(revision *model.ContentRevision) error {
	return p.deleteContentRevisionFromTable(revision, contRevTableName)
}

// GovernanceEventsByCriteria retrieves governance events based on criteria
func (p *PostgresPersister) GovernanceEventsByCriteria(criteria *model.GovernanceEventCriteria) ([]*model.GovernanceEvent, error) {
	return p.governanceEventsByCriteriaFromTable(criteria, govEventTableName)
}

// GovernanceEventsByListingAddress retrieves governance events based on listing address
func (p *PostgresPersister) GovernanceEventsByListingAddress(address common.Address) ([]*model.GovernanceEvent, error) {
	return p.governanceEventsByListingAddressFromTable(address, govEventTableName)
}

// GovernanceEventsByTxHash retrieves governance events based on TxHash
func (p *PostgresPersister) GovernanceEventsByTxHash(txHash common.Hash) ([]*model.GovernanceEvent, error) {
	return p.governanceEventsByTxHashFromTable(txHash, govEventTableName)
}

// GovernanceEventByChallengeID retrieves challenge by challengeID
func (p *PostgresPersister) GovernanceEventByChallengeID(challengeID int) (*model.GovernanceEvent, error) {
	challengeIDs := []int{challengeID}
	govEvents, err := p.govEventsByChallengeIDsFromTable(challengeIDs, govEventTableName)
	if len(govEvents) > 0 {
		return govEvents[0], err
	}
	return nil, err
}

// GovernanceEventsByChallengeIDs retrieves challenges by challengeIDs
func (p *PostgresPersister) GovernanceEventsByChallengeIDs(challengeIDs []int) ([]*model.GovernanceEvent, error) {
	return p.govEventsByChallengeIDsFromTable(challengeIDs, govEventTableName)
}

// CreateGovernanceEvent creates a new governance event
func (p *PostgresPersister) CreateGovernanceEvent(govEvent *model.GovernanceEvent) error {
	return p.createGovernanceEventInTable(govEvent, govEventTableName)
}

// UpdateGovernanceEvent updates fields on an existing governance event
func (p *PostgresPersister) UpdateGovernanceEvent(govEvent *model.GovernanceEvent, updatedFields []string) error {
	return p.updateGovernanceEventInTable(govEvent, updatedFields, govEventTableName)
}

// DeleteGovernanceEvent removes a governance event
func (p *PostgresPersister) DeleteGovernanceEvent(govEvent *model.GovernanceEvent) error {
	return p.deleteGovernanceEventFromTable(govEvent, govEventTableName)
}

// TimestampOfLastEventForCron returns the last timestamp from cron
func (p *PostgresPersister) TimestampOfLastEventForCron() (int64, error) {
	return p.lastCronTimestampFromTable(cronTableName)
}

// UpdateTimestampForCron updates the timestamp saved in cron table
func (p *PostgresPersister) UpdateTimestampForCron(timestamp int64) error {
	return p.updateCronTimestampInTable(timestamp, cronTableName)
}

// CreateTables creates the tables for processor if they don't exist
func (p *PostgresPersister) CreateTables() error {
	// this needs to get all the event tables for processor
	contRevTableQuery := postgres.CreateContentRevisionTableQuery()
	govEventTableQuery := postgres.CreateGovernanceEventTableQuery()
	listingTableQuery := postgres.CreateListingTableQuery()
	cronTableQuery := postgres.CreateCronTableQuery()

	_, err := p.db.Exec(contRevTableQuery)
	if err != nil {
		return fmt.Errorf("Error creating content_revision table in postgres: %v", err)
	}
	_, err = p.db.Exec(govEventTableQuery)
	if err != nil {
		return fmt.Errorf("Error creating governance_event table in postgres: %v", err)
	}
	_, err = p.db.Exec(listingTableQuery)
	if err != nil {
		return fmt.Errorf("Error creating listing table in postgres: %v", err)
	}
	_, err = p.db.Exec(cronTableQuery)
	if err != nil {
		return fmt.Errorf("Error creating listing table in postgres: %v", err)
	}
	return err
}

func (p *PostgresPersister) insertIntoDBQueryString(tableName string, dbModelStruct interface{}) string {
	fieldNames, fieldNamesColon := postgres.StructFieldsForQuery(dbModelStruct, true)
	queryString := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s);", tableName, fieldNames, fieldNamesColon) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) updateDBQueryBuffer(updatedFields []string, tableName string, dbModelStruct interface{}) (bytes.Buffer, error) {
	var queryBuf bytes.Buffer
	queryBuf.WriteString("UPDATE ") // nolint: gosec
	queryBuf.WriteString(tableName) // nolint: gosec
	queryBuf.WriteString(" SET ")   // nolint: gosec
	for idx, field := range updatedFields {
		dbFieldName, err := postgres.DbFieldNameFromModelName(dbModelStruct, field)
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

func (p *PostgresPersister) listingsByCriteriaFromTable(criteria *model.ListingCriteria,
	tableName string) ([]*model.Listing, error) {
	dbListings := []postgres.Listing{}
	queryString := p.listingsByCriteriaQuery(criteria, tableName)
	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error preparing query with sqlx: %v", err)
	}
	err = nstmt.Select(&dbListings, criteria)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving listings from table: %v", err)
	}
	listings := make([]*model.Listing, len(dbListings))
	for index, dbListing := range dbListings {
		modelListing := dbListing.DbToListingData()
		listings[index] = modelListing
	}
	return listings, nil
}

func (p *PostgresPersister) listingsByAddressesFromTable(addresses []common.Address, tableName string) ([]*model.Listing, error) {
	stringAddresses := postgres.ListCommonAddressToListString(addresses)
	queryString := p.listingByAddressesQuery(tableName)
	query, args, err := sqlx.In(queryString, stringAddresses)
	if err != nil {
		return nil, fmt.Errorf("Error preparing 'IN' statement for listings by address query: %v", err)
	}
	query = p.db.Rebind(query)
	rows, err := p.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving listings from table: %v", err)
	}

	listings := []*model.Listing{}
	for rows.Next() {
		var dbListing postgres.Listing
		err = rows.StructScan(&dbListing)
		if err != nil {
			return nil, fmt.Errorf("Error scanning row from IN query: %v", err)
		}
		listings = append(listings, dbListing.DbToListingData())
	}
	return listings, nil
}

func (p *PostgresPersister) listingsByAddressesFromTableInOrder(addresses []common.Address, tableName string) ([]*model.Listing, error) {
	stringAddresses := postgres.ListCommonAddressToListString(addresses)
	queryString := p.listingByAddressesQuery(tableName)
	query, args, err := sqlx.In(queryString, stringAddresses)
	if err != nil {
		return nil, fmt.Errorf("Error preparing 'IN' statement: %v", err)
	}
	query = p.db.Rebind(query)
	rows, err := p.db.Queryx(query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			err = model.ErrPersisterNoResults
		}
		return nil, fmt.Errorf("Error retrieving listings from table: %v", err)
	}

	listingsMap := map[common.Address]*model.Listing{}
	for rows.Next() {
		var dbListing postgres.Listing
		err = rows.StructScan(&dbListing)
		if err != nil {
			return nil, fmt.Errorf("Error scanning row from IN query: %v", err)
		}
		modelListing := dbListing.DbToListingData()
		listingsMap[modelListing.ContractAddress()] = modelListing
	}
	// NOTE(IS): This is not ideal, but we should return the listings in same order as addresses (also needed for dataloader in api-server)
	// so looping through listings again.
	listings := make([]*model.Listing, len(addresses))
	for i, address := range addresses {
		retrievedListing, ok := listingsMap[address]
		if ok {
			listings[i] = retrievedListing
		} else {
			listings[i] = nil
		}
	}
	return listings, nil
}

func (p *PostgresPersister) listingByAddressFromTable(address common.Address, tableName string) (*model.Listing, error) {
	listings, err := p.listingsByAddressesFromTable([]common.Address{address}, tableName)
	if len(listings) > 0 {
		return listings[0], err
	}
	return nil, err

}

func (p *PostgresPersister) listingsByCriteriaQuery(criteria *model.ListingCriteria,
	tableName string) string {
	queryBuf := bytes.NewBufferString("SELECT ")
	fieldNames, _ := postgres.StructFieldsForQuery(postgres.Listing{}, false)
	queryBuf.WriteString(fieldNames) // nolint: gosec
	queryBuf.WriteString(" FROM ")   // nolint: gosec
	queryBuf.WriteString(tableName)  // nolint: gosec
	if criteria.CreatedFromTs > 0 {
		queryBuf.WriteString(" WHERE creation_timestamp > :created_fromts") // nolint: gosec
	}
	if criteria.WhitelistedOnly {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" whitelisted = true") // nolint: gosec
	}
	if criteria.CreatedBeforeTs > 0 {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" creation_timestamp < :created_beforets") // nolint: gosec
	}
	if criteria.Offset > 0 {
		queryBuf.WriteString(" OFFSET :offset") // nolint: gosec
	}
	if criteria.Count > 0 {
		queryBuf.WriteString(" LIMIT :count") // nolint: gosec
	}
	return queryBuf.String()
}

func (p *PostgresPersister) listingByAddressesQuery(tableName string) string {
	fieldNames, _ := postgres.StructFieldsForQuery(postgres.Listing{}, false)
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE contract_address IN (?);", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) createListingForTable(listing *model.Listing, tableName string) error {
	dbListing := postgres.NewListing(listing)
	queryString := p.insertIntoDBQueryString(tableName, postgres.Listing{})
	_, err := p.db.NamedExec(queryString, dbListing)
	if err != nil {
		return fmt.Errorf("Error saving listing to table: %v", err)
	}
	return nil
}

func (p *PostgresPersister) updateListingInTable(listing *model.Listing, updatedFields []string, tableName string) error {
	// Update the last updated timestamp
	listing.SetLastUpdatedDateTs(crawlerutils.CurrentEpochSecsInInt64())
	updatedFields = append(updatedFields, lastUpdatedDateDBModelName)

	queryString, err := p.updateListingQuery(updatedFields, tableName)
	if err != nil {
		return fmt.Errorf("Error creating query string for update: %v ", err)
	}
	dbListing := postgres.NewListing(listing)
	_, err = p.db.NamedExec(queryString, dbListing)
	if err != nil {
		return fmt.Errorf("Error updating fields in db: %v", err)
	}
	return nil
}

func (p *PostgresPersister) updateListingQuery(updatedFields []string, tableName string) (string, error) {
	queryString, err := p.updateDBQueryBuffer(updatedFields, tableName, postgres.Listing{})
	if err != nil {
		return "", err
	}
	queryString.WriteString(" WHERE contract_address=:contract_address;") // nolint: gosec
	return queryString.String(), nil
}

func (p *PostgresPersister) deleteListingFromTable(listing *model.Listing, tableName string) error {
	dbListing := postgres.NewListing(listing)
	queryString := p.deleteListingQuery(tableName)
	_, err := p.db.NamedExec(queryString, dbListing)
	if err != nil {
		return fmt.Errorf("Error deleting listing in db: %v", err)
	}
	return nil
}

func (p *PostgresPersister) deleteListingQuery(tableName string) string {
	queryString := fmt.Sprintf("DELETE FROM %s WHERE contract_address=:contract_address", tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) createContentRevisionForTable(revision *model.ContentRevision, tableName string) error {
	queryString := p.insertIntoDBQueryString(tableName, postgres.ContentRevision{})
	dbContRev := postgres.NewContentRevision(revision)
	_, err := p.db.NamedExec(queryString, dbContRev)
	if err != nil {
		return fmt.Errorf("Error saving contentRevision to table: %v", err)
	}
	return nil
}

func (p *PostgresPersister) contentRevisionFromTable(address common.Address, contentID *big.Int, revisionID *big.Int, tableName string) (*model.ContentRevision, error) {
	dbContRev := postgres.ContentRevision{}
	queryString := p.contentRevisionQuery(tableName)
	err := p.db.Get(&dbContRev, queryString, address.Hex(), contentID.Int64(), revisionID.Int64())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrPersisterNoResults
		}
		return nil, fmt.Errorf("Wasn't able to get ContentRevision from postgres table: %v", err)
	}
	contRev := dbContRev.DbToContentRevisionData()
	return contRev, err
}

func (p *PostgresPersister) contentRevisionQuery(tableName string) string {
	fieldNames, _ := postgres.StructFieldsForQuery(postgres.ContentRevision{}, false)
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE (listing_address=$1 AND contract_content_id=$2 AND contract_revision_id=$3)", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) contentRevisionsFromTable(address common.Address, contentID *big.Int, tableName string) ([]*model.ContentRevision, error) {
	contRevs := []*model.ContentRevision{}
	dbContRevs := []postgres.ContentRevision{}
	queryString := p.contentRevisionsQuery(tableName)
	err := p.db.Select(&dbContRevs, queryString, address.Hex(), contentID.Int64())
	if err != nil {
		if err == sql.ErrNoRows {
			return contRevs, model.ErrPersisterNoResults
		}
		return contRevs, fmt.Errorf("Wasn't able to get ContentRevisions from postgres table: %v", err)
	}
	for _, dbContRev := range dbContRevs {
		contRevs = append(contRevs, dbContRev.DbToContentRevisionData())
	}
	return contRevs, err
}

func (p *PostgresPersister) contentRevisionsQuery(tableName string) string {
	fieldNames, _ := postgres.StructFieldsForQuery(postgres.ContentRevision{}, false)
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE (listing_address=$1 AND contract_content_id=$2)", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) contentRevisionsByCriteriaFromTable(criteria *model.ContentRevisionCriteria,
	tableName string) ([]*model.ContentRevision, error) {
	dbContRevs := []postgres.ContentRevision{}
	queryString := p.contentRevisionsByCriteriaQuery(criteria, tableName)

	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error preparing query with sqlx: %v", err)
	}
	err = nstmt.Select(&dbContRevs, criteria)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving content revisions from table: %v", err)
	}
	revisions := make([]*model.ContentRevision, len(dbContRevs))
	for index, dbContRev := range dbContRevs {
		modelRev := dbContRev.DbToContentRevisionData()
		revisions[index] = modelRev
	}
	return revisions, err
}

func (p *PostgresPersister) contentRevisionsByCriteriaQuery(criteria *model.ContentRevisionCriteria,
	tableName string) string {
	queryBuf := bytes.NewBufferString("SELECT ")
	fieldNames, _ := postgres.StructFieldsForQuery(postgres.ContentRevision{}, false)
	queryBuf.WriteString(fieldNames) // nolint: gosec
	queryBuf.WriteString(" FROM ")   // nolint: gosec
	queryBuf.WriteString(tableName)  // nolint: gosec
	queryBuf.WriteString(" r1 ")     // nolint: gosec

	if criteria.ListingAddress != "" {
		queryBuf.WriteString(" WHERE r1.listing_address = :listing_address") // nolint: gosec
	}
	if criteria.LatestOnly {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" r1.revision_timestamp =")                              // nolint: gosec
		queryBuf.WriteString(" (SELECT max(revision_timestamp) FROM ")                // nolint: gosec
		queryBuf.WriteString(tableName)                                               // nolint: gosec
		queryBuf.WriteString(" r2 WHERE r1.listing_address = r2.listing_address AND") // nolint: gosec
		queryBuf.WriteString(" r1.contract_content_id = r2.contract_content_id)")     // nolint: gosec
	} else {
		if criteria.FromTs > 0 {
			p.addWhereAnd(queryBuf)
			queryBuf.WriteString(" r1.revision_timestamp > :fromts") // nolint: gosec
		}
		if criteria.BeforeTs > 0 {
			p.addWhereAnd(queryBuf)
			queryBuf.WriteString(" r1.revision_timestamp < :beforets") // nolint: gosec
		}
	}
	if criteria.Offset > 0 {
		queryBuf.WriteString(" OFFSET :offset") // nolint: gosec
	}
	if criteria.Count > 0 {
		queryBuf.WriteString(" LIMIT :count") // nolint: gosec
	}
	return queryBuf.String()
}

func (p *PostgresPersister) updateContentRevisionInTable(revision *model.ContentRevision, updatedFields []string, tableName string) error {
	queryString, err := p.updateContentRevisionQuery(updatedFields, tableName)
	if err != nil {
		return fmt.Errorf("Error creating query string for update: %v ", err)
	}
	dbContentRevision := postgres.NewContentRevision(revision)
	_, err = p.db.NamedExec(queryString, dbContentRevision)
	if err != nil {
		return fmt.Errorf("Error updating fields in db: %v", err)
	}
	return nil
}

func (p *PostgresPersister) updateContentRevisionQuery(updatedFields []string, tableName string) (string, error) {
	queryString, err := p.updateDBQueryBuffer(updatedFields, tableName, postgres.ContentRevision{})
	if err != nil {
		return "", err
	}
	queryString.WriteString(" WHERE (listing_address=:listing_address AND contract_content_id=:contract_content_id AND contract_revision_id=:contract_revision_id);") // nolint: gosec
	return queryString.String(), nil
}

func (p *PostgresPersister) deleteContentRevisionFromTable(revision *model.ContentRevision, tableName string) error {
	dbContRev := postgres.NewContentRevision(revision)
	queryString := p.deleteContentRevisionQuery(tableName)
	_, err := p.db.NamedExec(queryString, dbContRev)
	if err != nil {
		return fmt.Errorf("Error deleting content revision in db: %v", err)
	}
	return nil
}

func (p *PostgresPersister) deleteContentRevisionQuery(tableName string) string {
	queryString := fmt.Sprintf("DELETE FROM %s WHERE (listing_address=:listing_address AND contract_content_id=:contract_content_id AND contract_revision_id=:contract_revision_id)", tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) governanceEventsByListingAddressFromTable(address common.Address, tableName string) ([]*model.GovernanceEvent, error) {
	govEvents := []*model.GovernanceEvent{}
	queryString := p.govEventsQuery(tableName)
	dbGovEvents := []postgres.GovernanceEvent{}
	err := p.db.Select(&dbGovEvents, queryString, address.Hex())
	if err != nil {
		if err == sql.ErrNoRows {
			err = model.ErrPersisterNoResults
		}
		return govEvents, fmt.Errorf("Error retrieving governance events from table: %v", err)
	}
	// retrieved correctly
	for _, dbGovEvent := range dbGovEvents {
		govEvents = append(govEvents, dbGovEvent.DbToGovernanceData())
	}
	return govEvents, nil
}

func (p *PostgresPersister) governanceEventsByTxHashFromTable(txHash common.Hash, tableName string) ([]*model.GovernanceEvent, error) {
	queryString := p.governanceEventsByTxHashQuery(txHash, tableName)
	rows, err := p.db.Queryx(queryString)
	if err != nil {
		if err == sql.ErrNoRows {
			err = model.ErrPersisterNoResults
		}
		return nil, fmt.Errorf("Error retrieving governance events from table: %v", err)
	}
	return p.scanGovEvents(rows)
}

func (p *PostgresPersister) govEventsByChallengeIDsFromTable(challengeIDs []int, tableName string) ([]*model.GovernanceEvent, error) {
	queryString := p.govEventsByChallengeIDQuery(tableName, challengeIDs)
	rows, err := p.db.Queryx(queryString)
	if err != nil {
		if err == sql.ErrNoRows {
			err = model.ErrPersisterNoResults
		}
		return nil, fmt.Errorf("Error retrieving governance events from table: %v", err)
	}

	govEventsMap := map[int]*model.GovernanceEvent{}
	for rows.Next() {
		var dbGovEvent postgres.GovernanceEvent
		err = rows.StructScan(&dbGovEvent)
		if err != nil {
			return nil, fmt.Errorf("Error scanning governance_event row from IN query: %v", err)
		}
		modelGovEvent := dbGovEvent.DbToGovernanceData()
		challengeID := int(modelGovEvent.Metadata()["ChallengeID"].(float64))
		govEventsMap[challengeID] = modelGovEvent
	}
	// Return govEvents in order
	modelGovEvents := make([]*model.GovernanceEvent, len(challengeIDs))
	for i, id := range challengeIDs {
		retrievedGovEvent, ok := govEventsMap[id]
		if ok {
			modelGovEvents[i] = retrievedGovEvent
		} else {
			modelGovEvents[i] = nil
		}
	}
	return modelGovEvents, err
}

func (p *PostgresPersister) scanGovEvents(rows *sqlx.Rows) ([]*model.GovernanceEvent, error) {
	govEvents := []*model.GovernanceEvent{}
	govEvent := postgres.GovernanceEvent{}
	for rows.Next() {
		err := rows.StructScan(&govEvent)
		govEvents = append(govEvents, govEvent.DbToGovernanceData())
		if err != nil {
			return govEvents, fmt.Errorf("Error scanning results from governance event query: %v", err)
		}
	}
	return govEvents, nil
}

func (p *PostgresPersister) governanceEventsByTxHashQuery(txHash common.Hash, tableName string) string {
	fieldNames, _ := postgres.StructFieldsForQuery(postgres.GovernanceEvent{}, false)
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE block_data @> '{\"txHash\": \"%s\" }'", fieldNames,
		tableName, txHash.Hex())
	return queryString
}

func (p *PostgresPersister) govEventsQuery(tableName string) string {
	fieldNames, _ := postgres.StructFieldsForQuery(postgres.GovernanceEvent{}, false)
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE listing_address=$1", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) govEventsByChallengeIDQuery(tableName string, challengeIDs []int) string {
	fieldNames, _ := postgres.StructFieldsForQuery(postgres.GovernanceEvent{}, false)
	var idbuf bytes.Buffer
	for _, id := range challengeIDs {
		idbuf.WriteString(fmt.Sprintf("'%d',", id)) // nolint: gosec
	}
	// take out extra comma
	idbuf.Truncate(idbuf.Len() - 1)
	ids := idbuf.String()
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE gov_event_type='Challenge' AND metadata ->>'ChallengeID' IN (%s);",
		fieldNames, tableName, ids)
	return queryString
}

func (p *PostgresPersister) createGovernanceEventInTable(govEvent *model.GovernanceEvent, tableName string) error {
	dbGovEvent := postgres.NewGovernanceEvent(govEvent)
	queryString := p.insertIntoDBQueryString(tableName, postgres.GovernanceEvent{})
	_, err := p.db.NamedExec(queryString, dbGovEvent)
	if err != nil {
		return fmt.Errorf("Error saving GovernanceEvent to table: %v", err)
	}
	return nil
}

func (p *PostgresPersister) governanceEventsByCriteriaFromTable(criteria *model.GovernanceEventCriteria,
	tableName string) ([]*model.GovernanceEvent, error) {
	dbGovEvents := []postgres.GovernanceEvent{}
	queryString := p.governanceEventsByCriteriaQuery(criteria, tableName)
	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error preparing query with sqlx: %v", err)
	}
	err = nstmt.Select(&dbGovEvents, criteria)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving gov events from table: %v", err)
	}
	events := make([]*model.GovernanceEvent, len(dbGovEvents))
	for index, event := range dbGovEvents {
		modelEvent := event.DbToGovernanceData()
		events[index] = modelEvent
	}
	return events, err
}

func (p *PostgresPersister) governanceEventsByCriteriaQuery(criteria *model.GovernanceEventCriteria,
	tableName string) string {
	queryBuf := bytes.NewBufferString("SELECT ")
	fieldNames, _ := postgres.StructFieldsForQuery(postgres.GovernanceEvent{}, false)
	queryBuf.WriteString(fieldNames) // nolint: gosec
	queryBuf.WriteString(" FROM ")   // nolint: gosec
	queryBuf.WriteString(tableName)  // nolint: gosec
	queryBuf.WriteString(" r1 ")     // nolint: gosec

	if criteria.ListingAddress != "" {
		queryBuf.WriteString(" WHERE r1.listing_address = :listing_address") // nolint: gosec
	}
	if criteria.CreatedFromTs > 0 {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" r1.creation_date > :created_fromts") // nolint: gosec
	}
	if criteria.CreatedBeforeTs > 0 {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" r1.creation_date < :created_beforets") // nolint: gosec
	}
	if criteria.Offset > 0 {
		queryBuf.WriteString(" OFFSET :offset") // nolint: gosec
	}
	if criteria.Count > 0 {
		queryBuf.WriteString(" LIMIT :count") // nolint: gosec
	}
	return queryBuf.String()
}

func (p *PostgresPersister) updateGovernanceEventInTable(govEvent *model.GovernanceEvent, updatedFields []string, tableName string) error {
	// Update the last updated timestamp
	govEvent.SetLastUpdatedDateTs(crawlerutils.CurrentEpochSecsInInt64())
	updatedFields = append(updatedFields, lastUpdatedDateDBModelName)

	queryString, err := p.updateGovEventsQuery(updatedFields, tableName)
	if err != nil {
		return fmt.Errorf("Error creating query string for update: %v ", err)
	}
	dbGovEvent := postgres.NewGovernanceEvent(govEvent)
	_, err = p.db.NamedExec(queryString, dbGovEvent)
	if err != nil {
		return fmt.Errorf("Error updating fields in db: %v", err)
	}
	return nil
}

func (p *PostgresPersister) updateGovEventsQuery(updatedFields []string, tableName string) (string, error) {
	queryString, err := p.updateDBQueryBuffer(updatedFields, tableName, postgres.GovernanceEvent{})
	if err != nil {
		return "", err
	}
	queryString.WriteString(" WHERE event_hash=:event_hash;") // nolint: gosec
	return queryString.String(), nil
}

func (p *PostgresPersister) deleteGovernanceEventFromTable(govEvent *model.GovernanceEvent, tableName string) error {
	dbGovEvent := postgres.NewGovernanceEvent(govEvent)
	queryString := p.deleteGovEventQuery(tableName)
	_, err := p.db.NamedExec(queryString, dbGovEvent)
	if err != nil {
		return fmt.Errorf("Error deleting governanceEvent in db: %v", err)
	}
	return nil
}

func (p *PostgresPersister) deleteGovEventQuery(tableName string) string {
	queryString := fmt.Sprintf("DELETE FROM %s WHERE event_hash=:event_hash;", tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) lastCronTimestampFromTable(tableName string) (int64, error) {
	var timestampInt int64
	// See if row with type timestamp exists
	timestampString, err := p.typeExistsInCronTable(tableName, postgres.TimestampDataType)
	if err != nil {
		if err == sql.ErrNoRows {
			// If there are no rows in DB, call updateCronTimestampInTable to do an insert of 0
			err = p.updateCronTimestampInTable(timestampInt, tableName) // nolint: gosec
			if err != nil {
				return timestampInt, fmt.Errorf("No row in %s with timestamp. Error updating table, %v", tableName, err)
			}
			return timestampInt, nil
		}
		return timestampInt, fmt.Errorf("Wasn't able to get listing from postgres table: %v", err)
	}
	timestampInt, err = postgres.StringToTimestamp(timestampString)
	return timestampInt, err
}

func (p *PostgresPersister) updateCronTimestampInTable(timestamp int64, tableName string) error {
	// Check if timestamp row exists
	timestampExists := true
	cronData := postgres.NewCronData(postgres.TimestampToString(timestamp), postgres.TimestampDataType)

	_, err := p.typeExistsInCronTable(tableName, cronData.DataType)
	if err != nil {
		if err == sql.ErrNoRows {
			timestampExists = false
		} else {
			return fmt.Errorf("Error checking DB for cron row, %v", err)
		}
	}

	var queryString string
	if timestampExists {
		// update query
		updatedFields := []string{postgres.DataPersistedModelName}
		queryBuff, errBuff := p.updateDBQueryBuffer(updatedFields, tableName, postgres.CronData{})
		if errBuff != nil {
			return err
		}
		queryString = queryBuff.String()
	} else {
		//insert query
		queryString = p.insertIntoDBQueryString(tableName, postgres.CronData{})
	}

	_, err = p.db.NamedExec(queryString, cronData)
	if err != nil {
		return fmt.Errorf("Error updating fields in db: %v", err)
	}

	return nil
}

func (p *PostgresPersister) typeExistsInCronTable(tableName string, dataType string) (string, error) {
	dbCronData := []postgres.CronData{}
	queryString := fmt.Sprintf(`SELECT * FROM %s WHERE data_type=$1;`, tableName) // nolint: gosec
	err := p.db.Select(&dbCronData, queryString, dataType)
	if err != nil {
		return "", err
	}
	if len(dbCronData) == 0 {
		return "", sql.ErrNoRows
	}
	if len(dbCronData) > 1 {
		return "", fmt.Errorf("There should not be more than 1 row with type %s in %s table", dataType, tableName)
	}
	return dbCronData[0].DataPersisted, nil
}

func (p *PostgresPersister) addWhereAnd(buf *bytes.Buffer) {
	if !strings.Contains(buf.String(), "WHERE") {
		buf.WriteString(" WHERE") // nolint: gosec
	} else {
		buf.WriteString(" AND") // nolint: gosec
	}
}
