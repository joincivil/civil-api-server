// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"github.com/joincivil/civil-events-processor/pkg/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// ResetListingFieldsEvents is the list of governance events emitted that reset fields in the Listing
// model to 0 when UpdateStatus() is called
// NOTE(IS): This is just a union of ResetChallengeIDEvents and ResetAppExpiryEvents
var ResetListingFieldsEvents = []GovernanceState{
	GovernanceStateAppWhitelisted,
	GovernanceStateRemoved,
	GovernanceStateAppRemoved,
}

// ResetChallengeIDEvents is the list of governance events that reset challengeID to 0
var ResetChallengeIDEvents = []GovernanceState{
	GovernanceStateAppWhitelisted,
	GovernanceStateRemoved,
	GovernanceStateAppRemoved,
}

// ResetAppExpiryEvents is the list of governance events that reset appExpiry to 0
var ResetAppExpiryEvents = []GovernanceState{
	GovernanceStateRemoved,
	GovernanceStateAppRemoved,
}

// CharterParams are params to create a new populated charter struct via
// NewCharter
type CharterParams struct {
	URI         string
	ContentID   *big.Int
	RevisionID  *big.Int
	Signature   []byte
	Author      common.Address
	ContentHash [32]byte
	Timestamp   *big.Int
}

// NewCharter is a convenience func to create a new Charter struct
func NewCharter(params *CharterParams) *Charter {
	return &Charter{
		uri:         params.URI,
		contentID:   params.ContentID,
		revisionID:  params.RevisionID,
		signature:   params.Signature,
		author:      params.Author,
		contentHash: params.ContentHash,
		timestamp:   params.Timestamp,
	}
}

const (
	uriMapStr         = "uri"
	contentIDMapStr   = "content_id"
	revisionIDMapStr  = "revision_id"
	signatureMapStr   = "signature"
	authorMapStr      = "author"
	contentHashMapStr = "content_hash"
	timestampMapStr   = "timestamp"
)

// Charter represents data for the newsroom/listing charter
type Charter struct {
	uri string

	contentID *big.Int

	revisionID *big.Int

	signature []byte

	author common.Address

	contentHash [32]byte

	timestamp *big.Int
}

// URI returns the charter uri
func (c *Charter) URI() string {
	return c.uri
}

// ContentID returns the charter content id
func (c *Charter) ContentID() *big.Int {
	return c.contentID
}

// RevisionID returns the charter revision id
func (c *Charter) RevisionID() *big.Int {
	return c.revisionID
}

// Signature returns the charter signature
func (c *Charter) Signature() []byte {
	return c.signature
}

// Author returns the charter author address
func (c *Charter) Author() common.Address {
	return c.author
}

// ContentHash returns the charter content hash
func (c *Charter) ContentHash() [32]byte {
	return c.contentHash
}

// Timestamp returns the charter timestamp
func (c *Charter) Timestamp() *big.Int {
	return c.timestamp
}

// AsMap returns the charter data as a map[string]interface{}
func (c *Charter) AsMap() map[string]interface{} {
	newMap := map[string]interface{}{}
	newMap[uriMapStr] = c.uri
	if c.contentID != nil {
		newMap[contentIDMapStr] = c.contentID.Int64()
	}
	if c.revisionID != nil {
		newMap[revisionIDMapStr] = c.revisionID.Int64()
	}
	newMap[signatureMapStr] = string(c.signature)
	newMap[authorMapStr] = c.author.Hex()
	newMap[contentHashMapStr] = utils.Byte32ToHexString(c.contentHash)

	if c.timestamp != nil {
		newMap[timestampMapStr] = c.timestamp.Int64()
	}
	return newMap
}

// FromMap converts the charter data from map[string]interface{} to
// charter data
func (c *Charter) FromMap(charterMap map[string]interface{}) error {
	val, ok := charterMap[uriMapStr]
	if ok && val != nil {
		c.uri = val.(string)
	}
	val, ok = charterMap[contentIDMapStr]
	if ok && val != nil {
		switch t := val.(type) {
		case int64:
			c.contentID = big.NewInt(t)
		case float64:
			c.contentID = big.NewInt(int64(t))
		}
	}
	val, ok = charterMap[revisionIDMapStr]
	if ok && val != nil {
		switch t := val.(type) {
		case int64:
			c.revisionID = big.NewInt(t)
		case float64:
			c.revisionID = big.NewInt(int64(t))
		}
	}
	val, ok = charterMap[signatureMapStr]
	if ok && val != nil {
		c.signature = []byte(val.(string))
	}
	val, ok = charterMap[authorMapStr]
	if ok && val != nil {
		c.author = common.HexToAddress(val.(string))
	}
	val, ok = charterMap[contentHashMapStr]
	if ok && val != nil {
		fixed, err := utils.HexStringToByte32(val.(string))
		if err != nil {
			return err
		}
		c.contentHash = fixed
	}
	val, ok = charterMap[timestampMapStr]
	if ok && val != nil {
		switch t := val.(type) {
		case int64:
			c.timestamp = big.NewInt(t)
		case float64:
			c.timestamp = big.NewInt(int64(t))
		}
	}
	return nil
}

// NewListingParams represents all the necessary data to create a new listing
// using NewListing
type NewListingParams struct {
	Name                 string
	ContractAddress      common.Address
	Whitelisted          bool
	LastState            GovernanceState
	URL                  string
	Charter              *Charter
	Owner                common.Address
	OwnerAddresses       []common.Address
	ContributorAddresses []common.Address
	CreatedDateTs        int64
	ApplicationDateTs    int64
	ApprovalDateTs       int64
	LastUpdatedDateTs    int64
	AppExpiry            *big.Int
	UnstakedDeposit      *big.Int
	ChallengeID          *big.Int
}

// NewListing is a convenience function to initialize a new Listing struct
func NewListing(params *NewListingParams) *Listing {
	return &Listing{
		name:                 params.Name,
		contractAddress:      params.ContractAddress,
		whitelisted:          params.Whitelisted,
		lastGovernanceState:  params.LastState,
		url:                  params.URL,
		charter:              params.Charter,
		owner:                params.Owner,
		ownerAddresses:       params.OwnerAddresses,
		contributorAddresses: params.ContributorAddresses,
		createdDateTs:        params.CreatedDateTs,
		applicationDateTs:    params.ApplicationDateTs,
		approvalDateTs:       params.ApprovalDateTs,
		lastUpdatedDateTs:    params.LastUpdatedDateTs,
		appExpiry:            params.AppExpiry,
		unstakedDeposit:      params.UnstakedDeposit,
		challengeID:          params.ChallengeID,
	}
}

// Listing represents a newsroom listing in the Civil TCR
type Listing struct {
	name string

	contractAddress common.Address

	whitelisted bool

	lastGovernanceState GovernanceState

	url string

	charter *Charter

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

// Charter returns the data regarding charter post for the newsroom
func (l *Listing) Charter() *Charter {
	return l.charter
}

// SetCharter set the data regarding charter post for the newsroom
func (l *Listing) SetCharter(c *Charter) {
	l.charter = c
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

// SetAppExpiry returns the expiration date of the application to this newsroom
func (l *Listing) SetAppExpiry(appExpiry *big.Int) {
	l.appExpiry = appExpiry
}

// UnstakedDeposit returns the unstaked deposit of the newsroom
func (l *Listing) UnstakedDeposit() *big.Int {
	return l.unstakedDeposit
}

// SetUnstakedDeposit returns the unstaked deposit of the newsroom
func (l *Listing) SetUnstakedDeposit(unstakedDeposit *big.Int) {
	l.unstakedDeposit = unstakedDeposit
}

// SetChallengeID sets the challenge ID of the listing.
func (l *Listing) SetChallengeID(id *big.Int) {
	l.challengeID = id
}

// ChallengeID returns the most recent challengeID of the listing
func (l *Listing) ChallengeID() *big.Int {
	return l.challengeID
}
