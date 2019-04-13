package postgres

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/go-common/pkg/numbers"

	"github.com/joincivil/civil-events-processor/pkg/model"
	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
)

const (
	defaultTokenTransferTableName = "token_transfer"
)

// CreateTokenTransferTableQuery returns the query to create the token_transfer table
func CreateTokenTransferTableQuery() string {
	return CreateTokenTransferTableQueryString(defaultTokenTransferTableName)
}

// CreateTokenTransferTableQueryString returns the query to create the token_transfer table
func CreateTokenTransferTableQueryString(tableName string) string {
	queryString := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s(
			to_address TEXT,
			from_address TEXT,
			amount NUMERIC,
			transfer_date INT,
			block_data JSONB
		);
	`, tableName)
	return queryString
}

// CreateTokenTransferEventTableIndices returns the query to create indices for this table
func CreateTokenTransferEventTableIndices() string {
	return CreateTokenTransferEventTableIndicesString(defaultTokenTransferTableName)
}

// CreateTokenTransferEventTableIndicesString returns the query to create indices for this table
func CreateTokenTransferEventTableIndicesString(tableName string) string {
	// queryString := fmt.Sprintf(`
	// `, tableName, tableName)
	// return queryString
	return ""
}

// NewTokenTransfer creates a new postgres TokenTransfer from model.TokenTransfer
func NewTokenTransfer(transfer *model.TokenTransfer) *TokenTransfer {
	dbTransfer := &TokenTransfer{}
	dbTransfer.ToAddress = transfer.ToAddress().Hex()
	dbTransfer.FromAddress = transfer.FromAddress().Hex()
	dbTransfer.Amount = numbers.BigIntToFloat64(transfer.Amount())
	dbTransfer.TransferDate = transfer.TransferDate()
	dbTransfer.BlockData = make(cpostgres.JsonbPayload)
	dbTransfer.fillBlockData(transfer.BlockData())
	return dbTransfer
}

// TokenTransfer is the postgres definition of a model.TokenTransfer
type TokenTransfer struct {
	ToAddress string `db:"to_address"`

	FromAddress string `db:"from_address"`

	Amount float64 `db:"amount"` // Amount in gwei, not token

	TransferDate int64 `db:"transfer_date"`

	BlockData cpostgres.JsonbPayload `db:"block_data"`
}

// DbToTokenTransfer creates a model.TokenTransfer from a postgres.TokenTransfer
func (t *TokenTransfer) DbToTokenTransfer() *model.TokenTransfer {
	params := &model.TokenTransferParams{}
	params.ToAddress = common.HexToAddress(t.ToAddress)
	params.FromAddress = common.HexToAddress(t.FromAddress)
	params.Amount = numbers.Float64ToBigInt(t.Amount)
	params.TransferDate = t.TransferDate

	params.BlockNumber = uint64(t.BlockData["blockNumber"].(float64))
	params.BlockHash = common.HexToHash(t.BlockData["blockHash"].(string))
	params.TxHash = common.HexToHash(t.BlockData["txHash"].(string))
	// NOTE: TxIndex is stored in DB as float64
	params.TxIndex = uint(t.BlockData["txIndex"].(float64))
	// NOTE: Index is stored in DB as float64
	params.Index = uint(t.BlockData["index"].(float64))

	return model.NewTokenTransfer(params)
}

func (t *TokenTransfer) fillBlockData(blockData model.BlockData) {
	t.BlockData["blockNumber"] = blockData.BlockNumber()
	t.BlockData["txHash"] = blockData.TxHash()
	t.BlockData["txIndex"] = blockData.TxIndex()
	t.BlockData["blockHash"] = blockData.BlockHash()
	t.BlockData["index"] = blockData.Index()
}
