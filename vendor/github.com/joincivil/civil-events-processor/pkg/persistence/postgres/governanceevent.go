package postgres

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	crawlerpostgres "github.com/joincivil/civil-events-crawler/pkg/persistence/postgres"
	"github.com/joincivil/civil-events-processor/pkg/model"
)

// CreateGovernanceEventTableQuery returns the query to create the governance_event table
func CreateGovernanceEventTableQuery() string {
	return CreateGovernanceEventTableQueryString("governance_event")
}

// CreateGovernanceEventTableQueryString returns the query to create this table
func CreateGovernanceEventTableQueryString(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            listing_address TEXT,
            sender_address TEXT,
            metadata JSONB,
            gov_event_type TEXT,
            creation_date BIGINT,
            last_updated BIGINT,
            event_hash TEXT
        );
    `, tableName)
	return queryString
}

// NewGovernanceEvent creates a new postgres GovernanceEvent
func NewGovernanceEvent(governanceEvent *model.GovernanceEvent) *GovernanceEvent {
	listingAddress := governanceEvent.ListingAddress().Hex()
	senderAddress := governanceEvent.SenderAddress().Hex()
	metadata := crawlerpostgres.JsonbPayload(governanceEvent.Metadata())
	return &GovernanceEvent{
		ListingAddress:      listingAddress,
		SenderAddress:       senderAddress,
		Metadata:            metadata,
		GovernanceEventType: governanceEvent.GovernanceEventType(),
		CreationDateTs:      governanceEvent.CreationDateTs(),
		LastUpdatedDateTs:   governanceEvent.LastUpdatedDateTs(),
		EventHash:           governanceEvent.EventHash(),
	}
}

// GovernanceEvent is postgres definition of model.GovernanceEvent
/// TODO (IS) : update with metadata params that are newly defined in processor
type GovernanceEvent struct {
	ListingAddress string `db:"listing_address"`

	SenderAddress string `db:"sender_address"`

	Metadata crawlerpostgres.JsonbPayload `db:"metadata"`

	GovernanceEventType string `db:"gov_event_type"`

	CreationDateTs int64 `db:"creation_date"`

	LastUpdatedDateTs int64 `db:"last_updated"`

	EventHash string `db:"event_hash"`
}

// DbToGovernanceData creates a model.GovernanceEvent from postgres.GovernanceEvent
func (ge *GovernanceEvent) DbToGovernanceData() *model.GovernanceEvent {
	listingAddress := common.HexToAddress(ge.ListingAddress)
	senderAddress := common.HexToAddress(ge.SenderAddress)
	metadata := model.Metadata(ge.Metadata)
	return model.NewGovernanceEvent(listingAddress, senderAddress, metadata, ge.GovernanceEventType,
		ge.CreationDateTs, ge.LastUpdatedDateTs, ge.EventHash)
}
