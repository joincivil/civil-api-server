package persistence

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/joincivil/civil-events-processor/pkg/model"
)

// NullPersister is a persister that does not save any values and always returns
// defaults for interface methods. Handy for testing and for one off use scenarios.
// Implements the ListingPersister, ContentRevisionPersister, and GovernanceEventPersister
type NullPersister struct{}

// ListingsByAddresses returns a slice of Listings based on addresses
func (n *NullPersister) ListingsByAddresses(addresses []common.Address) ([]*model.Listing, error) {
	return []*model.Listing{}, nil
}

// ListingByAddress retrieves listings based on addresses
func (n *NullPersister) ListingByAddress(address common.Address) (*model.Listing, error) {
	return &model.Listing{}, nil
}

// CreateListing creates a new listing
func (n *NullPersister) CreateListing(listing *model.Listing) error {
	return nil
}

// UpdateListing updates fields on an existing listing
func (n *NullPersister) UpdateListing(listing *model.Listing, updatedFields []string) error {
	return nil
}

// DeleteListing removes a listing
func (n *NullPersister) DeleteListing(listing *model.Listing) error {
	return nil
}

// ContentRevisions retrieves the revisions for content on a listing
func (n *NullPersister) ContentRevisions(address common.Address, contentID *big.Int) ([]*model.ContentRevision, error) {
	return []*model.ContentRevision{}, nil
}

// ContentRevision retrieves a specific content revision for newsroom content
func (n *NullPersister) ContentRevision(address common.Address, contentID *big.Int, revisionID *big.Int) (*model.ContentRevision, error) {
	return &model.ContentRevision{}, nil
}

// CreateContentRevision creates a new content revision
func (n *NullPersister) CreateContentRevision(revision *model.ContentRevision) error {
	return nil
}

// UpdateContentRevision updates fields on an existing content revision
func (n *NullPersister) UpdateContentRevision(revision *model.ContentRevision, updatedFields []string) error {
	return nil
}

// DeleteContentRevision removes a content revision
func (n *NullPersister) DeleteContentRevision(revision *model.ContentRevision) error {
	return nil
}

// GovernanceEventsByListingAddress retrieves governance events based on criteria
func (n *NullPersister) GovernanceEventsByListingAddress(address common.Address) ([]*model.GovernanceEvent, error) {
	return []*model.GovernanceEvent{}, nil
}

// CreateGovernanceEvent creates a new governance event
func (n *NullPersister) CreateGovernanceEvent(govEvent *model.GovernanceEvent) error {
	return nil
}

// UpdateGovernanceEvent updates fields on an existing governance event
func (n *NullPersister) UpdateGovernanceEvent(govEvent *model.GovernanceEvent, updatedFields []string) error {
	return nil
}

// DeleteGovenanceEvent removes a governance event
func (n *NullPersister) DeleteGovenanceEvent(govEvent *model.GovernanceEvent) error {
	return nil
}
