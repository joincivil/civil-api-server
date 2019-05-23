// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

const (
	// ChallengePollType is the name for regular challenge type
	ChallengePollType = "CHALLENGE"
	// AppealChallengePollType is the name for appeal challenge type
	AppealChallengePollType = "APPEAL_CHALLENGE"
	// ParamProposalPollType is the name for param proposal challenge type
	ParamProposalPollType = "PARAMETER_PROPOSAL_CHALLENGE"
	// GovProposalPollType is the name for gov param proposal challenge type
	GovProposalPollType = "GOVT_PARAMETER_PROPOSAL"
)

// NewChallenge is a convenience function to initialize a new Challenge struct
func NewChallenge(challengeID *big.Int, listingAddress common.Address, statement string, rewardPool *big.Int,
	challenger common.Address, resolved bool, stake *big.Int, totalTokens *big.Int,
	requestAppealExpiry *big.Int, challengeType string, lastUpdatedDateTs int64) *Challenge {
	return &Challenge{
		challengeID:         challengeID,
		listingAddress:      listingAddress,
		statement:           statement,
		rewardPool:          rewardPool,
		challenger:          challenger,
		resolved:            resolved,
		stake:               stake,
		totalTokens:         totalTokens,
		requestAppealExpiry: requestAppealExpiry,
		challengeType:       challengeType,
		lastUpdatedDateTs:   lastUpdatedDateTs,
	}
}

// Challenge represents a ChallengeData object
type Challenge struct {
	challengeID *big.Int

	listingAddress common.Address

	statement string

	rewardPool *big.Int

	challenger common.Address

	resolved bool

	stake *big.Int

	totalTokens *big.Int

	requestAppealExpiry *big.Int

	challengeType string

	lastUpdatedDateTs int64
}

// ChallengeID returns the challenge ID
func (c *Challenge) ChallengeID() *big.Int {
	return c.challengeID
}

// ListingAddress returns the listing address associataed with this challenge
func (c *Challenge) ListingAddress() common.Address {
	return c.listingAddress
}

// Statement returns the statement
func (c *Challenge) Statement() string {
	return c.statement
}

// RewardPool returns the RewardPool
func (c *Challenge) RewardPool() *big.Int {
	return c.rewardPool
}

// SetRewardPool sets the RewardPool
func (c *Challenge) SetRewardPool(rewardPool *big.Int) {
	c.rewardPool = rewardPool
}

// Challenger returns the challenger address
func (c *Challenge) Challenger() common.Address {
	return c.challenger
}

// Resolved returns whether this challenge was resolved
func (c *Challenge) Resolved() bool {
	return c.resolved
}

// SetResolved sets resolved boolean
func (c *Challenge) SetResolved(resolved bool) {
	c.resolved = resolved
}

// Stake returns the stake of this challenge
func (c *Challenge) Stake() *big.Int {
	return c.stake
}

// SetStake sets the stake of this challenge
func (c *Challenge) SetStake(stake *big.Int) {
	c.stake = stake
}

// TotalTokens returns the totaltokens for reward distribution purposes
func (c *Challenge) TotalTokens() *big.Int {
	return c.totalTokens
}

// SetTotalTokens sets totaltokens
func (c *Challenge) SetTotalTokens(totalTokens *big.Int) {
	c.totalTokens = totalTokens
}

// RequestAppealExpiry returns the requestAppealExpiry from challenge
func (c *Challenge) RequestAppealExpiry() *big.Int {
	return c.requestAppealExpiry
}

// ChallengeType returns the type of challenge this is
func (c *Challenge) ChallengeType() string {
	return c.challengeType
}

// LastUpdatedDateTs returns the ts of last update
func (c *Challenge) LastUpdatedDateTs() int64 {
	return c.lastUpdatedDateTs
}

// SetLastUpdateDateTs sets the date of last update
func (c *Challenge) SetLastUpdateDateTs(ts int64) {
	c.lastUpdatedDateTs = ts
}
