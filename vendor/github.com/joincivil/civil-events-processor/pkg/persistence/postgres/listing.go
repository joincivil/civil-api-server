package postgres // import "github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
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
            charter_uri TEXT,
            owner_addresses TEXT,
            contributor_addresses TEXT,
            creation_timestamp BIGINT,
            application_timestamp BIGINT,
            approval_timestamp BIGINT,
            last_updated_timestamp BIGINT
        );
    `, tableName)
	return queryString
}

// Listing is the model definition for listing table in crawler db
// NOTE: bigint in postgres: -9223372036854775808 to +9223372036854775807
// uint64 in golang: 0 to 18446744073709551615
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

	ContributorAddresses string `db:"contributor_addresses"`

	CreatedDateTs int64 `db:"creation_timestamp"`

	ApplicationDateTs int64 `db:"application_timestamp"`

	ApprovalDateTs int64 `db:"approval_timestamp"`

	LastUpdatedDateTs int64 `db:"last_updated_timestamp"`
}

// NewListing constructs a listing for DB from a model.Listing
func NewListing(listing *model.Listing) *Listing {
	ownerAddresses := ListCommonAddressesToString(listing.OwnerAddresses())
	contributorAddresses := ListCommonAddressesToString(listing.ContributorAddresses())
	lastGovernanceState := int(listing.LastGovernanceState())
	return &Listing{
		Name:                 listing.Name(),
		ContractAddress:      listing.ContractAddress().Hex(),
		Whitelisted:          listing.Whitelisted(),
		LastGovernanceState:  lastGovernanceState,
		URL:                  listing.URL(),
		CharterURI:           listing.CharterURI(),
		OwnerAddresses:       ownerAddresses,
		ContributorAddresses: contributorAddresses,
		CreatedDateTs:        listing.CreatedDateTs(),
		ApplicationDateTs:    listing.ApplicationDateTs(),
		ApprovalDateTs:       listing.ApprovalDateTs(),
		LastUpdatedDateTs:    listing.LastUpdatedDateTs(),
	}
}

// DbToListingData creates a model.Listing from postgres Listing
func (l *Listing) DbToListingData() *model.Listing {
	contractAddress := common.HexToAddress(l.ContractAddress)
	governanceState := model.GovernanceState(l.LastGovernanceState)
	ownerAddresses := StringToCommonAddressesList(l.OwnerAddresses)
	contributorAddresses := StringToCommonAddressesList(l.ContributorAddresses)
	return model.NewListing(l.Name, contractAddress, l.Whitelisted, governanceState, l.URL, l.CharterURI,
		ownerAddresses, contributorAddresses, l.CreatedDateTs, l.ApplicationDateTs, l.ApprovalDateTs, l.LastUpdatedDateTs)
}
