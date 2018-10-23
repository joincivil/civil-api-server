package postgres // import "github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

import (
	"fmt"
	log "github.com/golang/glog"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	crawlerpg "github.com/joincivil/civil-events-crawler/pkg/persistence/postgres"
	"github.com/joincivil/civil-events-processor/pkg/model"
)

// CreateListingTableQuery returns the query to create the listing table
func CreateListingTableQuery() string {
	return CreateListingTableQueryString("listing")
}

// CreateListingTableQueryString returns the query to create this table
func CreateListingTableQueryString(tableName string) string {
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

// Listing is the model definition for listing table in crawler db
// NOTE(IS) : golang<->postgres doesn't support list of strings. for now, OwnerAddresses and ContributorAddresses
// will be strings
type Listing struct {
	Name string `db:"name"`

	ContractAddress string `db:"contract_address"`

	Whitelisted bool `db:"whitelisted"`

	LastGovernanceState int `db:"last_governance_state"`

	URL string `db:"url"`

	Charter crawlerpg.JsonbPayload `db:"charter"`

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

	ChallengeID uint64 `db:"challenge_id"`
}

// NewListing constructs a listing for DB from a model.Listing
func NewListing(listing *model.Listing) *Listing {
	ownerAddresses := ListCommonAddressesToString(listing.OwnerAddresses())
	contributorAddresses := ListCommonAddressesToString(listing.ContributorAddresses())
	lastGovernanceState := int(listing.LastGovernanceState())
	owner := listing.Owner().Hex()

	var appExpiry int64
	var unstakedDeposit float64
	var challengeID uint64
	if listing.AppExpiry() != nil {
		appExpiry = listing.AppExpiry().Int64()
	}
	if listing.UnstakedDeposit() != nil {
		f := new(big.Float).SetInt(listing.UnstakedDeposit())
		unstakedDeposit, _ = f.Float64()
	}
	if listing.ChallengeID() != nil {
		challengeID = listing.ChallengeID().Uint64()
	}

	charter := crawlerpg.JsonbPayload(listing.Charter().AsMap())

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
	ownerAddresses := StringToCommonAddressesList(l.OwnerAddresses)
	contributorAddresses := StringToCommonAddressesList(l.ContributorAddresses)
	ownerAddress := common.HexToAddress(l.Owner)
	appExpiry := big.NewInt(l.AppExpiry)
	unstakedDeposit := new(big.Int)
	unstakedDeposit.SetString(strconv.FormatFloat(l.UnstakedDeposit, 'f', -1, 64), 10)

	challengeID := big.NewInt(0)
	challengeID.SetUint64(l.ChallengeID)

	charter := &model.Charter{}
	err := charter.FromMap(l.Charter)
	if err != nil {
		log.Errorf("Error decoding map to charter: err: ", err)
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
