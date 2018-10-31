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
	// GovernanceStateApplied is _Application when a listing has just applied
	GovernanceStateApplied
	// GovernanceStateAppRemoved is _ApplicationRemoved when a listing's application has been removed
	GovernanceStateAppRemoved

	// GovernanceStateChallenged is _Challenge when a listing has been challenged
	GovernanceStateChallenged
	// GovernanceStateChallengeFailed is _ChallengeFailed when a challenge on a listing has failed
	GovernanceStateChallengeFailed
	// GovernanceStateChallengeSucceeded is _ChallengeSucceeded when a challenge on a listing has succeeded
	GovernanceStateChallengeSucceeded
	// GovernanceStateFailedChallengeOverturned is _FailedChallengeOverturned when the original challenge on a
	// listing has failed
	GovernanceStateFailedChallengeOverturned
	// GovernanceStateSuccessfulChallengeOverturned _SuccessfulChallengeOverturned is when the original challenge
	// on a listing has succeeded
	GovernanceStateSuccessfulChallengeOverturned

	// GovernanceStateAppWhitelisted is _ApplicationWhitelisted when a listing has been whitelisted
	GovernanceStateAppWhitelisted
	// GovernanceStateRemoved is _ListingRemoved when a listing has been removed
	GovernanceStateRemoved
	// GovernanceStateWithdrawn is _ListingWithdrawn when a listing has been withdrawn before it
	// has been whitelisted
	GovernanceStateWithdrawn

	// GovernanceStateAppealGranted is _AppealGranted when an appeal is granted by the CC
	GovernanceStateAppealGranted
	// GovernanceStateAppealRequested is _AppealRequested when an appeal against a listing is requested
	GovernanceStateAppealRequested
	// GovernanceStateGrantedAppealChallenged is _GrantedAppealChallenged when a granted appeal is challenged
	GovernanceStateGrantedAppealChallenged
	// GovernanceStateGrantedAppealConfirmed is _GrantedAppealConfirmed when a granted appeal is confirmed and the
	// appeal challenge has failed
	GovernanceStateGrantedAppealConfirmed
	// GovernanceStateGrantedAppealOverturned is _GrantedAppealOverturned when a granted appeal decision is overturned
	GovernanceStateGrantedAppealOverturned

	// GovernanceStateDeposit is _Deposit when the owner of a listing increases their unstaked deposit
	GovernanceStateDeposit
	// GovernanceStateDepositWithdrawl is _Withdrawal when the owner of a listing decreases their unstaked deposit
	GovernanceStateDepositWithdrawl
)

// Events unused for states
// "GovernmentTransfered"
// "RewardClaimed"
// "TouchAndRemoved"
