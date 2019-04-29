// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// ParameterProposalParams are params to create a new parameter proposal
type ParameterProposalParams struct {
	Name              string
	Value             *big.Int
	PropID            [32]byte
	Deposit           *big.Int
	AppExpiry         *big.Int
	ChallengeID       *big.Int
	Proposer          common.Address
	Accepted          bool
	Expired           bool
	LastUpdatedDateTs int64
}

// NewParameterProposal is a convenience function to create a new parameter proposal
func NewParameterProposal(params *ParameterProposalParams) *ParameterProposal {
	return &ParameterProposal{
		name:              params.Name,
		value:             params.Value,
		propID:            params.PropID,
		deposit:           params.Deposit,
		appExpiry:         params.AppExpiry,
		challengeID:       params.ChallengeID,
		proposer:          params.Proposer,
		accepted:          params.Accepted,
		expired:           params.Expired,
		lastUpdatedDateTs: params.LastUpdatedDateTs,
	}
}

// ParameterProposal represents a parameterizer proposal
type ParameterProposal struct {
	name string

	value *big.Int

	propID [32]byte

	deposit *big.Int

	appExpiry *big.Int

	challengeID *big.Int

	proposer common.Address

	accepted bool

	expired bool

	lastUpdatedDateTs int64
}

// Name is the name of the parameter field
func (p *ParameterProposal) Name() string {
	return p.name
}

// Value is the value of the parameter
func (p *ParameterProposal) Value() *big.Int {
	return p.value
}

// PropID is the id of proposal
func (p *ParameterProposal) PropID() [32]byte {
	return p.propID
}

// Deposit is the deposit
func (p *ParameterProposal) Deposit() *big.Int {
	return p.deposit
}

// AppExpiry is the proposal's date of expiration
func (p *ParameterProposal) AppExpiry() *big.Int {
	return p.appExpiry
}

// ChallengeID is the challenge id of this proposal
func (p *ParameterProposal) ChallengeID() *big.Int {
	return p.challengeID
}

// Proposer is the address of proposer
func (p *ParameterProposal) Proposer() common.Address {
	return p.proposer
}

// Accepted is whether this proposal has been accepted
func (p *ParameterProposal) Accepted() bool {
	return p.accepted
}

// Expired is whether this proposal is expired
func (p *ParameterProposal) Expired() bool {
	return p.expired
}

// LastUpdatedDateTs is the timestamp of last update
func (p *ParameterProposal) LastUpdatedDateTs() int64 {
	return p.lastUpdatedDateTs
}

// SetAccepted sets accepted field
func (p *ParameterProposal) SetAccepted(accepted bool) {
	p.accepted = accepted
}

// SetExpired sets expired field
func (p *ParameterProposal) SetExpired(expired bool) {
	p.expired = expired
}

// SetLastUpdatedDateTs sets the value of the last time this proposal was updated
func (p *ParameterProposal) SetLastUpdatedDateTs(date int64) {
	p.lastUpdatedDateTs = date
}
