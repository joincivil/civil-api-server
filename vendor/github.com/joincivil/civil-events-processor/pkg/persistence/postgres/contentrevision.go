package postgres // import "github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-events-processor/pkg/model"
	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
)

const (
	// ContentRevisionTableBaseName is the type of table this code defines
	ContentRevisionTableBaseName = "content_revision"
)

// CreateContentRevisionTableQuery returns the query to create this table
func CreateContentRevisionTableQuery(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            id SERIAL PRIMARY KEY,
            listing_address TEXT,
            article_payload JSONB,
            article_payload_hash TEXT,
            editor_address TEXT,
            contract_content_id BIGINT,
            contract_revision_id BIGINT,
            revision_uri TEXT,
            revision_timestamp INT
        );
    `, tableName)
	return queryString
}

// CreateContentRevisionTableIndicesQuery returns the query to create indices for this table
func CreateContentRevisionTableIndicesQuery(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE INDEX IF NOT EXISTS revision_addr_type_idx ON %s (listing_address);
    `, tableName)
	return queryString
}

// ContentRevision is the model for content_revision table in db
// Make IDs strings?
type ContentRevision struct {
	ListingAddress     string                 `db:"listing_address"`
	ArticlePayload     cpostgres.JsonbPayload `db:"article_payload"`
	ArticlePayloadHash string                 `db:"article_payload_hash"`
	EditorAddress      string                 `db:"editor_address"`
	ContractContentID  int64                  `db:"contract_content_id"`
	ContractRevisionID int64                  `db:"contract_revision_id"`
	RevisionURI        string                 `db:"revision_uri"`
	RevisionDateTs     int64                  `db:"revision_timestamp"`
}

// NewContentRevision constructs a content_revision for DB from a model.ContentRevision
func NewContentRevision(contentRevision *model.ContentRevision) *ContentRevision {
	listingAddress := contentRevision.ListingAddress().Hex()
	articlePayload := cpostgres.JsonbPayload(contentRevision.Payload())
	editorAddress := contentRevision.EditorAddress().Hex()
	contractContentID := contentRevision.ContractContentID().Int64()
	contractRevisionID := contentRevision.ContractRevisionID().Int64()
	return &ContentRevision{
		ListingAddress:     listingAddress,
		ArticlePayload:     articlePayload,
		ArticlePayloadHash: contentRevision.PayloadHash(),
		EditorAddress:      editorAddress,
		ContractContentID:  contractContentID,
		ContractRevisionID: contractRevisionID,
		RevisionURI:        contentRevision.RevisionURI(),
		RevisionDateTs:     contentRevision.RevisionDateTs(),
	}
}

// DbToContentRevisionData creates a model.ContentRevision from postgres ContentRevision
func (cr *ContentRevision) DbToContentRevisionData() *model.ContentRevision {
	listingAddress := common.HexToAddress(cr.ListingAddress)
	// TODO (IS): maybe should do a generic conversion of jsonb types back to map[string]interface{}
	payload := model.ArticlePayload(cr.ArticlePayload)
	editorAddress := common.HexToAddress(cr.EditorAddress)
	contractContentID := big.NewInt(cr.ContractContentID)
	contractRevisionID := big.NewInt(cr.ContractRevisionID)
	return model.NewContentRevision(listingAddress, payload, cr.ArticlePayloadHash, editorAddress, contractContentID,
		contractRevisionID, cr.RevisionURI, cr.RevisionDateTs)
}
