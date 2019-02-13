// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Appeal represents the appealdata for a Challenge
type Appeal struct {
	originalChallengeID *big.Int

	requester common.Address

	appealFeePaid *big.Int

	appealPhaseExpiry *big.Int

	appealGranted bool

	appealOpenToChallengeExpiry *big.Int

	statement string // this is type ContentData in dapp

	appealGrantedStatementURI string

	appealChallengeID *big.Int

	lastUpdatedDateTs int64
}

// NewAppeal creates a new appeal object
func NewAppeal(originalChallengeID *big.Int, requester common.Address, appealFeePaid *big.Int,
	appealPhaseExpiry *big.Int, appealGranted bool, statement string, lastUpdatedTs int64,
	appealGrantedStatementURI string) *Appeal {
	return &Appeal{
		originalChallengeID:       originalChallengeID,
		requester:                 requester,
		appealFeePaid:             appealFeePaid,
		appealPhaseExpiry:         appealPhaseExpiry,
		appealGranted:             appealGranted,
		statement:                 statement,
		lastUpdatedDateTs:         lastUpdatedTs,
		appealGrantedStatementURI: appealGrantedStatementURI,
	}
}

// OriginalChallengeID returns the original challenge ID
func (a *Appeal) OriginalChallengeID() *big.Int {
	return a.originalChallengeID
}

// SetOriginalChallengeID sets the original challenge ID
func (a *Appeal) SetOriginalChallengeID(challengeID *big.Int) {
	a.originalChallengeID = challengeID
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

// AppealGranted returns whether appeal was granted.
func (a *Appeal) AppealGranted() bool {
	return a.appealGranted
}

// SetAppealGranted sets appealGranted
func (a *Appeal) SetAppealGranted(appealGranted bool) {
	a.appealGranted = appealGranted
}

// AppealOpenToChallengeExpiry returns AppealOpenToChallengeExpiry
func (a *Appeal) AppealOpenToChallengeExpiry() *big.Int {
	return a.appealOpenToChallengeExpiry
}

// SetAppealOpenToChallengeExpiry sets appealOpenToChallengeExpiry if there is an appealChallenge
func (a *Appeal) SetAppealOpenToChallengeExpiry(appealOpenToChallengeExpiry *big.Int) {
	a.appealOpenToChallengeExpiry = appealOpenToChallengeExpiry
}

// Statement returns statement
func (a *Appeal) Statement() string {
	return a.statement
}

// AppealChallengeID returns appealchallengeid
func (a *Appeal) AppealChallengeID() *big.Int {
	return a.appealChallengeID
}

// SetAppealChallengeID sets appealChallengeID if there is an appealChallenge
func (a *Appeal) SetAppealChallengeID(challengeID *big.Int) {
	a.appealChallengeID = challengeID
}

// LastUpdatedDateTs is the ts of the last time the processor updated this struct
func (a *Appeal) LastUpdatedDateTs() int64 {
	return a.lastUpdatedDateTs
}

// SetLastUpdatedDateTs updates the lastUpdatedTs
func (a *Appeal) SetLastUpdatedDateTs(lastUpdatedTs int64) {
	a.lastUpdatedDateTs = lastUpdatedTs
}

// AppealGrantedStatementURI is the uri of the appeal granted statement
func (a *Appeal) AppealGrantedStatementURI() string {
	return a.appealGrantedStatementURI
}

// SetAppealGrantedStatementURI updates the appealGrantedStatementUri
func (a *Appeal) SetAppealGrantedStatementURI(uri string) {
	a.appealGrantedStatementURI = uri
}
