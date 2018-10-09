// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var (
	// ErrPersisterNoResults is an error that represents no results from
	// the persister on queries.  Should be returned by the persisters
	// on event of no results in retrieval queries
	ErrPersisterNoResults = errors.New("No results from persister")
)

// errors must not be returned in valid conditions, such as when there is no
// record for a query.  In this case, return the empty value for the return
// type. errors must be reserved for actual internal errors.

// ListingCriteria contains the retrieval criteria for the ListingsByCriteria
// query.
type ListingCriteria struct {
	Offset          int   `db:"offset"`
	Count           int   `db:"count"`
	WhitelistedOnly bool  `db:"whitelisted_only"`
	CreatedFromTs   int64 `db:"created_fromts"`
	CreatedBeforeTs int64 `db:"created_beforets"`
}

// ListingPersister is the interface to store the listings data related to the processor
// and the aggregated data from the events.  Potentially to be used to service
// the APIs to pull data.
type ListingPersister interface {
	// Listings returns all listings by ListingCriteria
	ListingsByCriteria(criteria *ListingCriteria) ([]*Listing, error)
	// ListingsByAddress returns a slice of Listings based on addresses
	ListingsByAddresses(addresses []common.Address) ([]*Listing, error)
	// ListingByAddress retrieves listings based on addresses
	ListingByAddress(address common.Address) (*Listing, error)
	// CreateListing creates a new listing
	CreateListing(listing *Listing) error
	// UpdateListing updates fields on an existing listing
	UpdateListing(listing *Listing, updatedFields []string) error
	// DeleteListing removes a listing
	DeleteListing(listing *Listing) error
}

// ContentRevisionCriteria contains the retrieval criteria for a ContentRevisionsByCriteria
// query.
type ContentRevisionCriteria struct {
	ListingAddress string `db:"listing_address"`
	Offset         int    `db:"offset"`
	Count          int    `db:"count"`
	LatestOnly     bool   `db:"latest_only"`
	FromTs         int64  `db:"fromts"`
	BeforeTs       int64  `db:"beforets"`
}

// ContentRevisionPersister is the interface to store the content data related to the processor
// and the aggregated data from the events.  Potentially to be used to service
// the APIs to pull data.
type ContentRevisionPersister interface {
	// ContentRevisionsByCriteria returns all content revisions by ContentRevisionCriteria
	ContentRevisionsByCriteria(criteria *ContentRevisionCriteria) ([]*ContentRevision, error)
	// ContentRevisions retrieves the revisions for content on a listing
	ContentRevisions(address common.Address, contentID *big.Int) ([]*ContentRevision, error)
	// ContentRevision retrieves a specific content revision for newsroom content
	ContentRevision(address common.Address, contentID *big.Int, revisionID *big.Int) (*ContentRevision, error)
	// CreateContentRevision creates a new content revision
	CreateContentRevision(revision *ContentRevision) error
	// UpdateContentRevision updates fields on an existing content revision
	UpdateContentRevision(revision *ContentRevision, updatedFields []string) error
	// DeleteContentRevision removes a content revision
	DeleteContentRevision(revision *ContentRevision) error
}

// GovernanceEventCriteria contains the retrieval criteria for a GovernanceEventsByCriteria
// query.
type GovernanceEventCriteria struct {
	ListingAddress  string `db:"listing_address"`
	Offset          int    `db:"offset"`
	Count           int    `db:"count"`
	CreatedFromTs   int64  `db:"created_fromts"`
	CreatedBeforeTs int64  `db:"created_beforets"`
}

// GovernanceEventPersister is the interface to store the governance event data related to the processor
// and the aggregated data from the events.  Potentially to be used to service
// the APIs to pull data.
type GovernanceEventPersister interface {
	//GovernanceEventsByTxHash gets governance events based on txhash
	GovernanceEventsByTxHash(txHash common.Hash) ([]*GovernanceEvent, error)
	// GovernanceEventsByCriteria retrieves governance events based on criteria
	GovernanceEventsByCriteria(criteria *GovernanceEventCriteria) ([]*GovernanceEvent, error)
	// GovernanceEventsByListingAddress retrieves governance events based on listing address
	GovernanceEventsByListingAddress(address common.Address) ([]*GovernanceEvent, error)
	// GovernanceEventByChallengeID retrieves challenge by challengeID
	GovernanceEventByChallengeID(challengeID int) (*GovernanceEvent, error)
	// GovernanceEventsByChallengeIDs retrieves challenges by challengeIDs
	GovernanceEventsByChallengeIDs(challengeIDs []int) ([]*GovernanceEvent, error)
	// CreateGovernanceEvent creates a new governance event
	CreateGovernanceEvent(govEvent *GovernanceEvent) error
	// UpdateGovernanceEvent updates fields on an existing governance event
	UpdateGovernanceEvent(govEvent *GovernanceEvent, updatedFields []string) error
	// DeleteGovernanceEvent removes a governance event
	DeleteGovernanceEvent(govEvent *GovernanceEvent) error
}

// CronPersister persists information needed for the cron to run
type CronPersister interface {
	// TimestampOfLastEventForCron returns the timestamp for the last event seen by the processor
	TimestampOfLastEventForCron() (int64, error)
	// UpdateTimestampForCron updates the timestamp of the last event seen by the cron
	UpdateTimestampForCron(timestamp int64) error
}
