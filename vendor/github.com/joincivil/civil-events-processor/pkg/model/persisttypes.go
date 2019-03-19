// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// errors must not be returned in valid conditions, such as when there is no
// record for a query.  In this case, return the empty value for the return
// type. errors must be reserved for actual internal errors.  Use ErrPersisterNoResults.

// ListingCriteria contains the retrieval criteria for the ListingsByCriteria
// query. Only one of WhitelistedOnly, RejectedOnly, ActiveChallenge, CurrentApplication can
// be true in one instance.
type ListingCriteria struct {
	Offset int `db:"offset"`
	Count  int `db:"count"`
	// Listings that are currently whitelisted, whitelisted = true
	WhitelistedOnly bool `db:"whitelisted_only"`
	// Listings that were challenged and rejected, they could have an active application.
	RejectedOnly bool `db:"rejected_only"`
	// Listings that have a challenge in progress.
	ActiveChallenge bool `db:"active_challenge"`
	// Listings that have a current application in progress, or listings that have passed their appExpiry
	// and have not been updated yet
	CurrentApplication bool  `db:"current_application"`
	CreatedFromTs      int64 `db:"created_fromts"`
	CreatedBeforeTs    int64 `db:"created_beforets"`
}

// ListingPersister is the interface to store the listings data related to the processor
// and the aggregated data from the events.  Potentially to be used to service
// the APIs to pull data.
type ListingPersister interface {
	// Listings returns all listings by ListingCriteria sorted by creation ts
	ListingsByCriteria(criteria *ListingCriteria) ([]*Listing, error)
	// ListingsByAddress returns a slice of Listings in order based on addresses
	ListingsByAddresses(addresses []common.Address) ([]*Listing, error)
	// ListingByAddress retrieves listings based on addresses
	ListingByAddress(address common.Address) (*Listing, error)
	// CreateListing creates a new listing
	CreateListing(listing *Listing) error
	// UpdateListing updates fields on an existing listing
	UpdateListing(listing *Listing, updatedFields []string) error
	// DeleteListing removes a listing
	DeleteListing(listing *Listing) error
	// Close shuts down the persister
	Close() error
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
	// Close shuts down the persister
	Close() error
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
	// CreateGovernanceEvent creates a new governance event
	CreateGovernanceEvent(govEvent *GovernanceEvent) error
	// UpdateGovernanceEvent updates fields on an existing governance event
	UpdateGovernanceEvent(govEvent *GovernanceEvent, updatedFields []string) error
	// DeleteGovernanceEvent removes a governance event
	DeleteGovernanceEvent(govEvent *GovernanceEvent) error
	// Close shuts down the persister
	Close() error
}

// CronPersister persists information needed for the cron to run
type CronPersister interface {
	// TimestampOfLastEventForCron returns the timestamp for the last event seen by the processor
	TimestampOfLastEventForCron() (int64, error)
	// UpdateTimestampForCron updates the timestamp of the last event seen by the cron
	UpdateTimestampForCron(timestamp int64) error
	// EventHashesOfLastTimestampForCron returns the event hashes processed for the last timestamp from cron
	EventHashesOfLastTimestampForCron() ([]string, error)
	// UpdateEventHashesForCron updates the eventHashes saved in cron table
	UpdateEventHashesForCron(eventHashes []string) error
	// Close shuts down the persister
	Close() error
}

// ChallengePersister is the interface to store ChallengeData
type ChallengePersister interface {
	// ChallengeByChallengeID gets a challenge by challengeID
	ChallengeByChallengeID(challengeID int) (*Challenge, error)
	// ChallengesByChallengeIDs returns a slice of challenges in order based on challenge IDs
	ChallengesByChallengeIDs(challengeIDs []int) ([]*Challenge, error)
	// ChallengesByListingAddress gets list of challenges for a listing sorted by
	// challenge id
	ChallengesByListingAddress(addr common.Address) ([]*Challenge, error)
	// ChallengesByListingAddresses gets slice of challenges in order by challenge ID
	// for a each listing address in order of addresses
	ChallengesByListingAddresses(addr []common.Address) ([][]*Challenge, error)
	// CreateChallenge creates a new challenge
	CreateChallenge(challenge *Challenge) error
	// UpdateChallenge updates a challenge
	UpdateChallenge(challenge *Challenge, updatedFields []string) error
	// Close shuts down the persister
	Close() error
}

// PollPersister is the interface to store PollData
type PollPersister interface {
	// PollByPollID gets a poll by pollID
	PollByPollID(pollID int) (*Poll, error)
	// PollsByPollIDs returns a slice of polls in order based on poll IDs
	PollsByPollIDs(pollIDs []int) ([]*Poll, error)
	// CreatePoll creates a new poll
	CreatePoll(poll *Poll) error
	// UpdatePoll updates a poll
	UpdatePoll(poll *Poll, updatedFields []string) error
	// Close shuts down the persister
	Close() error
}

// AppealPersister is the interface to store AppealData
type AppealPersister interface {
	// AppealByChallengeID gets an appeal by challengeID
	AppealByChallengeID(challengeID int) (*Appeal, error)
	// AppealsByChallengeIDs returns a slice of appeals in order based on challenge IDs
	AppealsByChallengeIDs(challengeIDs []int) ([]*Appeal, error)
	// CreateAppeal creates a new appeal
	CreateAppeal(appeal *Appeal) error
	// UpdateAppeal updates an appeal
	UpdateAppeal(appeal *Appeal, updatedFields []string) error
	// Close shuts down the persister
	Close() error
}

// TokenTransferPersister is the persister interface to store TokenTransfer
type TokenTransferPersister interface {
	// TokenTransfersByToAddress gets a list of token transfers by purchaser address
	TokenTransfersByToAddress(addr common.Address) ([]*TokenTransfer, error)
	// CreateTokenTransfer creates a new token transfer
	CreateTokenTransfer(purchase *TokenTransfer) error
	// Close shuts down the persister
	Close() error
}
