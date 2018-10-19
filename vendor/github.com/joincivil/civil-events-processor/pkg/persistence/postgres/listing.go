package postgres // import "github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-events-processor/pkg/model"
	"math/big"
	"strconv"
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
            charter_uri TEXT,
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

	CharterURI string `db:"charter_uri"`

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
	return &Listing{
		Name:                 listing.Name(),
		ContractAddress:      listing.ContractAddress().Hex(),
		Whitelisted:          listing.Whitelisted(),
		LastGovernanceState:  lastGovernanceState,
		URL:                  listing.URL(),
		CharterURI:           listing.CharterURI(),
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
	listing := model.NewListing(l.Name, contractAddress, l.Whitelisted, governanceState, l.URL, l.CharterURI, ownerAddress,
		ownerAddresses, contributorAddresses, l.CreatedDateTs, l.ApplicationDateTs, l.ApprovalDateTs, l.LastUpdatedDateTs)
	listing.SetAppExpiry(appExpiry)
	listing.SetUnstakedDeposit(unstakedDeposit)
	listing.SetChallengeID(challengeID)
	return listing
}
