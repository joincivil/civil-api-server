// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

// Use golang.org/x/tools/cmd/stringer to generate the
// governancestate_string.go stringer method.
// stringer -type=GovernanceState

// GovernanceState specifies the current state of a listing
type GovernanceState int

const (
	// GovernanceStateNone is an invalid/empty governance state
	GovernanceStateNone GovernanceState = iota
	// GovernanceStateApplied is when a listing has just applied
	GovernanceStateApplied
	// GovernanceStateAppRemoved is when a listing's application has been removed
	GovernanceStateAppRemoved

	// GovernanceStateChallenged is when a listing has been challenged
	GovernanceStateChallenged
	// GovernanceStateChallengeFailed is when a challenge on a listing has failed
	GovernanceStateChallengeFailed
	// GovernanceStateChallengeSucceeded is when a challenge on a listing has succeeded
	GovernanceStateChallengeSucceeded
	// GovernanceStateFailedChallengeOverturned is when the original challenge on a
	// listing has failed
	GovernanceStateFailedChallengeOverturned
	// GovernanceStateSuccessfulChallengeOverturned is when the original challenge
	// on a listing has succeeded
	GovernanceStateSuccessfulChallengeOverturned

	// GovernanceStateAppWhitelisted is when a listing has been whitelisted
	GovernanceStateAppWhitelisted
	// GovernanceStateRemoved is when a listing has been removed
	GovernanceStateRemoved
	// GovernanceStateWithdrawn is when a listing has been withdrawn before it
	// has been whitelisted
	GovernanceStateWithdrawn

	// GovernanceStateAppealGranted is when an appeal is granted by the CC
	GovernanceStateAppealGranted
	// GovernanceStateAppealRequested is when an appeal against a listing is requested
	GovernanceStateAppealRequested
	// GovernanceStateGrantedAppealChallenged is when a granted appeal is challenged
	GovernanceStateGrantedAppealChallenged
	// GovernanceStateGrantedAppealConfirmed is when a granted appeal is confirmed and the
	// appeal challenge has failed
	GovernanceStateGrantedAppealConfirmed
	// GovernanceStateGrantedAppealOverturned is when a granted appeal decision is overturned
	GovernanceStateGrantedAppealOverturned
)

// Events unused for states
// "Deposit"
// "GovernmentTransfered"
// "RewardClaimed"
// "TouchAndRemoved"
// "Withdrawal"
