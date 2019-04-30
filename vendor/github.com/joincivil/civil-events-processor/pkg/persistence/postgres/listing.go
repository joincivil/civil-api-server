package postgres // import "github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

import (
	"fmt"
	"math/big"
	"strconv"

	log "github.com/golang/glog"

	"github.com/ethereum/go-ethereum/common"

	"github.com/joincivil/civil-events-processor/pkg/model"
	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
	cstrings "github.com/joincivil/go-common/pkg/strings"
)

const (
	// ListingTableBaseName is the type of table this code defines
	ListingTableBaseName = "listing"
	nilChallengeID       = int64(-1)
)

// CreateListingTableQuery returns the query to create the listing table
func CreateListingTableQuery(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            name TEXT,
            contract_address TEXT PRIMARY KEY,
            whitelisted BOOL,
            last_governance_state BIGINT,
            url TEXT,
            charter JSONB,
            owner_addresses TEXT,
            owner TEXT,
            contributor_addresses TEXT,
            creation_timestamp INT,
            application_timestamp INT,
            approval_timestamp INT,
            last_updated_timestamp INT,
            app_expiry INT,
            challenge_id INT,
            unstaked_deposit NUMERIC
        );
    `, tableName)
	return queryString
}

// CreateListingTableIndicesQuery returns the query to create indices for this table
func CreateListingTableIndicesQuery(tableName string) string {
	queryString := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS listing_whitelisted_type_idx ON %s (whitelisted);
		CREATE INDEX IF NOT EXISTS listing_creation_timestamp_idx ON %s (creation_timestamp);
	`, tableName, tableName)
	return queryString
}

// Listing is the model definition for listing table in crawler db
// NOTE(IS) : golang<->postgres doesn't support list of strings. for now, OwnerAddresses and ContributorAddresses
// will be strings
type Listing struct {
	Name string `db:"name"`

	ContractAddress string `db:"contract_address"`

	Whitelisted bool `db:"whitelisted"`

	LastGovernanceState int `db:"last_governance_state"`

	URL string `db:"url"`

	Charter cpostgres.JsonbPayload `db:"charter"`

	// OwnerAddresses is a comma delimited string
	OwnerAddresses string `db:"owner_addresses"`

	Owner string `db:"owner"`

	ContributorAddresses string `db:"contributor_addresses"`

	CreatedDateTs int64 `db:"creation_timestamp"`

	ApplicationDateTs int64 `db:"application_timestamp"`

	ApprovalDateTs int64 `db:"approval_timestamp"`

	LastUpdatedDateTs int64 `db:"last_updated_timestamp"`

	AppExpiry int64 `db:"app_expiry"`

	UnstakedDeposit float64 `db:"unstaked_deposit"`

	ChallengeID int64 `db:"challenge_id"`
}

// NewListing constructs a listing for DB from a model.Listing
func NewListing(listing *model.Listing) *Listing {
	ownerAddresses := cstrings.ListCommonAddressesToString(listing.OwnerAddresses())
	contributorAddresses := cstrings.ListCommonAddressesToString(listing.ContributorAddresses())
	lastGovernanceState := int(listing.LastGovernanceState())
	owner := listing.Owner().Hex()

	var appExpiry int64
	var unstakedDeposit float64
	var challengeID int64
	if listing.AppExpiry() != nil {
		appExpiry = listing.AppExpiry().Int64()
	}
	if listing.UnstakedDeposit() != nil {
		f := new(big.Float).SetInt(listing.UnstakedDeposit())
		unstakedDeposit, _ = f.Float64()
	}
	if listing.ChallengeID() != nil {
		challengeID = listing.ChallengeID().Int64()
	} else {
		challengeID = nilChallengeID
	}
	charter := cpostgres.JsonbPayload(listing.Charter().AsMap())

	return &Listing{
		Name:                 listing.Name(),
		ContractAddress:      listing.ContractAddress().Hex(),
		Whitelisted:          listing.Whitelisted(),
		LastGovernanceState:  lastGovernanceState,
		URL:                  listing.URL(),
		Charter:              charter,
		OwnerAddresses:       ownerAddresses,
		Owner:                owner,
		ContributorAddresses: contributorAddresses,
		CreatedDateTs:        listing.CreatedDateTs(),
		ApplicationDateTs:    listing.ApplicationDateTs(),
		ApprovalDateTs:       listing.ApprovalDateTs(),
		LastUpdatedDateTs:    listing.LastUpdatedDateTs(),
		AppExpiry:            appExpiry,
		UnstakedDeposit:      unstakedDeposit,
		ChallengeID:          challengeID,
	}
}

// DbToListingData creates a model.Listing from postgres Listing
func (l *Listing) DbToListingData() *model.Listing {
	contractAddress := common.HexToAddress(l.ContractAddress)
	governanceState := model.GovernanceState(l.LastGovernanceState)
	ownerAddresses := cstrings.StringToCommonAddressesList(l.OwnerAddresses)
	contributorAddresses := cstrings.StringToCommonAddressesList(l.ContributorAddresses)
	ownerAddress := common.HexToAddress(l.Owner)
	appExpiry := big.NewInt(l.AppExpiry)
	unstakedDeposit := new(big.Int)
	unstakedDeposit.SetString(strconv.FormatFloat(l.UnstakedDeposit, 'f', -1, 64), 10)

	challengeID := big.NewInt(l.ChallengeID)

	charter := &model.Charter{}
	if len(l.Charter) != 0 {
		err := charter.FromMap(l.Charter)
		if err != nil {
			log.Errorf("Error decoding map to charter: err: %v", err)
		}
	} else {
		charter = nil
	}

	testListingParams := &model.NewListingParams{
		Name:                 l.Name,
		ContractAddress:      contractAddress,
		Whitelisted:          l.Whitelisted,
		LastState:            governanceState,
		URL:                  l.URL,
		Charter:              charter,
		Owner:                ownerAddress,
		OwnerAddresses:       ownerAddresses,
		ContributorAddresses: contributorAddresses,
		CreatedDateTs:        l.CreatedDateTs,
		ApplicationDateTs:    l.ApplicationDateTs,
		ApprovalDateTs:       l.ApprovalDateTs,
		LastUpdatedDateTs:    l.LastUpdatedDateTs,
		AppExpiry:            appExpiry,
		UnstakedDeposit:      unstakedDeposit,
		ChallengeID:          challengeID,
	}
	return model.NewListing(testListingParams)
}
