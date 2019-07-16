package discourse

import "github.com/ethereum/go-ethereum/common"

// ListingMap is the model definition for a listing to discourse mapping table
// in postgresql.
type ListingMap struct {
	ListingAddress string `db:"listing_address"`
	TopicID        int64  `db:"topic_id"`
	CreatedTs      int64  `db:"created_ts"`
	UpdatedTs      int64  `db:"updated_ts"`
}

// ListingAddressAsAddr returns the listing address as a common.Address
func (l *ListingMap) ListingAddressAsAddr() common.Address {
	return common.HexToAddress(l.ListingAddress)
}

// AddrToListingAddress sets the listing address from a common.Address
func (l *ListingMap) AddrToListingAddress(a common.Address) {
	l.ListingAddress = a.Hex()
}

// ListingMapPersister defines the interface for a listing discourse map
// persister
type ListingMapPersister interface {
	// RetrieveListingMap returns the listing discourse map from a listing address
	RetrieveListingMap(listingAddress string) (*ListingMap, error)
	// SaveListingMap stores the given listing discourse map to the store.
	// Must use the listing address as the primary key.
	SaveListingMap(ldm *ListingMap) error
}
