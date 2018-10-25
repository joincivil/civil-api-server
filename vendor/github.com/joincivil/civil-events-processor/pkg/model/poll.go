// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"math/big"
)

// Poll represents pollData for a Challenge
type Poll struct {
	pollID *big.Int

	commitEndDate *big.Int

	revealEndDate *big.Int

	voteQuorum *big.Int

	votesFor *big.Int

	votesAgainst *big.Int

	lastUpdatedDateTs int64
}

// NewPoll creates a new Poll
func NewPoll(pollID *big.Int, commitEndDate *big.Int, revealEndDate *big.Int, voteQuorum *big.Int,
	votesFor *big.Int, votesAgainst *big.Int, lastUpdatedTs int64) *Poll {
	return &Poll{
		pollID:            pollID,
		commitEndDate:     commitEndDate,
		revealEndDate:     revealEndDate,
		voteQuorum:        voteQuorum,
		votesFor:          votesFor,
		votesAgainst:      votesAgainst,
		lastUpdatedDateTs: lastUpdatedTs,
	}
}

// PollID returns the pollID of this poll. It corresponds to a challengeID in challenge.go
func (p *Poll) PollID() *big.Int {
	return p.pollID
}

// CommitEndDate returns the commitenddate
func (p *Poll) CommitEndDate() *big.Int {
	return p.commitEndDate
}

// RevealEndDate returns the RevealEndDate
func (p *Poll) RevealEndDate() *big.Int {
	return p.revealEndDate
}

// VoteQuorum returns the VoteQuorum
func (p *Poll) VoteQuorum() *big.Int {
	return p.voteQuorum
}

// VotesFor returns the VotesFor
func (p *Poll) VotesFor() *big.Int {
	return p.votesFor
}

// UpdateVotesFor updates votes for poll
func (p *Poll) UpdateVotesFor(votesFor *big.Int) {
	p.votesFor = votesFor
}

// VotesAgainst returns the VotesAgainst
func (p *Poll) VotesAgainst() *big.Int {
	return p.votesAgainst
}

// UpdateVotesAgainst updates votes against poll
func (p *Poll) UpdateVotesAgainst(votesAgainst *big.Int) {
	p.votesAgainst = votesAgainst
}

// LastUpdatedDateTs is the ts of the last time the processor updated this struct
func (p *Poll) LastUpdatedDateTs() int64 {
	return p.lastUpdatedDateTs
}

// SetLastUpdatedDateTs updates the lastUpdatedTs
func (p *Poll) SetLastUpdatedDateTs(lastUpdatedTs int64) {
	p.lastUpdatedDateTs = lastUpdatedTs
}
