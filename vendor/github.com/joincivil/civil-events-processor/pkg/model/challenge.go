// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// Poll represents pollData for a Challenge
type Poll struct {
	commitEndDate int64

	revealEndDate int64

	voteQuorum uint64

	votesFor uint64

	votesAgainst uint64
}

// NewPoll creates a new Poll
func NewPoll(commitEndDate int64, revealEndDate int64, voteQuorum uint64, votesFor uint64,
	votesAgainst uint64) *Poll {
	return &Poll{
		commitEndDate: commitEndDate,
		revealEndDate: revealEndDate,
		voteQuorum:    voteQuorum,
		votesFor:      votesFor,
		votesAgainst:  votesAgainst,
	}
}

// UpdateVotesFor updates votes for poll
func (p *Poll) UpdateVotesFor(votesFor uint64) {
	p.votesFor = votesFor
}

// UpdateVotesAgainst updates votes against poll
func (p *Poll) UpdateVotesAgainst(votesAgainst uint64) {
	p.votesAgainst = votesAgainst
}

//CommitEndDate returns the commitenddate
func (p *Poll) CommitEndDate() int64 {
	return p.commitEndDate
}

// RevealEndDate returns the RevealEndDate
func (p *Poll) RevealEndDate() int64 {
	return p.revealEndDate
}

// VoteQuorum returns the VoteQuorum
func (p *Poll) VoteQuorum() uint64 {
	return p.voteQuorum
}

// VotesFor returns the VotesFor
func (p *Poll) VotesFor() uint64 {
	return p.votesFor
}

// VotesAgainst returns the VotesAgainst
func (p *Poll) VotesAgainst() uint64 {
	return p.votesAgainst
}

// Appeal represents the appealdata for a Challenge
// NOTE(IS): will leave this unfilled for now
type Appeal struct {
	requester common.Address

	appealFeePaid *big.Int

	appealPhaseExpiry *big.Int

	appealGranted bool

	appealOpenToChallengeExpiry *big.Int

	statement string // this is type ContentData in dapp

	// NOTE(IS): This corresponds to another challengeID in the challenge table
	appealChallengeID *big.Int
}

// Requester returns the Appeal requester
func (a *Appeal) Requester() common.Address {
	return a.requester
}

// AppealFeePaid returns the AppealFeePaid
func (a *Appeal) AppealFeePaid() *big.Int {
	return a.appealFeePaid
}

// AppealPhaseExpiry returns the AppealPhaseExpiry
func (a *Appeal) AppealPhaseExpiry() *big.Int {
	return a.appealPhaseExpiry
}

// AppealGranted returns wheter appeal was granted.
func (a *Appeal) AppealGranted() bool {
	return a.appealGranted
}

// AppealOpenToChallengeExpiry returns AppealOpenToChallengeExpiry
func (a *Appeal) AppealOpenToChallengeExpiry() *big.Int {
	return a.appealOpenToChallengeExpiry
}

// Statement returns statement
func (a *Appeal) Statement() string {
	return a.statement
}

// AppealChallengeID returns appealchallengeid
func (a *Appeal) AppealChallengeID() *big.Int {
	return a.appealChallengeID
}

// NewChallenge is a convenience function to initialize a new Challenge struct
// NOTE(IS): Temporarily ignoring Appeal
func NewChallenge(challengeID *big.Int, listingAddress common.Address, statement string, rewardPool *big.Int,
	challenger common.Address, resolved bool, stake *big.Int, totalTokens *big.Int, poll *Poll,
	requestAppealExpiry *big.Int, lastUpdatedDateTs int64) *Challenge {
	return &Challenge{
		challengeID:         challengeID,
		listingAddress:      listingAddress,
		statement:           statement,
		rewardPool:          rewardPool,
		challenger:          challenger,
		resolved:            resolved,
		stake:               stake,
		totalTokens:         totalTokens,
		poll:                poll,
		requestAppealExpiry: requestAppealExpiry,
		lastUpdatedDateTs:   lastUpdatedDateTs,
		appeal:              Appeal{},
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

	poll *Poll

	requestAppealExpiry *big.Int

	appeal Appeal

	lastUpdatedDateTs int64
}

// SetPoll sets the poll for this challenge
func (c *Challenge) SetPoll(poll *Poll) {
	c.poll = poll
}

// SetLastUpdateDateTs sets the date of last update
func (c *Challenge) SetLastUpdateDateTs(ts int64) {
	c.lastUpdatedDateTs = ts
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

// Challenger returns the challenger address
func (c *Challenge) Challenger() common.Address {
	return c.challenger
}

// Resolved returns whether this challenge was resolved
func (c *Challenge) Resolved() bool {
	return c.resolved
}

// Stake returns the stake of this challenge
func (c *Challenge) Stake() *big.Int {
	return c.stake
}

// TotalTokens returns the totaltokens for reward distribution purposes
func (c *Challenge) TotalTokens() *big.Int {
	return c.totalTokens
}

// Poll returns the poll object from this challenge
func (c *Challenge) Poll() *Poll {
	return c.poll
}

// RequestAppealExpiry returns the requestAppealExpiry from challenge
func (c *Challenge) RequestAppealExpiry() *big.Int {
	return c.requestAppealExpiry
}

// Appeal returns the appeal object data from this challenge
func (c *Challenge) Appeal() Appeal {
	return c.appeal
}

// LastUpdatedDateTs returns the ts of last update
func (c *Challenge) LastUpdatedDateTs() int64 {
	return c.lastUpdatedDateTs
}
