package discourse

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	defaultTableName = "listing_discourse_map"
)

// ListingMap is the model definition for a listing to discourse mapping table
// in postgresql.
type ListingMap struct {
	ListingAddress string `gorm:"primary_key"`
	TopicID        int64  `gorm:"not null;default: 0"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

// TableName returns the table name to use for the Discourse listing map
func (l *ListingMap) TableName() string {
	return defaultTableName
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
	// RetrieveListingMaps returns the listing discourse map for each item in a
	// slice of listing addresses
	RetrieveListingMaps(listingAddresses []string) ([]*ListingMap, error)
	// RetrieveListingMap returns the listing discourse map from a listing address
	RetrieveListingMap(listingAddress string) (*ListingMap, error)
	// SaveListingMap stores the given listing discourse map to the store.
	// Must use the listing address as the primary key.
	SaveListingMap(ldm *ListingMap) error
}
