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
            creation_date INT,
            last_updated INT,
            event_hash TEXT,
            block_data JSONB
        );
    `, tableName)
	return queryString
}

// NewGovernanceEvent creates a new postgres GovernanceEvent
func NewGovernanceEvent(governanceEvent *model.GovernanceEvent) *GovernanceEvent {
	govEvent := &GovernanceEvent{}
	govEvent.ListingAddress = governanceEvent.ListingAddress().Hex()
	govEvent.SenderAddress = governanceEvent.SenderAddress().Hex()
	govEvent.Metadata = crawlerpostgres.JsonbPayload(governanceEvent.Metadata())
	govEvent.GovernanceEventType = governanceEvent.GovernanceEventType()
	govEvent.CreationDateTs = governanceEvent.CreationDateTs()
	govEvent.LastUpdatedDateTs = governanceEvent.LastUpdatedDateTs()
	govEvent.EventHash = governanceEvent.EventHash()
	govEvent.BlockData = make(crawlerpostgres.JsonbPayload)
	govEvent.fillBlockData(governanceEvent.BlockData())
	return govEvent
}

// GovernanceEvent is postgres definition of model.GovernanceEvent
type GovernanceEvent struct {
	ListingAddress string `db:"listing_address"`

	SenderAddress string `db:"sender_address"`

	Metadata crawlerpostgres.JsonbPayload `db:"metadata"`

	GovernanceEventType string `db:"gov_event_type"`

	CreationDateTs int64 `db:"creation_date"`

	LastUpdatedDateTs int64 `db:"last_updated"`

	EventHash string `db:"event_hash"`

	BlockData crawlerpostgres.JsonbPayload `db:"block_data"`
}

// DbToGovernanceData creates a model.GovernanceEvent from postgres.GovernanceEvent
// NOTE: jsonb payloads are stored in DB as map[string]interface{}, Postgres converts some fields, see notes in function.
func (ge *GovernanceEvent) DbToGovernanceData() *model.GovernanceEvent {
	listingAddress := common.HexToAddress(ge.ListingAddress)
	senderAddress := common.HexToAddress(ge.SenderAddress)
	metadata := model.Metadata(ge.Metadata)
	// NOTE: BlockNumber is stored in DB as float64
	blockNumber := uint64(ge.BlockData["blockNumber"].(float64))
	txHash := common.HexToHash(ge.BlockData["txHash"].(string))
	// NOTE: TxIndex is stored in DB as float64
	txIndex := uint(ge.BlockData["txIndex"].(float64))
	blockHash := common.HexToHash(ge.BlockData["blockHash"].(string))
	// NOTE: Index is stored in DB as float64
	index := uint(ge.BlockData["index"].(float64))
	return model.NewGovernanceEvent(listingAddress, senderAddress, metadata, ge.GovernanceEventType,
		ge.CreationDateTs, ge.LastUpdatedDateTs, ge.EventHash, blockNumber, txHash, txIndex, blockHash, index)
}

func (ge *GovernanceEvent) fillBlockData(blockData model.BlockData) {
	ge.BlockData["blockNumber"] = blockData.BlockNumber()
	ge.BlockData["txHash"] = blockData.TxHash()
	ge.BlockData["txIndex"] = blockData.TxIndex()
	ge.BlockData["blockHash"] = blockData.BlockHash()
	ge.BlockData["index"] = blockData.Index()
}
