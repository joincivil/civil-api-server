package postgres

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-events-processor/pkg/model"
	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
)

const (
	// GovernanceEventTableBaseName is the base name of the table this code defines
	GovernanceEventTableBaseName = "governance_event"
)

// CreateGovernanceEventTableQuery returns the query to create this table
func CreateGovernanceEventTableQuery(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            listing_address TEXT,
            metadata JSONB,
            gov_event_type TEXT,
            creation_date INT,
            last_updated_timestamp INT,
            event_hash TEXT UNIQUE,
            block_data JSONB
        );
    `, tableName)
	return queryString
}

// CreateGovernanceEventTableIndicesQuery returns the query to create indices for this table
func CreateGovernanceEventTableIndicesQuery(tableName string) string {
	queryString := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS govevent_addr_idx ON %s (listing_address);
		CREATE INDEX IF NOT EXISTS govevent_block_data_idx ON %s USING GIN (block_data);
	`, tableName, tableName)
	return queryString
}

// NewGovernanceEvent creates a new postgres GovernanceEvent
func NewGovernanceEvent(governanceEvent *model.GovernanceEvent) *GovernanceEvent {
	govEvent := &GovernanceEvent{}
	govEvent.ListingAddress = governanceEvent.ListingAddress().Hex()
	govEvent.Metadata = cpostgres.JsonbPayload(governanceEvent.Metadata())
	govEvent.GovernanceEventType = governanceEvent.GovernanceEventType()
	govEvent.CreationDateTs = governanceEvent.CreationDateTs()
	govEvent.LastUpdatedDateTs = governanceEvent.LastUpdatedDateTs()
	govEvent.EventHash = governanceEvent.EventHash()
	govEvent.BlockData = make(cpostgres.JsonbPayload)
	govEvent.fillBlockData(governanceEvent.BlockData())
	return govEvent
}

// GovernanceEvent is postgres definition of model.GovernanceEvent
type GovernanceEvent struct {
	ListingAddress string `db:"listing_address"`

	Metadata cpostgres.JsonbPayload `db:"metadata"`

	GovernanceEventType string `db:"gov_event_type"`

	CreationDateTs int64 `db:"creation_date"`

	LastUpdatedDateTs int64 `db:"last_updated_timestamp"`

	EventHash string `db:"event_hash"`

	BlockData cpostgres.JsonbPayload `db:"block_data"`
}

// DbToGovernanceData creates a model.GovernanceEvent from postgres.GovernanceEvent
// NOTE: jsonb payloads are stored in DB as map[string]interface{}, Postgres converts some fields, see notes in function.
func (ge *GovernanceEvent) DbToGovernanceData() *model.GovernanceEvent {
	listingAddress := common.HexToAddress(ge.ListingAddress)
	metadata := model.Metadata(ge.Metadata)
	// NOTE: BlockNumber is stored in DB as float64
	blockNumber := uint64(ge.BlockData["blockNumber"].(float64))
	txHash := common.HexToHash(ge.BlockData["txHash"].(string))
	// NOTE: TxIndex is stored in DB as float64
	txIndex := uint(ge.BlockData["txIndex"].(float64))
	blockHash := common.HexToHash(ge.BlockData["blockHash"].(string))
	// NOTE: Index is stored in DB as float64
	index := uint(ge.BlockData["index"].(float64))
	return model.NewGovernanceEvent(listingAddress, metadata, ge.GovernanceEventType, ge.CreationDateTs,
		ge.LastUpdatedDateTs, ge.EventHash, blockNumber, txHash, txIndex, blockHash, index)
}

func (ge *GovernanceEvent) fillBlockData(blockData model.BlockData) {
	ge.BlockData["blockNumber"] = blockData.BlockNumber()
	ge.BlockData["txHash"] = blockData.TxHash()
	ge.BlockData["txIndex"] = blockData.TxIndex()
	ge.BlockData["blockHash"] = blockData.BlockHash()
	ge.BlockData["index"] = blockData.Index()
}
