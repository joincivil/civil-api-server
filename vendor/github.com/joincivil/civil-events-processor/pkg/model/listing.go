// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// ResetChallengeIDEvents is the list of governance events that reset challengeID to 0
var ResetChallengeIDEvents = []GovernanceState{
	GovernanceStateAppWhitelisted,
	GovernanceStateRemoved,
	GovernanceStateAppRemoved}

// NewListing is a convenience function to initialize a new Listing struct
func NewListing(name string, contractAddress common.Address, whitelisted bool,
	lastState GovernanceState, url string, charterURI string, owner common.Address, ownerAddresses []common.Address,
	contributorAddresses []common.Address, createdDateTs int64, applicationDateTs int64,
	approvalDateTs int64, lastUpdatedDateTs int64, appExpiry *big.Int, unstakedDeposit *big.Int,
	challengeID *big.Int) *Listing {
	return &Listing{
		name:                 name,
		contractAddress:      contractAddress,
		whitelisted:          whitelisted,
		lastGovernanceState:  lastState,
		url:                  url,
		charterURI:           charterURI,
		owner:                owner,
		ownerAddresses:       ownerAddresses,
		contributorAddresses: contributorAddresses,
		createdDateTs:        createdDateTs,
		applicationDateTs:    applicationDateTs,
		approvalDateTs:       approvalDateTs,
		lastUpdatedDateTs:    lastUpdatedDateTs,
		appExpiry:            appExpiry,
		unstakedDeposit:      unstakedDeposit,
		challengeID:          challengeID,
	}
}

// Listing represents a newsroom listing in the Civil TCR
type Listing struct {
	name string

	contractAddress common.Address

	whitelisted bool

	lastGovernanceState GovernanceState

	url string

	charterURI string // Updated to reflect how we are storing the charter

	owner common.Address

	ownerAddresses []common.Address

	contributorAddresses []common.Address

	createdDateTs int64

	applicationDateTs int64

	approvalDateTs int64

	lastUpdatedDateTs int64

	appExpiry *big.Int

	unstakedDeposit *big.Int

	challengeID *big.Int
}

// Name returns the newsroom name
func (l *Listing) Name() string {
	return l.name
}

// SetName sets the value for name
func (l *Listing) SetName(name string) {
	l.name = name
}

// Whitelisted returns a bool to indicate if the newsroom is whitelisted
// and on the registry
func (l *Listing) Whitelisted() bool {
	return l.whitelisted
}

// SetWhitelisted sets the value for whitelisted field
func (l *Listing) SetWhitelisted(whitelisted bool) {
	l.whitelisted = whitelisted
}

// LastGovernanceStateString returns the last governance event on this newsroom
func (l *Listing) LastGovernanceStateString() string {
	return l.lastGovernanceState.String()
}

// LastGovernanceState returns the last governance event on this newsroom
func (l *Listing) LastGovernanceState() GovernanceState {
	return l.lastGovernanceState
}

// SetLastGovernanceState sets the value of the last governance state
func (l *Listing) SetLastGovernanceState(lastState GovernanceState) {
	l.lastGovernanceState = lastState
}

// ContractAddress returns the newsroom contract address
func (l *Listing) ContractAddress() common.Address {
	return l.contractAddress
}

// URL is the homepage of the newsroom
func (l *Listing) URL() string {
	return l.url
}

// CharterURI returns the URI to the charter post for the newsroom
func (l *Listing) CharterURI() string {
	return l.charterURI
}

// OwnerAddresses is the addresses of the owners of the newsroom - all members of multisig
func (l *Listing) OwnerAddresses() []common.Address {
	return l.ownerAddresses
}

// Owner is the address of the multisig owner of the newsroom
func (l *Listing) Owner() common.Address {
	return l.owner
}

// AddOwnerAddress adds another address to the list of owner addresses
func (l *Listing) AddOwnerAddress(addr common.Address) {
	l.ownerAddresses = append(l.ownerAddresses, addr)
}

// RemoveOwnerAddress removes an address from the list of owner addresses
func (l *Listing) RemoveOwnerAddress(addr common.Address) {
	numAddrs := len(l.ownerAddresses)
	if numAddrs <= 1 {
		l.ownerAddresses = []common.Address{}
		return
	}
	addrs := make([]common.Address, numAddrs-1)
	addrsIndex := 0
	for _, existingAddr := range l.ownerAddresses {
		if existingAddr != addr {
			addrs[addrsIndex] = existingAddr
			addrsIndex++
		}
	}
	l.ownerAddresses = addrs
}

// ContributorAddresses returns a list of contributor data to a newsroom
func (l *Listing) ContributorAddresses() []common.Address {
	return l.contributorAddresses
}

// AddContributorAddress adds another address to the list of contributor addresses
func (l *Listing) AddContributorAddress(addr common.Address) {
	l.contributorAddresses = append(l.contributorAddresses, addr)
}

// RemoveContributorAddress removes an address from the list of owner addresses
func (l *Listing) RemoveContributorAddress(addr common.Address) {
	addrs := make([]common.Address, len(l.contributorAddresses)-1)
	addrsIndex := 0
	for _, existingAddr := range l.contributorAddresses {
		if existingAddr != addr {
			addrs[addrsIndex] = existingAddr
			addrsIndex++
		}
	}
	l.contributorAddresses = addrs
}

// ApplicationDateTs returns the timestamp of the listing's initial application
func (l *Listing) ApplicationDateTs() int64 {
	return l.applicationDateTs
}

// ApprovalDateTs returns the timestamp of the listing's whitelisted/approved
func (l *Listing) ApprovalDateTs() int64 {
	return l.approvalDateTs
}

// SetApprovalDateTs sets the date of the last time this listing was whitelisted/approval
func (l *Listing) SetApprovalDateTs(date int64) {
	l.approvalDateTs = date
}

// LastUpdatedDateTs returns the timestamp of the last update to the listing
func (l *Listing) LastUpdatedDateTs() int64 {
	return l.lastUpdatedDateTs
}

// SetLastUpdatedDateTs sets the value of the last time this listing was updated
func (l *Listing) SetLastUpdatedDateTs(date int64) {
	l.lastUpdatedDateTs = date
}

// CreatedDateTs returns the timestamp of listing creation
func (l *Listing) CreatedDateTs() int64 {
	return l.createdDateTs
}

// AppExpiry returns the expiration date of the application to this newsroom
func (l *Listing) AppExpiry() *big.Int {
	return l.appExpiry
}

// UnstakedDeposit returns the unstaked deposit of the newsroom
func (l *Listing) UnstakedDeposit() *big.Int {
	return l.unstakedDeposit
}

// SetChallengeID sets the challenge ID of the listing.
func (l *Listing) SetChallengeID(id *big.Int) {
	l.challengeID = id
}

// ChallengeID returns the most recent challengeID of the listing
func (l *Listing) ChallengeID() *big.Int {
	return l.challengeID
}
