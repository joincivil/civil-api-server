// Package persistence contains components to interact with the DB
package persistence // import "github.com/joincivil/civil-events-processor/pkg/persistence"

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"

	"math/big"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"

	// driver for postgresql
	_ "github.com/lib/pq"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

	crawlerPostgres "github.com/joincivil/civil-events-crawler/pkg/persistence/postgres"

	cbytes "github.com/joincivil/go-common/pkg/bytes"
	cpersist "github.com/joincivil/go-common/pkg/persistence"
	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
	cstrings "github.com/joincivil/go-common/pkg/strings"

	ctime "github.com/joincivil/go-common/pkg/time"
)

// NOTE(IS): cpersist.ErrPersisterNoResults is only returned for single queries

var (
	// ErrNoRowsAffected is returned when a query affects no rows. Mainly returned
	// by update methods.
	ErrNoRowsAffected = errors.New("no rows affected on update")
)

const (
	// ProcessorServiceName is the name for the processor service
	ProcessorServiceName       = "processor"
	lastUpdatedDateDBModelName = "LastUpdatedDateTs"

	// Could make this configurable later if needed
	maxOpenConns    = 50
	maxIdleConns    = 10
	connMaxLifetime = time.Second * 1800 // 30 mins
)

// NewPostgresPersister creates a new postgres persister
func NewPostgresPersister(host string, port int, user string, password string, dbname string) (*PostgresPersister, error) {
	pgPersister := &PostgresPersister{}
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return pgPersister, errors.Wrap(err, "error connecting to sqlx")
	}
	pgPersister.db = db
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	return pgPersister, nil
}

// PostgresPersister holds the DB connection and persistence
type PostgresPersister struct {
	db      *sqlx.DB
	version *string
}

// GetTableName formats tabletype with version of this persister to return the table name
func (p *PostgresPersister) GetTableName(tableType string) string {
	if p.version == nil || *p.version == "" {
		return tableType
	}
	return fmt.Sprintf("%s_%s", tableType, *p.version)
}

// Close shuts down the connections to postgres
func (p *PostgresPersister) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *PostgresPersister) closeRows(rows *sqlx.Rows) {
	if rows == nil {
		return
	}
	err := rows.Close()
	if err != nil {
		log.Errorf("Error closing rows: err: %v", err)
	}
}

// ListingsByCriteria returns a slice of Listings by ListingCriteria sorted by creation timestamp
func (p *PostgresPersister) ListingsByCriteria(criteria *model.ListingCriteria) ([]*model.Listing, error) {
	listingTableName := p.GetTableName(postgres.ListingTableBaseName)
	challengeTableName := p.GetTableName(postgres.ChallengeTableBaseName)
	return p.listingsByCriteriaFromTable(criteria, listingTableName, challengeTableName)
}

// ListingsByAddresses returns a slice of Listings in order based on addresses
// NOTE(IS): If one of these listings is not found, empty *model.Listing will be returned in the list
func (p *PostgresPersister) ListingsByAddresses(addresses []common.Address) ([]*model.Listing, error) {
	listingTableName := p.GetTableName(postgres.ListingTableBaseName)
	return p.listingsByAddressesFromTableInOrder(addresses, listingTableName)
}

// ListingByAddress retrieves listings based on addresses
func (p *PostgresPersister) ListingByAddress(address common.Address) (*model.Listing, error) {
	listingTableName := p.GetTableName(postgres.ListingTableBaseName)
	return p.listingByAddressFromTable(address, listingTableName)
}

// CreateListing creates a new listing
func (p *PostgresPersister) CreateListing(listing *model.Listing) error {
	listingTableName := p.GetTableName(postgres.ListingTableBaseName)
	return p.createListingForTable(listing, listingTableName)
}

// UpdateListing updates fields on an existing listing
func (p *PostgresPersister) UpdateListing(listing *model.Listing, updatedFields []string) error {
	listingTableName := p.GetTableName(postgres.ListingTableBaseName)
	return p.updateListingInTable(listing, updatedFields, listingTableName)
}

// DeleteListing removes a listing
func (p *PostgresPersister) DeleteListing(listing *model.Listing) error {
	listingTableName := p.GetTableName(postgres.ListingTableBaseName)
	return p.deleteListingFromTable(listing, listingTableName)
}

// CreateContentRevision creates a new content revision
func (p *PostgresPersister) CreateContentRevision(revision *model.ContentRevision) error {
	contRevTableName := p.GetTableName(postgres.ContentRevisionTableBaseName)
	return p.createContentRevisionForTable(revision, contRevTableName)
}

// ContentRevision retrieves a specific content revision for newsroom content
func (p *PostgresPersister) ContentRevision(address common.Address, contentID *big.Int, revisionID *big.Int) (*model.ContentRevision, error) {
	contRevTableName := p.GetTableName(postgres.ContentRevisionTableBaseName)
	return p.contentRevisionFromTable(address, contentID, revisionID, contRevTableName)
}

// ContentRevisionsByCriteria returns a list of ContentRevision by ContentRevisionCriteria sorted by revision timestamp
func (p *PostgresPersister) ContentRevisionsByCriteria(criteria *model.ContentRevisionCriteria) (
	[]*model.ContentRevision, error) {
	contRevTableName := p.GetTableName(postgres.ContentRevisionTableBaseName)
	return p.contentRevisionsByCriteriaFromTable(criteria, contRevTableName)
}

// ContentRevisions retrieves the revisions for content on a listing sorted by revision timestamp
func (p *PostgresPersister) ContentRevisions(address common.Address, contentID *big.Int) ([]*model.ContentRevision, error) {
	contRevTableName := p.GetTableName(postgres.ContentRevisionTableBaseName)
	return p.contentRevisionsFromTable(address, contentID, contRevTableName)
}

// UpdateContentRevision updates fields on an existing content revision
func (p *PostgresPersister) UpdateContentRevision(revision *model.ContentRevision, updatedFields []string) error {
	contRevTableName := p.GetTableName(postgres.ContentRevisionTableBaseName)
	return p.updateContentRevisionInTable(revision, updatedFields, contRevTableName)
}

// DeleteContentRevision removes a content revision
func (p *PostgresPersister) DeleteContentRevision(revision *model.ContentRevision) error {
	contRevTableName := p.GetTableName(postgres.ContentRevisionTableBaseName)
	return p.deleteContentRevisionFromTable(revision, contRevTableName)
}

// GovernanceEventsByCriteria retrieves governance events based on criteria sorted by revision timestamp
func (p *PostgresPersister) GovernanceEventsByCriteria(criteria *model.GovernanceEventCriteria) ([]*model.GovernanceEvent, error) {
	govEventTableName := p.GetTableName(postgres.GovernanceEventTableBaseName)
	return p.governanceEventsByCriteriaFromTable(criteria, govEventTableName)
}

// GovernanceEventsByListingAddress retrieves governance events based on listing address
func (p *PostgresPersister) GovernanceEventsByListingAddress(address common.Address) ([]*model.GovernanceEvent, error) {
	govEventTableName := p.GetTableName(postgres.GovernanceEventTableBaseName)
	return p.governanceEventsByListingAddressFromTable(address, govEventTableName)
}

// GovernanceEventsByTxHash retrieves governance events based on TxHash sorted by revision timestamp
func (p *PostgresPersister) GovernanceEventsByTxHash(txHash common.Hash) ([]*model.GovernanceEvent, error) {
	govEventTableName := p.GetTableName(postgres.GovernanceEventTableBaseName)
	return p.governanceEventsByTxHashFromTable(txHash, govEventTableName)
}

// CreateGovernanceEvent creates a new governance event
func (p *PostgresPersister) CreateGovernanceEvent(govEvent *model.GovernanceEvent) error {
	govEventTableName := p.GetTableName(postgres.GovernanceEventTableBaseName)
	return p.createGovernanceEventInTable(govEvent, govEventTableName)
}

// UpdateGovernanceEvent updates fields on an existing governance event
func (p *PostgresPersister) UpdateGovernanceEvent(govEvent *model.GovernanceEvent, updatedFields []string) error {
	govEventTableName := p.GetTableName(postgres.GovernanceEventTableBaseName)
	return p.updateGovernanceEventInTable(govEvent, updatedFields, govEventTableName)
}

// DeleteGovernanceEvent removes a governance event
func (p *PostgresPersister) DeleteGovernanceEvent(govEvent *model.GovernanceEvent) error {
	govEventTableName := p.GetTableName(postgres.GovernanceEventTableBaseName)
	return p.deleteGovernanceEventFromTable(govEvent, govEventTableName)
}

// TimestampOfLastEventForCron returns the last timestamp from cron
func (p *PostgresPersister) TimestampOfLastEventForCron() (int64, error) {
	cronTableName := p.GetTableName(postgres.CronTableBaseName)
	return p.lastCronTimestampFromTable(cronTableName)
}

// EventHashesOfLastTimestampForCron returns the event hashes processed for the last timestamp from cron
func (p *PostgresPersister) EventHashesOfLastTimestampForCron() ([]string, error) {
	cronTableName := p.GetTableName(postgres.CronTableBaseName)
	return p.lastEventHashesFromTable(cronTableName)
}

// UpdateTimestampForCron updates the timestamp saved in cron table
func (p *PostgresPersister) UpdateTimestampForCron(timestamp int64) error {
	cronTableName := p.GetTableName(postgres.CronTableBaseName)
	return p.updateCronTimestampInTable(timestamp, cronTableName)
}

// UpdateEventHashesForCron updates the eventHashes saved in cron table
func (p *PostgresPersister) UpdateEventHashesForCron(eventHashes []string) error {
	cronTableName := p.GetTableName(postgres.CronTableBaseName)
	return p.updateEventHashesInTable(eventHashes, cronTableName)
}

// CreateChallenge creates a new challenge
func (p *PostgresPersister) CreateChallenge(challenge *model.Challenge) error {
	challengeTableName := p.GetTableName(postgres.ChallengeTableBaseName)
	return p.createChallengeInTable(challenge, challengeTableName)
}

// UpdateChallenge updates a challenge
func (p *PostgresPersister) UpdateChallenge(challenge *model.Challenge, updatedFields []string) error {
	challengeTableName := p.GetTableName(postgres.ChallengeTableBaseName)
	return p.updateChallengeInTable(challenge, updatedFields, challengeTableName)
}

// ChallengesByChallengeIDs returns a slice of challenges based on challenge IDs. Returns order of given challengeIDs
func (p *PostgresPersister) ChallengesByChallengeIDs(challengeIDs []int) ([]*model.Challenge, error) {
	challengeTableName := p.GetTableName(postgres.ChallengeTableBaseName)
	return p.challengesByChallengeIDsInTableInOrder(challengeIDs, challengeTableName)
}

// ChallengeByChallengeID gets a challenge by challengeID
func (p *PostgresPersister) ChallengeByChallengeID(challengeID int) (*model.Challenge, error) {
	challengeTableName := p.GetTableName(postgres.ChallengeTableBaseName)
	return p.challengeByChallengeIDFromTable(challengeID, challengeTableName)
}

// ChallengesByListingAddresses gets slice of challenges for a each listing address in order of given addresses
func (p *PostgresPersister) ChallengesByListingAddresses(addrs []common.Address) ([][]*model.Challenge, error) {
	challengeTableName := p.GetTableName(postgres.ChallengeTableBaseName)
	return p.challengesByListingAddressesInTable(addrs, challengeTableName)
}

// ChallengesByListingAddress gets a list of challenges for a listing sorted by challenge_id
func (p *PostgresPersister) ChallengesByListingAddress(addr common.Address) ([]*model.Challenge, error) {
	challengeTableName := p.GetTableName(postgres.ChallengeTableBaseName)
	return p.challengesByListingAddressInTable(addr, challengeTableName)
}

// PollByPollID gets a poll by pollID
func (p *PostgresPersister) PollByPollID(pollID int) (*model.Poll, error) {
	pollTableName := p.GetTableName(postgres.PollTableBaseName)
	return p.pollByPollIDFromTable(pollID, pollTableName)
}

// PollsByPollIDs returns a slice of polls in order based on poll IDs
// NOTE: This returns nills for polls that DNE in db.
func (p *PostgresPersister) PollsByPollIDs(pollIDs []int) ([]*model.Poll, error) {
	pollTableName := p.GetTableName(postgres.PollTableBaseName)
	return p.pollsByPollIDsInTableInOrder(pollIDs, pollTableName)
}

// CreatePoll creates a new poll
func (p *PostgresPersister) CreatePoll(poll *model.Poll) error {
	pollTableName := p.GetTableName(postgres.PollTableBaseName)
	return p.createPollInTable(poll, pollTableName)
}

// UpdatePoll updates a poll
func (p *PostgresPersister) UpdatePoll(poll *model.Poll, updatedFields []string) error {
	pollTableName := p.GetTableName(postgres.PollTableBaseName)
	return p.updatePollInTable(poll, updatedFields, pollTableName)
}

// AppealByChallengeID gets an appeal by challengeID
func (p *PostgresPersister) AppealByChallengeID(challengeID int) (*model.Appeal, error) {
	appealTableName := p.GetTableName(postgres.AppealTableBaseName)
	return p.appealByChallengeIDFromTable(challengeID, appealTableName)
}

// AppealsByChallengeIDs returns a slice of appeals in order based on challenge IDs
func (p *PostgresPersister) AppealsByChallengeIDs(challengeIDs []int) ([]*model.Appeal, error) {
	appealTableName := p.GetTableName(postgres.AppealTableBaseName)
	return p.appealsByChallengeIDsInTableInOrder(challengeIDs, appealTableName)
}

// AppealByAppealChallengeID returns an appeal based on appealchallengeID
func (p *PostgresPersister) AppealByAppealChallengeID(appealChallengeID int) (*model.Appeal, error) {
	appealTableName := p.GetTableName(postgres.AppealTableBaseName)
	return p.appealByAppealChallengeIDInTable(appealChallengeID, appealTableName)
}

// CreateAppeal creates a new appeal
func (p *PostgresPersister) CreateAppeal(appeal *model.Appeal) error {
	appealTableName := p.GetTableName(postgres.AppealTableBaseName)
	return p.createAppealInTable(appeal, appealTableName)
}

// UpdateAppeal updates an appeal
func (p *PostgresPersister) UpdateAppeal(appeal *model.Appeal, updatedFields []string) error {
	appealTableName := p.GetTableName(postgres.AppealTableBaseName)
	return p.updateAppealInTable(appeal, updatedFields, appealTableName)
}

// TokenTransfersByTxHash all the token transfers for a given purchaser address
func (p *PostgresPersister) TokenTransfersByTxHash(txHash common.Hash) (
	[]*model.TokenTransfer, error) {
	tokenTransferTableName := p.GetTableName(postgres.TokenTransferTableBaseName)
	return p.tokenTransfersByTxHashFromTable(txHash, tokenTransferTableName)
}

// TokenTransfersByToAddress gets all the token transfers for a given purchaser address
func (p *PostgresPersister) TokenTransfersByToAddress(addr common.Address) (
	[]*model.TokenTransfer, error) {
	tokenTransferTableName := p.GetTableName(postgres.TokenTransferTableBaseName)
	return p.tokenTransfersByToAddressFromTable(addr, tokenTransferTableName)
}

// CreateTokenTransfer creates a new token transfer
func (p *PostgresPersister) CreateTokenTransfer(purchase *model.TokenTransfer) error {
	tokenTransferTableName := p.GetTableName(postgres.TokenTransferTableBaseName)
	return p.createTokenTransferInTable(purchase, tokenTransferTableName)
}

// SaveVersion saves the version for this persistence
func (p *PostgresPersister) SaveVersion(versionNumber *string) error {
	if versionNumber == nil || *versionNumber == "" {
		return nil
	}

	err := p.saveVersionToTable(crawlerPostgres.VersionTableName, versionNumber)
	if err != nil {
		return err
	}
	p.version = versionNumber
	return nil
}

// PersisterVersion returns the latest version of this persistence
func (p *PostgresPersister) PersisterVersion() (*string, error) {
	return p.persisterVersionFromTable(crawlerPostgres.VersionTableName)
}

// InitProcessorVersion inits this persistence version to versionNumber if specified,
// else gets version from db
func (p *PostgresPersister) InitProcessorVersion(versionNumber *string) error {
	currentVersion, err := p.PersisterVersion()
	if err != nil && err != cpersist.ErrPersisterNoResults {
		return err
	}

	// If no version found anywhere, don't use versioned tables
	if (currentVersion == nil || *currentVersion == "") && (versionNumber == nil || *versionNumber == "") {
		log.Infof("No version found, not using versioned tables")
		return nil
	}

	// If the incoming version is the same as the currentVersion, don't do anything
	if currentVersion != nil && versionNumber != nil && *currentVersion == *versionNumber {
		log.Infof("Using data version: %v", *versionNumber)
		return nil
	}

	// If version does not exist, but currentVersion does, use currentVersion
	if currentVersion != nil && (versionNumber == nil || *versionNumber == "") {
		// NOTE(IS): Use existing version, but update timestamp
		versionNumber = currentVersion
		log.Infof("Using data version from DB, updating ts: %v", *versionNumber)

	} else {
		log.Infof("Updating data version: %v", *versionNumber)
	}

	p.version = versionNumber
	return p.SaveVersion(versionNumber)
}

// CreateParameterProposal creates a new parameter proposal
func (p *PostgresPersister) CreateParameterProposal(paramProposal *model.ParameterProposal) error {
	paramProposalTableName := p.GetTableName(postgres.ParameterProposalTableBaseName)
	return p.createParameterProposalInTable(paramProposal, paramProposalTableName)
}

// ParamProposalByPropID gets parameter proposal by propID
func (p *PostgresPersister) ParamProposalByPropID(propID [32]byte) (*model.ParameterProposal, error) {
	paramProposalTableName := p.GetTableName(postgres.ParameterProposalTableBaseName)
	return p.paramProposalByPropIDFromTable(propID, paramProposalTableName)
}

// ParamProposalByName gets parameter proposals by name. active=true will get only active
func (p *PostgresPersister) ParamProposalByName(name string, active bool) ([]*model.ParameterProposal, error) {
	paramProposalTableName := p.GetTableName(postgres.ParameterProposalTableBaseName)
	return p.paramProposalByNameFromTable(name, active, paramProposalTableName)
}

// UpdateParamProposal updates a parameter proposal
func (p *PostgresPersister) UpdateParamProposal(paramProposal *model.ParameterProposal,
	updatedFields []string) error {
	paramProposalTableName := p.GetTableName(postgres.ParameterProposalTableBaseName)
	return p.updateParamProposalInTable(paramProposal, updatedFields, paramProposalTableName)
}

// CreateUserChallengeData creates a new UserChallengeData
func (p *PostgresPersister) CreateUserChallengeData(userChallengeData *model.UserChallengeData) error {
	userChallengeDataTableName := p.GetTableName(postgres.UserChallengeDataTableBaseName)
	return p.createUserChallengeDataInTable(userChallengeData, userChallengeDataTableName)
}

// UserChallengeDataByCriteria retrieves UserChallengeData based on criteria
func (p *PostgresPersister) UserChallengeDataByCriteria(
	criteria *model.UserChallengeDataCriteria) ([]*model.UserChallengeData, error) {
	userChallengeDataTableName := p.GetTableName(postgres.UserChallengeDataTableBaseName)
	return p.userChallengeDataByCriteriaFromTable(criteria, userChallengeDataTableName)
}

// UpdateUserChallengeData updates UserChallengeData in table
// user=true updates for user + pollID, user=false updates for pollID
func (p *PostgresPersister) UpdateUserChallengeData(userChallengeData *model.UserChallengeData,
	updatedFields []string, updateWithUserAddress bool, latestVote bool) error {
	userChallengeDataTableName := p.GetTableName(postgres.UserChallengeDataTableBaseName)
	return p.updateUserChallengeDataInTable(userChallengeData, updatedFields, updateWithUserAddress,
		latestVote, userChallengeDataTableName)
}

// CreateTables creates the tables for processor if they don't exist
func (p *PostgresPersister) CreateTables() error {
	contRevTableQuery := postgres.CreateContentRevisionTableQuery(p.GetTableName(postgres.ContentRevisionTableBaseName))
	govEventTableQuery := postgres.CreateGovernanceEventTableQuery(p.GetTableName(postgres.GovernanceEventTableBaseName))
	listingTableQuery := postgres.CreateListingTableQuery(p.GetTableName(postgres.ListingTableBaseName))
	cronTableQuery := postgres.CreateCronTableQuery(p.GetTableName(postgres.CronTableBaseName))
	challengeTableQuery := postgres.CreateChallengeTableQuery(p.GetTableName(postgres.ChallengeTableBaseName))
	pollTableQuery := postgres.CreatePollTableQuery(p.GetTableName(postgres.PollTableBaseName))
	appealTableQuery := postgres.CreateAppealTableQuery(p.GetTableName(postgres.AppealTableBaseName))
	tokenTransferQuery := postgres.CreateTokenTransferTableQuery(p.GetTableName(postgres.TokenTransferTableBaseName))
	parameterProposalQuery := postgres.CreateParameterProposalTableQuery(p.GetTableName(postgres.ParameterProposalTableBaseName))
	userChallengeDataQuery := postgres.CreateUserChallengeDataTableQuery(p.GetTableName(postgres.UserChallengeDataTableBaseName))

	_, err := p.db.Exec(contRevTableQuery)
	if err != nil {
		return errors.Wrap(err, "error creating content_revision table in postgres")
	}
	_, err = p.db.Exec(govEventTableQuery)
	if err != nil {
		return errors.Wrap(err, "error creating governance_event table in postgres")
	}
	_, err = p.db.Exec(listingTableQuery)
	if err != nil {
		return errors.Wrap(err, "error creating listing table in postgres")
	}
	_, err = p.db.Exec(cronTableQuery)
	if err != nil {
		return errors.Wrap(err, "error creating cron table in postgres")
	}
	_, err = p.db.Exec(challengeTableQuery)
	if err != nil {
		return errors.Wrap(err, "error creating challenge table in postgres")
	}
	_, err = p.db.Exec(pollTableQuery)
	if err != nil {
		return errors.Wrap(err, "error creating poll table in postgres")
	}
	_, err = p.db.Exec(appealTableQuery)
	if err != nil {
		return errors.Wrap(err, "error creating appeal table in postgres")
	}
	_, err = p.db.Exec(tokenTransferQuery)
	if err != nil {
		return errors.Wrap(err, "error creating token transfer table in postgres")
	}
	_, err = p.db.Exec(parameterProposalQuery)
	if err != nil {
		return fmt.Errorf("Error creating parameter proposal table in postgres: %v", err)
	}
	_, err = p.db.Exec(userChallengeDataQuery)
	if err != nil {
		return fmt.Errorf("Error creating user_challenge_data table in postgres: %v", err)
	}
	return nil
}

// CreateIndices creates the indices for DB if they don't exist
func (p *PostgresPersister) CreateIndices() error {
	indexQuery := postgres.CreateContentRevisionTableIndicesQuery(p.GetTableName(postgres.ContentRevisionTableBaseName))
	_, err := p.db.Exec(indexQuery)
	if err != nil {
		return errors.Wrap(err, "error creating content revision table indices")
	}
	indexQuery = postgres.CreateGovernanceEventTableIndicesQuery(p.GetTableName(postgres.GovernanceEventTableBaseName))
	_, err = p.db.Exec(indexQuery)
	if err != nil {
		return errors.Wrap(err, "error creating gov events table indices")
	}
	indexQuery = postgres.CreateListingTableIndicesQuery(p.GetTableName(postgres.ListingTableBaseName))
	_, err = p.db.Exec(indexQuery)
	if err != nil {
		return errors.Wrap(err, "error creating listing table indices")
	}
	indexQuery = postgres.CreateChallengeTableIndicesQuery(p.GetTableName(postgres.ChallengeTableBaseName))
	_, err = p.db.Exec(indexQuery)
	if err != nil {
		return errors.Wrap(err, "error creating challenge table indices")
	}
	indexQuery = postgres.UserChallengeDataTableIndicesQuery(p.GetTableName(postgres.UserChallengeDataTableBaseName))
	_, err = p.db.Exec(indexQuery)
	if err != nil {
		return errors.Wrap(err, "error creating user_challenge_data table indices")
	}
	// indexQuery = postgres.CreatePollTableIndicesQuery(postgres.PollTableBaseName)
	// _, err = p.db.Exec(indexQuery)
	// if err != nil {
	// 	return errors.Wrap(err, "Error creating poll table indices in postgres")
	// }
	indexQuery = postgres.CreateAppealTableIndicesQuery(p.GetTableName(postgres.AppealTableBaseName))
	_, err = p.db.Exec(indexQuery)
	if err != nil {
		return errors.Wrap(err, "Error creating appeal table indices in postgres")
	}
	indexQuery = postgres.CreateTokenTransferTableIndicesQuery(p.GetTableName(postgres.TokenTransferTableBaseName))
	_, err = p.db.Exec(indexQuery)
	if err != nil {
		return errors.Wrap(err, "error creating token_transfer table indices")
	}
	return err
}

func (p *PostgresPersister) RunMigrations() error {
	return nil
}

func (p *PostgresPersister) persisterVersionFromTable(tableName string) (*string, error) {
	if p.version == nil {
		version, err := p.retrieveVersionFromTable(tableName)
		if err != nil {
			return nil, err
		}
		p.version = version
	}
	return p.version, nil
}

func (p *PostgresPersister) retrieveVersionFromTable(tableName string) (*string, error) {
	dbVersion := []crawlerPostgres.Version{}
	queryString := fmt.Sprintf(`SELECT * FROM %s WHERE service_name=$1 ORDER BY last_updated_timestamp DESC LIMIT 1;`, tableName) // nolint: gosec
	err := p.db.Select(&dbVersion, queryString, ProcessorServiceName)
	if err != nil {
		return nil, err
	}
	if len(dbVersion) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}
	return dbVersion[0].Version, nil
}

// saveVersionToTable saves the version
func (p *PostgresPersister) saveVersionToTable(tableName string, versionNumber *string) error {
	dbVersionStruct := crawlerPostgres.Version{
		Version:           versionNumber,
		ServiceName:       ProcessorServiceName,
		LastUpdatedDateTs: ctime.CurrentEpochSecsInInt64(),
		Exists:            true}
	onConflict := fmt.Sprintf("%s, %s", crawlerPostgres.VersionFieldName, crawlerPostgres.ServiceFieldName)
	updateFields := []string{crawlerPostgres.LastUpdatedTsFieldName, crawlerPostgres.ExistsFieldName}
	queryString := p.upsertVersionDataQueryString(tableName, dbVersionStruct, onConflict,
		updateFields)
	_, err := p.db.NamedExec(queryString, dbVersionStruct)
	if err != nil {
		return fmt.Errorf("Error saving version to table: %v", err)
	}
	return nil
}

func (p *PostgresPersister) upsertVersionDataQueryString(tableName string, dbModelStruct interface{},
	onConflict string, updatedFields []string) string {
	var queryString strings.Builder
	fieldNames, fieldNamesColon := cpostgres.StructFieldsForQuery(dbModelStruct, true, "")
	// nolint
	queryString.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s) ON CONFLICT(%s) DO UPDATE SET ",
		tableName, fieldNames, fieldNamesColon, onConflict))
	for idx, field := range updatedFields {
		queryString.WriteString(fmt.Sprintf("%s=:%s", field, field)) // nolint: gosec
		if idx+1 < len(updatedFields) {
			queryString.WriteString(", ") // nolint: gosec
		}
	}
	return queryString.String()
}

func (p *PostgresPersister) insertIntoDBQueryString(tableName string, dbModelStruct interface{}) string {
	fieldNames, fieldNamesColon := cpostgres.StructFieldsForQuery(dbModelStruct, true, "")
	queryString := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s);", tableName, fieldNames, fieldNamesColon) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) updateDBQueryBuffer(updatedFields []string, tableName string, dbModelStruct interface{}) (bytes.Buffer, error) {
	var queryBuf bytes.Buffer
	queryBuf.WriteString("UPDATE ") // nolint: gosec
	queryBuf.WriteString(tableName) // nolint: gosec
	queryBuf.WriteString(" SET ")   // nolint: gosec
	for idx, field := range updatedFields {
		dbFieldName, err := cpostgres.DbFieldNameFromModelName(dbModelStruct, field)
		if err != nil {
			return queryBuf, errors.Wrapf(err, "error getting %s from %s table DB struct tag", field, tableName)
		}
		queryBuf.WriteString(fmt.Sprintf("%s=:%s", dbFieldName, dbFieldName)) // nolint: gosec
		if idx+1 < len(updatedFields) {
			queryBuf.WriteString(", ") // nolint: gosec
		}
	}
	return queryBuf, nil
}

func (p *PostgresPersister) listingsByCriteriaFromTable(criteria *model.ListingCriteria,
	tableName string, joinTableName string) ([]*model.Listing, error) {
	dbListings := []postgres.Listing{}
	queryString, err := p.listingsByCriteriaQuery(criteria, tableName, joinTableName)
	if err != nil {
		return nil, err
	}
	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing query with sqlx")
	}
	err = nstmt.Select(&dbListings, criteria)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving listings from table")
	}
	listings := make([]*model.Listing, len(dbListings))
	for index, dbListing := range dbListings {
		modelListing := dbListing.DbToListingData()
		listings[index] = modelListing
	}
	return listings, nil
}

func (p *PostgresPersister) listingsByAddressesFromTableInOrder(addresses []common.Address,
	tableName string) ([]*model.Listing, error) {
	if len(addresses) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}

	stringAddresses := cstrings.ListCommonAddressToListString(addresses)
	queryString := p.listingByAddressesQuery(tableName)
	query, args, err := sqlx.In(queryString, stringAddresses)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing 'IN' statement")
	}

	query = p.db.Rebind(query)
	rows, err := p.db.Queryx(query, args...)
	defer p.closeRows(rows)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving listings from table")
	}

	listingsMap := map[common.Address]*model.Listing{}
	for rows.Next() {
		var dbListing postgres.Listing
		err = rows.StructScan(&dbListing)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning row from IN query")
		}
		modelListing := dbListing.DbToListingData()
		listingsMap[modelListing.ContractAddress()] = modelListing
	}

	// NOTE(IS): This is not ideal, but we should return the listings in same
	// order as addresses (also needed for dataloader in api-server)
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
	listings, err := p.listingsByAddressesFromTableInOrder([]common.Address{address}, tableName)
	if len(listings) > 0 {
		if listings[0] == nil {
			err = cpersist.ErrPersisterNoResults
		}
		return listings[0], err
	}
	return nil, err

}

func (p *PostgresPersister) listingsByCriteriaQuery(criteria *model.ListingCriteria,
	tableName string, joinTableName string) (string, error) {
	queryBuf := bytes.NewBufferString("SELECT ")
	var fieldNames string
	if criteria.ActiveChallenge && criteria.CurrentApplication {
		fieldNames, _ = cpostgres.StructFieldsForQuery(postgres.Listing{}, false, "l")
	} else {
		fieldNames, _ = cpostgres.StructFieldsForQuery(postgres.Listing{}, false, "")
	}

	queryBuf.WriteString(fieldNames) // nolint: gosec
	queryBuf.WriteString(" FROM ")   // nolint: gosec
	queryBuf.WriteString(tableName)  // nolint: gosec

	if criteria.WhitelistedOnly {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" whitelisted = true") // nolint: gosec

	} else if criteria.RejectedOnly {
		p.addWhereAnd(queryBuf)
		// whitelisted = false
		// challenge_id = 0 (not -1 or greater)
		// last_gov_state != ListingWithdrawn (which indicates a complete withdrawal from the registry)
		queryBuf.WriteString(" whitelisted = false AND challenge_id = 0 AND last_governance_state != ") // nolint: gosec
		queryBuf.WriteString(strconv.Itoa(int(model.GovernanceStateListingWithdrawn)))                  // nolint: gosec

	} else if criteria.ActiveChallenge && criteria.CurrentApplication {
		if joinTableName == "" {
			return "", errors.New("Expecting joinTable Name, cannot construct query string")
		}

		joinQuery := fmt.Sprintf(` l LEFT JOIN %v c ON l.challenge_id=c.challenge_id WHERE
			(l.challenge_id > 0 AND c.resolved=false)
			OR (l.app_expiry > 0 AND l.whitelisted = false AND l.challenge_id <= 0)`, joinTableName) // nolint: gosec
		queryBuf.WriteString(joinQuery) // nolint: gosec

	} else if criteria.ActiveChallenge {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" challenge_id > 0") // nolint: gosec

	} else if criteria.CurrentApplication {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" app_expiry > 0 AND whitelisted = false AND challenge_id <= 0") // nolint: gosec
	}

	if criteria.CreatedBeforeTs > 0 {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" creation_timestamp < :created_beforets") // nolint: gosec
	}

	if criteria.SortBy == model.SortByUndefined || criteria.SortBy == model.SortByCreated {
		queryBuf.WriteString(" ORDER BY creation_timestamp") // nolint: gosec

	} else if criteria.SortBy == model.SortByName {
		queryBuf.WriteString(" ORDER BY name") // nolint: gosec

	} else if criteria.SortBy == model.SortByApplied {
		if !criteria.ActiveChallenge && !criteria.CurrentApplication {
			p.addWhereAnd(queryBuf)
			queryBuf.WriteString(" application_timestamp > 0") // nolint: gosec
		}
		queryBuf.WriteString(" ORDER BY application_timestamp") // nolint: gosec

	} else if criteria.SortBy == model.SortByWhitelisted {
		if !criteria.WhitelistedOnly {
			p.addWhereAnd(queryBuf)
			queryBuf.WriteString(" approval_timestamp > 0") // nolint: gosec
		}
		queryBuf.WriteString(" ORDER BY approval_timestamp") // nolint: gosec
	}

	if criteria.SortDesc {
		queryBuf.WriteString(" DESC") // nolint: gosec
	}

	if criteria.Offset > 0 {
		queryBuf.WriteString(" OFFSET :offset") // nolint: gosec
	}

	if criteria.Count > 0 {
		queryBuf.WriteString(" LIMIT :count") // nolint: gosec
	}
	return queryBuf.String(), nil
}

func (p *PostgresPersister) listingByAddressesQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.Listing{}, false, "")
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE contract_address IN (?);", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) createListingForTable(listing *model.Listing, tableName string) error {
	dbListing := postgres.NewListing(listing)
	queryString := p.insertIntoDBQueryString(tableName, postgres.Listing{})
	_, err := p.db.NamedExec(queryString, dbListing)
	if err != nil {
		return errors.Wrap(err, "error saving listing to table")
	}
	return nil
}

func (p *PostgresPersister) updateListingInTable(listing *model.Listing, updatedFields []string, tableName string) error {
	listing.SetLastUpdatedDateTs(ctime.CurrentEpochSecsInInt64())
	updatedFields = append(updatedFields, lastUpdatedDateDBModelName)

	queryString, err := p.updateListingQuery(updatedFields, tableName)
	if err != nil {
		return errors.Wrap(err, "error creating query string for update")
	}
	dbListing := postgres.NewListing(listing)
	result, err := p.db.NamedExec(queryString, dbListing)
	if err != nil {
		return errors.Wrap(err, "error updating fields in db")
	}
	err = p.checkUpdateRowsAffected(result)
	if err != nil {
		return err
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
		return errors.Wrap(err, "error deleting listing in db")
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
		return errors.Wrap(err, "error saving contentRevision to table")
	}
	return nil
}

func (p *PostgresPersister) contentRevisionFromTable(address common.Address, contentID *big.Int, revisionID *big.Int, tableName string) (*model.ContentRevision, error) {
	dbContRev := postgres.ContentRevision{}
	queryString := p.contentRevisionQuery(tableName)
	err := p.db.Get(&dbContRev, queryString, address.Hex(), contentID.Int64(), revisionID.Int64())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cpersist.ErrPersisterNoResults
		}
		return nil, errors.Wrap(err, "wasn't able to get ContentRevision from postgres table")
	}
	contRev := dbContRev.DbToContentRevisionData()
	if contRev == nil {
		return contRev, cpersist.ErrPersisterNoResults
	}
	return contRev, err
}

func (p *PostgresPersister) contentRevisionQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.ContentRevision{}, false, "")
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE (listing_address=$1 AND contract_content_id=$2 AND contract_revision_id=$3)", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) contentRevisionsFromTable(address common.Address, contentID *big.Int, tableName string) ([]*model.ContentRevision, error) {
	contRevs := []*model.ContentRevision{}
	dbContRevs := []postgres.ContentRevision{}
	queryString := p.contentRevisionsQuery(tableName)
	err := p.db.Select(&dbContRevs, queryString, address.Hex(), contentID.Int64())
	if err != nil {
		return contRevs, errors.Wrap(err, "wasn't able to get ContentRevisions from postgres table")
	}
	for _, dbContRev := range dbContRevs {
		contRevs = append(contRevs, dbContRev.DbToContentRevisionData())
	}
	return contRevs, err
}

func (p *PostgresPersister) contentRevisionsQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.ContentRevision{}, false, "")
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE (listing_address=$1 AND contract_content_id=$2) ORDER BY revision_timestamp", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) contentRevisionsByCriteriaFromTable(criteria *model.ContentRevisionCriteria,
	tableName string) ([]*model.ContentRevision, error) {
	dbContRevs := []postgres.ContentRevision{}
	queryString := p.contentRevisionsByCriteriaQuery(criteria, tableName)

	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing query with sqlx")
	}
	err = nstmt.Select(&dbContRevs, criteria)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving content revisions from table")
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
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.ContentRevision{}, false, "")
	queryBuf.WriteString(fieldNames) // nolint: gosec
	queryBuf.WriteString(" FROM ")   // nolint: gosec
	queryBuf.WriteString(tableName)  // nolint: gosec
	queryBuf.WriteString(" r1 ")     // nolint: gosec

	if criteria.ListingAddress != "" {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" r1.listing_address = :listing_address") // nolint: gosec
	}
	if criteria.LatestOnly {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" r1.revision_timestamp =")                              // nolint: gosec
		queryBuf.WriteString(" (SELECT max(revision_timestamp) FROM ")                // nolint: gosec
		queryBuf.WriteString(tableName)                                               // nolint: gosec
		queryBuf.WriteString(" r2 WHERE r1.listing_address = r2.listing_address AND") // nolint: gosec
		queryBuf.WriteString(" r1.contract_content_id = r2.contract_content_id)")     // nolint: gosec
	} else {
		// If addr and contentID are passed, only retrieve revisions for that content ID
		if criteria.ListingAddress != "" && criteria.ContentID != nil {
			p.addWhereAnd(queryBuf)
			queryBuf.WriteString(" r1.contract_content_id = :content_id") // nolint: gosec
			// Retrieve a specific revision
			if criteria.RevisionID != nil {
				p.addWhereAnd(queryBuf)
				queryBuf.WriteString(" r1.contract_revision_id = :revision_id") // nolint: gosec
			}
		}
		if criteria.FromTs > 0 {
			p.addWhereAnd(queryBuf)
			queryBuf.WriteString(" r1.revision_timestamp > :fromts") // nolint: gosec
		}
		if criteria.BeforeTs > 0 {
			p.addWhereAnd(queryBuf)
			queryBuf.WriteString(" r1.revision_timestamp < :beforets") // nolint: gosec
		}
	}
	queryBuf.WriteString(" ORDER BY revision_timestamp") // nolint: gosec
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
		return errors.WithMessage(err, "error creating query string for update")
	}
	dbContentRevision := postgres.NewContentRevision(revision)

	result, err := p.db.NamedExec(queryString, dbContentRevision)
	if err != nil {
		return errors.Wrap(err, "error updating fields in db")
	}
	err = p.checkUpdateRowsAffected(result)
	if err != nil {
		return err
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
		return errors.Wrap(err, "error deleting content revision in db")
	}
	return nil
}

func (p *PostgresPersister) deleteContentRevisionQuery(tableName string) string {
	queryString := fmt.Sprintf("DELETE FROM %s WHERE (listing_address=:listing_address AND contract_content_id=:contract_content_id AND contract_revision_id=:contract_revision_id)", tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) governanceEventsByListingAddressFromTable(address common.Address,
	tableName string) ([]*model.GovernanceEvent, error) {
	govEvents := []*model.GovernanceEvent{}
	queryString := p.govEventsQuery(tableName)
	dbGovEvents := []postgres.GovernanceEvent{}
	err := p.db.Select(&dbGovEvents, queryString, address.Hex())
	if err != nil {
		return govEvents, errors.Wrap(err, "error retrieving governance events from table")
	}
	// retrieved correctly
	for _, dbGovEvent := range dbGovEvents {
		govEvents = append(govEvents, dbGovEvent.DbToGovernanceData())
	}
	return govEvents, nil
}

func (p *PostgresPersister) governanceEventsByTxHashFromTable(txHash common.Hash,
	tableName string) ([]*model.GovernanceEvent, error) {
	queryString := p.governanceEventsByTxHashQuery(tableName)

	blockDataValue := fmt.Sprintf("{ \"txHash\": \"%s\" }", txHash.Hex())
	rows, err := p.db.Queryx(queryString, blockDataValue)
	defer p.closeRows(rows)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving governance events from table")
	}
	return p.scanGovEvents(rows)
}

func (p *PostgresPersister) scanGovEvents(rows *sqlx.Rows) ([]*model.GovernanceEvent, error) {
	govEvents := []*model.GovernanceEvent{}
	govEvent := postgres.GovernanceEvent{}
	for rows.Next() {
		err := rows.StructScan(&govEvent)
		govEvents = append(govEvents, govEvent.DbToGovernanceData())
		if err != nil {
			return govEvents, errors.Wrap(err, "error scanning results from governance event query")
		}
	}
	return govEvents, nil
}

func (p *PostgresPersister) governanceEventsByTxHashQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.GovernanceEvent{}, false, "")
	queryString := fmt.Sprintf( // nolint: gosec
		"SELECT %s FROM %s WHERE block_data @> $1 ORDER BY creation_date",
		fieldNames,
		tableName,
	)
	return queryString
}

func (p *PostgresPersister) govEventsQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.GovernanceEvent{}, false, "")
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE listing_address=$1", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) createGovernanceEventInTable(govEvent *model.GovernanceEvent, tableName string) error {
	dbGovEvent := postgres.NewGovernanceEvent(govEvent)
	queryString := p.insertIntoDBQueryString(tableName, postgres.GovernanceEvent{})
	_, err := p.db.NamedExec(queryString, dbGovEvent)
	if err != nil {
		return errors.Wrap(err, "error saving GovernanceEvent to table")
	}
	return nil
}

func (p *PostgresPersister) governanceEventsByCriteriaFromTable(criteria *model.GovernanceEventCriteria,
	tableName string) ([]*model.GovernanceEvent, error) {
	dbGovEvents := []postgres.GovernanceEvent{}
	queryString := p.governanceEventsByCriteriaQuery(criteria, tableName)
	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing query with sqlx")
	}
	err = nstmt.Select(&dbGovEvents, criteria)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving gov events from table")
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
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.GovernanceEvent{}, false, "")
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
	queryBuf.WriteString(" ORDER BY creation_date") // nolint: gosec
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
	govEvent.SetLastUpdatedDateTs(ctime.CurrentEpochSecsInInt64())
	updatedFields = append(updatedFields, lastUpdatedDateDBModelName)

	queryString, err := p.updateGovEventsQuery(updatedFields, tableName)
	if err != nil {
		return errors.Wrap(err, "error creating query string for update")
	}
	dbGovEvent := postgres.NewGovernanceEvent(govEvent)
	result, err := p.db.NamedExec(queryString, dbGovEvent)
	if err != nil {
		return errors.Wrap(err, "error updating fields in db")
	}
	err = p.checkUpdateRowsAffected(result)
	if err != nil {
		return err
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
		return errors.Wrap(err, "error deleting governanceEvent in db")
	}
	return nil
}

func (p *PostgresPersister) deleteGovEventQuery(tableName string) string {
	queryString := fmt.Sprintf("DELETE FROM %s WHERE event_hash=:event_hash;", tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) createChallengeInTable(challenge *model.Challenge, tableName string) error {
	dbChallenge := postgres.NewChallenge(challenge)
	queryString := p.insertIntoDBQueryString(tableName, postgres.Challenge{})
	_, err := p.db.NamedExec(queryString, dbChallenge)
	if err != nil {
		return errors.Wrap(err, "error saving Challenge to table")
	}
	return nil
}

func (p *PostgresPersister) updateChallengeInTable(challenge *model.Challenge, updatedFields []string,
	tableName string) error {
	// Update the last updated timestamp
	challenge.SetLastUpdateDateTs(ctime.CurrentEpochSecsInInt64())
	updatedFields = append(updatedFields, lastUpdatedDateDBModelName)

	queryString, err := p.updateChallengeQuery(updatedFields, tableName)
	if err != nil {
		return errors.Wrap(err, "error creating query string for update")
	}

	dbChallenge := postgres.NewChallenge(challenge)
	result, err := p.db.NamedExec(queryString, dbChallenge)
	if err != nil {
		return errors.Wrap(err, "error updating fields in challenge table")
	}
	err = p.checkUpdateRowsAffected(result)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresPersister) updateChallengeQuery(updatedFields []string, tableName string) (string, error) {
	queryString, err := p.updateDBQueryBuffer(updatedFields, tableName, postgres.Challenge{})
	if err != nil {
		return "", err
	}
	queryString.WriteString(" WHERE challenge_id=:challenge_id;") // nolint: gosec
	return queryString.String(), nil
}

func (p *PostgresPersister) challengeByChallengeIDFromTable(challengeID int, tableName string) (*model.Challenge, error) {
	challenges, err := p.challengesByChallengeIDsInTableInOrder([]int{challengeID}, tableName)
	if err != nil {
		return nil, err
	}
	if challenges[0] == nil {
		return nil, cpersist.ErrPersisterNoResults
	}
	return challenges[0], nil
}

func (p *PostgresPersister) challengesByChallengeIDsInTableInOrder(challengeIDs []int,
	tableName string) ([]*model.Challenge, error) {
	if len(challengeIDs) <= 0 {
		return nil, cpersist.ErrPersisterNoResults
	}

	challengeIDsString := cstrings.ListIntToListString(challengeIDs)
	queryString := p.challengesByChallengeIDsQuery(tableName)
	query, args, err := sqlx.In(queryString, challengeIDsString)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing 'IN' statement")
	}
	query = p.db.Rebind(query)

	rows, err := p.db.Queryx(query, args...)

	defer p.closeRows(rows)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving challenges from table")
	}

	challengesMap := map[int]*model.Challenge{}
	for rows.Next() {
		var dbChallenge postgres.Challenge
		err = rows.StructScan(&dbChallenge)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning row from IN query")
		}

		modelChallenge := dbChallenge.DbToChallengeData()
		challengesMap[int(modelChallenge.ChallengeID().Int64())] = modelChallenge
	}

	// NOTE(IS): Return challenges in same order
	challenges := make([]*model.Challenge, len(challengeIDs))
	for i, challengeID := range challengeIDs {
		retrievedChallenge, ok := challengesMap[challengeID]
		if ok {
			challenges[i] = retrievedChallenge
		} else {
			challenges[i] = nil
		}
	}

	return challenges, nil
}

func (p *PostgresPersister) challengesByChallengeIDsQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.Challenge{}, false, "")
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE challenge_id IN (?);", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) challengesByListingAddressesInTable(addrs []common.Address,
	tableName string) ([][]*model.Challenge, error) {
	if len(addrs) <= 0 {
		return nil, cpersist.ErrPersisterNoResults
	}

	listingAddrs := cstrings.ListCommonAddressToListString(addrs)
	queryString := p.challengesByListingAddressesQuery(tableName)

	query, args, err := sqlx.In(queryString, listingAddrs)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing 'IN' statement")
	}

	query = p.db.Rebind(query)
	rows, err := p.db.Queryx(query, args...)
	defer p.closeRows(rows)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving challenges from table")
	}

	challengesMap := map[string][]*model.Challenge{}
	for rows.Next() {
		var dbChallenge postgres.Challenge
		err = rows.StructScan(&dbChallenge)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning row from IN query")
		}
		modelChallenge := dbChallenge.DbToChallengeData()
		listingAddr := modelChallenge.ListingAddress().Hex()

		listingChallenges, ok := challengesMap[listingAddr]
		if !ok {
			challengesMap[listingAddr] = []*model.Challenge{modelChallenge}
		} else {
			challengesMap[listingAddr] = append(listingChallenges, modelChallenge)
		}
	}

	// Retain ordering of listing addresses
	listingChallenges := make([][]*model.Challenge, len(addrs))
	for i, addr := range addrs {
		retrievedChallenges, ok := challengesMap[addr.Hex()]
		if ok {
			listingChallenges[i] = retrievedChallenges
		} else {
			listingChallenges[i] = nil
		}
	}

	return listingChallenges, nil
}

// challengesByListingAddressesQuery returns the query string to retrieved a list of
// challenges for a list of listing addresses
func (p *PostgresPersister) challengesByListingAddressesQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.Challenge{}, false, "")
	queryString := fmt.Sprintf( // nolint: gosec
		"SELECT %s FROM %s WHERE listing_address IN (?)",
		fieldNames,
		tableName,
	)
	return queryString
}

// challengesByListingAddressInTable retrieves a list of challenges for a listing sorted
// by challenge_id
func (p *PostgresPersister) challengesByListingAddressInTable(addr common.Address,
	tableName string) ([]*model.Challenge, error) {
	challenges := []*model.Challenge{}
	queryString := p.challengesByListingAddressQuery(tableName)

	dbChallenges := []*postgres.Challenge{}
	err := p.db.Select(&dbChallenges, queryString, addr.Hex())
	if err != nil {
		return challenges, errors.Wrap(err, "error retrieving challenges from table")
	}

	if len(dbChallenges) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}

	for _, dbChallenge := range dbChallenges {
		challenges = append(challenges, dbChallenge.DbToChallengeData())
	}

	return challenges, nil
}

// challengesByListingAddressQuery returns the query string to retrieved a list of
// challenges for a listing sorted by challenge_id
func (p *PostgresPersister) challengesByListingAddressQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.Challenge{}, false, "")
	queryString := fmt.Sprintf( // nolint: gosec
		"SELECT %s FROM %s WHERE listing_address = $1 ORDER BY challenge_id;",
		fieldNames,
		tableName,
	)
	return queryString
}

func (p *PostgresPersister) createPollInTable(poll *model.Poll, tableName string) error {
	dbPoll := postgres.NewPoll(poll)
	queryString := p.insertIntoDBQueryString(tableName, postgres.Poll{})
	_, err := p.db.NamedExec(queryString, dbPoll)
	if err != nil {
		return errors.Wrap(err, "error saving Poll to table")
	}
	return nil
}

func (p *PostgresPersister) updatePollInTable(poll *model.Poll, updatedFields []string,
	tableName string) error {
	// Update the last updated timestamp
	poll.SetLastUpdatedDateTs(ctime.CurrentEpochSecsInInt64())
	updatedFields = append(updatedFields, lastUpdatedDateDBModelName)

	queryString, err := p.updatePollQuery(updatedFields, tableName)
	if err != nil {
		return errors.Wrap(err, "error creating query string for update")
	}
	dbPoll := postgres.NewPoll(poll)
	result, err := p.db.NamedExec(queryString, dbPoll)
	if err != nil {
		return errors.Wrap(err, "error updating fields in poll table")
	}
	err = p.checkUpdateRowsAffected(result)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresPersister) updatePollQuery(updatedFields []string, tableName string) (string, error) {
	queryString, err := p.updateDBQueryBuffer(updatedFields, tableName, postgres.Poll{})
	if err != nil {
		return "", err
	}
	queryString.WriteString(" WHERE poll_id=:poll_id;") // nolint: gosec
	return queryString.String(), nil
}

func (p *PostgresPersister) pollByPollIDFromTable(pollID int, tableName string) (*model.Poll, error) {
	polls, err := p.pollsByPollIDsInTableInOrder([]int{pollID}, tableName)
	if err != nil {
		return nil, err
	}
	if polls[0] == nil {
		return nil, cpersist.ErrPersisterNoResults
	}
	return polls[0], nil
}

func (p *PostgresPersister) pollsByPollIDsInTableInOrder(pollIDs []int, pollTableName string) ([]*model.Poll, error) {
	if len(pollIDs) <= 0 {
		return nil, cpersist.ErrPersisterNoResults
	}

	pollIDsString := cstrings.ListIntToListString(pollIDs)
	queryString := p.pollByPollIDsQuery(pollTableName)
	query, args, err := sqlx.In(queryString, pollIDsString)
	if err != nil {
		return nil, errors.Wrapf(err, "error preparing 'IN' statement")
	}

	query = p.db.Rebind(query)
	rows, err := p.db.Queryx(query, args...)
	defer p.closeRows(rows)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving challenges from table")
	}

	pollsMap := map[int]*model.Poll{}
	for rows.Next() {
		var dbPoll postgres.Poll
		err = rows.StructScan(&dbPoll)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning row from IN query")
		}
		modelPoll := dbPoll.DbToPollData()
		pollsMap[int(modelPoll.PollID().Int64())] = modelPoll
	}

	// NOTE(IS): Return challenges in same order
	polls := make([]*model.Poll, len(pollIDs))
	for i, pollID := range pollIDs {
		retrievedPoll, ok := pollsMap[pollID]
		if ok {
			polls[i] = retrievedPoll
		} else {
			polls[i] = nil
		}
	}
	return polls, nil
}

func (p *PostgresPersister) pollByPollIDsQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.Poll{}, false, "")
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE poll_id IN (?);", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) createAppealInTable(appeal *model.Appeal, tableName string) error {
	dbAppeal := postgres.NewAppeal(appeal)
	queryString := p.insertIntoDBQueryString(tableName, postgres.Appeal{})
	_, err := p.db.NamedExec(queryString, dbAppeal)
	if err != nil {
		return errors.Wrap(err, "error saving appeal to table")
	}
	return nil
}

func (p *PostgresPersister) updateAppealInTable(appeal *model.Appeal, updatedFields []string,
	tableName string) error {
	// Update the last updated timestamp
	appeal.SetLastUpdatedDateTs(ctime.CurrentEpochSecsInInt64())
	updatedFields = append(updatedFields, lastUpdatedDateDBModelName)

	queryString, err := p.updateAppealQuery(updatedFields, tableName)
	if err != nil {
		return errors.WithMessage(err, "error creating query string for update")
	}

	dbAppeal := postgres.NewAppeal(appeal)
	result, err := p.db.NamedExec(queryString, dbAppeal)
	if err != nil {
		return errors.Wrap(err, "error updating fields in appeal table")
	}
	err = p.checkUpdateRowsAffected(result)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresPersister) updateAppealQuery(updatedFields []string, tableName string) (string, error) {
	queryString, err := p.updateDBQueryBuffer(updatedFields, tableName, postgres.Appeal{})
	if err != nil {
		return "", err
	}
	queryString.WriteString(" WHERE original_challenge_id=:original_challenge_id;") // nolint: gosec
	return queryString.String(), nil
}

func (p *PostgresPersister) appealByChallengeIDFromTable(challengeID int, tableName string) (*model.Appeal, error) {
	appeals, err := p.appealsByChallengeIDsInTableInOrder([]int{challengeID}, tableName)
	if err != nil {
		return nil, err
	}
	if appeals[0] == nil {
		return nil, cpersist.ErrPersisterNoResults
	}
	return appeals[0], nil
}

func (p *PostgresPersister) appealsByChallengeIDsInTableInOrder(challengeIDs []int, tableName string) ([]*model.Appeal, error) {
	if len(challengeIDs) <= 0 {
		return nil, cpersist.ErrPersisterNoResults
	}

	challengeIDsString := cstrings.ListIntToListString(challengeIDs)
	queryString := p.appealsByChallengeIDsQuery(tableName)
	query, args, err := sqlx.In(queryString, challengeIDsString)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing 'IN' statement")
	}

	query = p.db.Rebind(query)
	rows, err := p.db.Queryx(query, args...)
	defer p.closeRows(rows)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving challenges from table")
	}

	appealsMap := map[int]*model.Appeal{}
	for rows.Next() {
		var dbAppeal postgres.Appeal
		err = rows.StructScan(&dbAppeal)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning row from IN query")
		}
		modelAppeal := dbAppeal.DbToAppealData()
		appealsMap[int(modelAppeal.OriginalChallengeID().Int64())] = modelAppeal
	}

	// NOTE(IS): Return challenges in same order
	appeals := make([]*model.Appeal, len(challengeIDs))
	for i, challengeID := range challengeIDs {
		retrievedAppeal, ok := appealsMap[challengeID]
		if ok {
			appeals[i] = retrievedAppeal
		} else {
			appeals[i] = nil
		}
	}
	return appeals, nil
}

func (p *PostgresPersister) appealByAppealChallengeIDInTable(appealChallengeID int,
	tableName string) (*model.Appeal, error) {

	appealData := []postgres.Appeal{}
	queryString := p.appealByAppealChallengeIDQuery(tableName)
	err := p.db.Select(&appealData, queryString, appealChallengeID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving appeal from table: %v", err)
	}
	if len(appealData) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}
	appeal := appealData[0].DbToAppealData()
	return appeal, nil
}

func (p *PostgresPersister) appealsByChallengeIDsQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.Appeal{}, false, "")
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE original_challenge_id IN (?);", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) appealByAppealChallengeIDQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.Appeal{}, false, "")
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE appeal_challenge_id=$1;", fieldNames, tableName) // nolint: gosec
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
				return timestampInt, errors.WithMessagef(err, "no row in %s with timestamp. Error updating table", tableName)
			}
			return timestampInt, nil
		}
		return timestampInt, errors.WithMessage(err, "wasn't able to get listing from postgres table")
	}
	timestampInt, err = ctime.StringToTimestamp(timestampString)
	return timestampInt, err
}

func (p *PostgresPersister) lastEventHashesFromTable(tableName string) ([]string, error) {
	lastHashesString, err := p.typeExistsInCronTable(tableName, postgres.EventHashesDataType)
	if err != nil {
		noLastHash := []string{}
		if err == sql.ErrNoRows {
			// If row doesn't exist, create row with nil value
			updateErr := p.updateEventHashesInTable(noLastHash, tableName)
			if updateErr != nil {
				return noLastHash, errors.WithMessagef(err, "no row in %s with hash. Error updating table", tableName)
			}
			return noLastHash, nil
		}
		return noLastHash, errors.WithMessage(err, "wasn't able to get listing from postgres table")
	}
	return strings.Split(lastHashesString, ","), nil
}

func (p *PostgresPersister) updateCronTimestampInTable(timestamp int64, tableName string) error {
	cronData := postgres.NewCronData(ctime.TimestampToString(timestamp), postgres.TimestampDataType)
	return p.updateCronTable(cronData, tableName)
}

func (p *PostgresPersister) updateEventHashesInTable(eventHashes []string, tableName string) error {
	cronData := postgres.NewCronData(strings.Join(eventHashes, ","), postgres.EventHashesDataType)
	return p.updateCronTable(cronData, tableName)
}

func (p *PostgresPersister) updateCronTable(cronData *postgres.CronData, tableName string) error {
	typeExists := true
	_, err := p.typeExistsInCronTable(tableName, cronData.DataType)
	if err != nil {
		if err == sql.ErrNoRows {
			typeExists = false
		} else {
			return errors.WithMessage(err, "error checking DB for cron row")
		}
	}
	var queryString string
	if typeExists {
		updatedFields := []string{postgres.DataPersistedModelName}
		queryBuff, errBuff := p.updateDBQueryBuffer(updatedFields, tableName, postgres.CronData{})
		if errBuff != nil {
			return err
		}
		queryBuff.WriteString(" WHERE data_type=:data_type;") // nolint: gosec
		queryString = queryBuff.String()
	} else {
		queryString = p.insertIntoDBQueryString(tableName, postgres.CronData{})
	}
	_, err = p.db.NamedExec(queryString, cronData)
	if err != nil {
		return errors.Wrap(err, "error updating fields in db")
	}

	return nil
}

func (p *PostgresPersister) tokenTransfersByTxHashFromTable(txHash common.Hash, tableName string) (
	[]*model.TokenTransfer, error) {
	purchases := []*model.TokenTransfer{}
	queryString := p.tokenTransfersByTxHashQuery(tableName)

	blockDataValue := fmt.Sprintf("{ \"txHash\": \"%s\" }", txHash.Hex())
	dbPurchases := []*postgres.TokenTransfer{}
	err := p.db.Select(&dbPurchases, queryString, blockDataValue)
	if err != nil {
		return purchases, errors.Wrap(err, "error retrieving token transfers from table")
	}

	if len(dbPurchases) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}

	for _, dbPurchase := range dbPurchases {
		purchases = append(purchases, dbPurchase.DbToTokenTransfer())
	}

	return purchases, nil
}

func (p *PostgresPersister) tokenTransfersByTxHashQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.TokenTransfer{}, false, "")
	queryString := fmt.Sprintf( // nolint: gosec
		"SELECT %s FROM %s WHERE block_data @> $1 ORDER BY transfer_date",
		fieldNames,
		tableName,
	)
	return queryString
}

func (p *PostgresPersister) tokenTransfersByToAddressFromTable(addr common.Address,
	tableName string) ([]*model.TokenTransfer, error) {
	purchases := []*model.TokenTransfer{}
	queryString := p.tokenTransfersByToAddressQuery(tableName)

	dbPurchases := []*postgres.TokenTransfer{}
	err := p.db.Select(&dbPurchases, queryString, addr.Hex())
	if err != nil {
		return purchases, errors.Wrap(err, "error retrieving token transfers from table")
	}

	if len(dbPurchases) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}

	for _, dbPurchase := range dbPurchases {
		purchases = append(purchases, dbPurchase.DbToTokenTransfer())
	}

	return purchases, nil
}

func (p *PostgresPersister) tokenTransfersByToAddressQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.TokenTransfer{}, false, "")
	queryString := fmt.Sprintf( // nolint: gosec
		"SELECT %s FROM %s WHERE to_address = $1 ORDER BY transfer_date;",
		fieldNames,
		tableName,
	)
	return queryString
}

func (p *PostgresPersister) createTokenTransferInTable(purchase *model.TokenTransfer,
	tableName string) error {
	dbPurchase := postgres.NewTokenTransfer(purchase)
	queryString := p.insertIntoDBQueryString(tableName, postgres.TokenTransfer{})
	_, err := p.db.NamedExec(queryString, dbPurchase)
	if err != nil {
		return errors.Wrap(err, "error saving token transfer to table")
	}
	return nil
}

func (p *PostgresPersister) createParameterProposalInTable(paramProposal *model.ParameterProposal,
	tableName string) error {
	dbParamProposal := postgres.NewParameterProposal(paramProposal)
	queryString := p.insertIntoDBQueryString(tableName, postgres.ParameterProposal{})
	_, err := p.db.NamedExec(queryString, dbParamProposal)
	if err != nil {
		return fmt.Errorf("Error saving parameter proposal to table: %v", err)
	}
	return nil
}

func (p *PostgresPersister) paramProposalByPropIDFromTable(propID [32]byte,
	tableName string) (*model.ParameterProposal, error) {
	paramProposalData := []postgres.ParameterProposal{}
	queryString := p.paramProposalQuery(tableName)
	propIDString := cbytes.Byte32ToHexString(propID)
	err := p.db.Select(&paramProposalData, queryString, propIDString)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving parameter proposal from table: %v", err)
	}
	if len(paramProposalData) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}
	paramProposal, err := paramProposalData[0].DbToParameterProposalData()
	if err != nil {
		return nil, err
	}
	return paramProposal, nil
}

func (p *PostgresPersister) paramProposalByNameFromTable(name string,
	active bool, tableName string) ([]*model.ParameterProposal, error) {

	paramProposalData := []postgres.ParameterProposal{}
	queryString := p.paramProposalQueryByName(tableName, active)
	err := p.db.Select(&paramProposalData, queryString, name)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving parameter proposals from table: %v", err)
	}

	if len(paramProposalData) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}

	paramProposals := make([]*model.ParameterProposal, len(paramProposalData))

	for index, dbProp := range paramProposalData {
		modelProp, err := dbProp.DbToParameterProposalData()
		if err != nil {
			return nil, err
		}
		paramProposals[index] = modelProp
	}

	return paramProposals, nil
}

func (p *PostgresPersister) updateParamProposalInTable(paramProposal *model.ParameterProposal,
	updatedFields []string, tableName string) error {

	paramProposal.SetLastUpdatedDateTs(ctime.CurrentEpochSecsInInt64())
	updatedFields = append(updatedFields, lastUpdatedDateDBModelName)

	queryString, err := p.updateParamProposalQuery(updatedFields, tableName)
	if err != nil {
		return fmt.Errorf("Error creating query string for update: %v ", err)
	}
	dbParamProposal := postgres.NewParameterProposal(paramProposal)

	result, err := p.db.NamedExec(queryString, dbParamProposal)
	if err != nil {
		return fmt.Errorf("Error updating fields in db: %v", err)
	}
	err = p.checkUpdateRowsAffected(result)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresPersister) updateParamProposalQuery(updatedFields []string, tableName string) (string, error) {
	queryString, err := p.updateDBQueryBuffer(updatedFields, tableName, postgres.ParameterProposal{})
	if err != nil {
		return "", err
	}
	queryString.WriteString(" WHERE prop_id=:prop_id;") // nolint: gosec
	return queryString.String(), nil
}

func (p *PostgresPersister) paramProposalQuery(tableName string) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.ParameterProposal{}, false, "")
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE prop_id=$1", fieldNames, tableName) // nolint: gosec
	return queryString
}

func (p *PostgresPersister) paramProposalQueryByName(tableName string, active bool) string {
	fieldNames, _ := cpostgres.StructFieldsForQuery(postgres.ParameterProposal{}, false, "")
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE name=$1", fieldNames, tableName) // nolint: gosec
	if active {
		queryString = fmt.Sprintf("%s AND expired=false;", queryString)
	}
	return queryString
}

func (p *PostgresPersister) createUserChallengeDataInTable(userChallengeData *model.UserChallengeData,
	tableName string) error {
	dbUserChall := postgres.NewUserChallengeData(userChallengeData)
	queryString := p.insertIntoDBQueryString(tableName, postgres.UserChallengeData{})
	_, err := p.db.NamedExec(queryString, dbUserChall)
	if err != nil {
		return fmt.Errorf("Error saving UserChallengData to table: %v", err)
	}
	return nil
}

func (p *PostgresPersister) userChallengeDataByCriteriaFromTable(criteria *model.UserChallengeDataCriteria,
	tableName string) ([]*model.UserChallengeData, error) {
	dbUserChalls := []postgres.UserChallengeData{}
	queryString, err := p.userChallengeDataByCriteriaQuery(criteria, tableName)

	if err != nil {
		return nil, fmt.Errorf("Error writing query: %v", err)
	}
	nstmt, err := p.db.PrepareNamed(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error preparing query with sqlx: %v", err)
	}
	err = nstmt.Select(&dbUserChalls, criteria)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving listings from table: %v", err)
	}

	if len(dbUserChalls) == 0 {
		return nil, cpersist.ErrPersisterNoResults
	}
	userChalls := make([]*model.UserChallengeData, len(dbUserChalls))

	for index, dbUserChall := range dbUserChalls {
		modelUserChall := dbUserChall.DbToUserChallengeData()
		userChalls[index] = modelUserChall
	}
	return userChalls, nil
}

func (p *PostgresPersister) userChallengeDataByCriteriaQuery(criteria *model.UserChallengeDataCriteria,
	tableName string) (string, error) {
	queryBuf := bytes.NewBufferString("SELECT ") // nolint: gosec

	var fieldNames string
	fieldNames, _ = cpostgres.StructFieldsForQuery(postgres.UserChallengeData{}, false, "u")

	queryBuf.WriteString(fieldNames) // nolint: gosec
	queryBuf.WriteString(" FROM ")   // nolint: gosec
	queryBuf.WriteString(tableName)  // nolint: gosec
	queryBuf.WriteString(" u ")      // nolint: gosec

	if criteria.UserAddress != "" {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" u.user_address=:user_address") // nolint: gosec
	}
	if criteria.PollID > 0 {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" u.poll_id=:poll_id") // nolint: gosec
	}
	if criteria.CanUserReveal {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(fmt.Sprintf(" u.poll_reveal_end_date > %v", ctime.CurrentEpochSecsInInt64())) // nolint: gosec

	} else if criteria.CanUserRescue {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(" u.user_did_reveal=false AND u.did_user_rescue=false AND")                   // nolint: gosec
		queryBuf.WriteString(fmt.Sprintf(" u.poll_reveal_end_date < %v", ctime.CurrentEpochSecsInInt64())) // nolint: gosec
	} else if criteria.CanUserCollect {
		p.addWhereAnd(queryBuf)
		queryBuf.WriteString(` ((u.poll_is_passed = true AND u.choice = 1) OR (u.poll_is_passed = false AND u.choice = 0))
		AND (u.did_user_collect = false)`) // nolint: gosec
	}

	// NOTE(IS): We always only return latest votes
	p.addWhereAnd(queryBuf)
	queryBuf.WriteString(` u.latest_vote = true`) //nolint: gosec

	if criteria.Offset > 0 {
		queryBuf.WriteString(" OFFSET :offset") // nolint: gosec
	}

	if criteria.Count > 0 {
		queryBuf.WriteString(" LIMIT :count") // nolint: gosec
	}

	// NOTE(IS): default ordering by pollID
	queryBuf.WriteString(" ORDER BY u.poll_id") // nolint: gosec
	return queryBuf.String(), nil
}

func (p *PostgresPersister) updateUserChallengeDataInTable(userChallengeData *model.UserChallengeData,
	updatedFields []string, updateWithUserAddress bool, latestVote bool, tableName string) error {
	userChallengeData.SetLastUpdatedDateTs(ctime.CurrentEpochSecsInInt64())
	updatedFields = append(updatedFields, lastUpdatedDateDBModelName)
	queryString, err := p.updateUserChallengeDataQuery(updatedFields, tableName,
		updateWithUserAddress, latestVote)
	if err != nil {
		return fmt.Errorf("Error creating query string for update: %v ", err)
	}
	dbUserChallengeData := postgres.NewUserChallengeData(userChallengeData)

	result, err := p.db.NamedExec(queryString, dbUserChallengeData)
	if err != nil {
		return fmt.Errorf("Error updating fields in db: %v", err)
	}
	err = p.checkUpdateRowsAffected(result)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresPersister) updateUserChallengeDataQuery(updatedFields []string,
	tableName string, updateWithUserAddress bool, latestVote bool) (string, error) {
	queryString, err := p.updateDBQueryBuffer(updatedFields, tableName, postgres.UserChallengeData{})
	if err != nil {
		return "", err
	}
	if updateWithUserAddress {
		queryString.WriteString(" WHERE user_address=:user_address AND poll_id=:poll_id") // nolint: gosec
	} else {
		queryString.WriteString(" WHERE poll_id=:poll_id") // nolint: gosec
	}
	if latestVote {
		queryString.WriteString(" AND latest_vote=true;") //nolint: gosec
	}
	return queryString.String(), nil
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
		return "", errors.Errorf("There should not be more than 1 row with type %s in %s table", dataType, tableName)
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

func (p *PostgresPersister) checkUpdateRowsAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error updating checking affected rows in db")
	}
	if affected <= 0 {
		return ErrNoRowsAffected
	}
	return nil
}
