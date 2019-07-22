// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

// NOTE: IF THE ENUM IS UPDATED, PLEASE DO THIS:
// Use golang.org/x/tools/cmd/stringer to generate the
// governancestate_string.go stringer method.
// stringer -type=GovernanceState

// GovernanceState specifies the current state of a listing
type GovernanceState int

const (
	// GovernanceStateNone is an invalid/empty governance state - 0
	GovernanceStateNone GovernanceState = iota
	// GovernanceStateApplied is _Application when a listing has just applied - 1
	GovernanceStateApplied
	// GovernanceStateAppRemoved is _ApplicationRemoved when a
	// listing's application has been removed - 2
	GovernanceStateAppRemoved

	// GovernanceStateChallenged is _Challenge when a listing has been challenged - 3
	GovernanceStateChallenged
	// GovernanceStateChallengeFailed is _ChallengeFailed when a challenge on
	// a listing has failed - 4
	GovernanceStateChallengeFailed
	// GovernanceStateChallengeSucceeded is _ChallengeSucceeded when a challenge
	// on a listing has succeeded - 5
	GovernanceStateChallengeSucceeded
	// GovernanceStateFailedChallengeOverturned is _FailedChallengeOverturned
	// when the original challenge on a - 6
	// listing has failed
	GovernanceStateFailedChallengeOverturned
	// GovernanceStateSuccessfulChallengeOverturned _SuccessfulChallengeOverturned
	// is when the original challenge - 7
	// on a listing has succeeded
	GovernanceStateSuccessfulChallengeOverturned

	// GovernanceStateAppWhitelisted is _ApplicationWhitelisted when a listing
	// has been whitelisted - 8
	GovernanceStateAppWhitelisted
	// GovernanceStateRemoved is _ListingRemoved when a listing has been removed - 9
	GovernanceStateRemoved

	// GovernanceStateAppealGranted is _AppealGranted when an appeal is granted
	// by the CC - 10
	GovernanceStateAppealGranted
	// GovernanceStateAppealRequested is _AppealRequested when an appeal against
	// a listing is requested - 11
	GovernanceStateAppealRequested
	// GovernanceStateGrantedAppealChallenged is _GrantedAppealChallenged when
	// a granted appeal is challenged - 12
	GovernanceStateGrantedAppealChallenged
	// GovernanceStateGrantedAppealConfirmed is _GrantedAppealConfirmed when a
	// granted appeal is confirmed and the appeal challenge has failed - 13
	GovernanceStateGrantedAppealConfirmed
	// GovernanceStateGrantedAppealOverturned is _GrantedAppealOverturned when a
	// granted appeal decision is overturned - 14
	GovernanceStateGrantedAppealOverturned

	// GovernanceStateDeposit is _Deposit when the owner of a listing
	// increases their unstaked deposit - 15
	GovernanceStateDeposit
	// GovernanceStateWithdrawal is _Withdrawal when the owner of a listing
	// decreases their unstaked deposit - 16
	GovernanceStateWithdrawal

	// GovernanceStateRewardClaimed is _RewardClaimed when a voted claims their
	// reward - 17
	GovernanceStateRewardClaimed

	// GovernanceStateTouchRemoved is _TouchAndRemoved. This event is not
	// actionable - 18
	GovernanceStateTouchRemoved

	// GovernanceStateListingWithdrawn is _ListingWithdrawn. This event is
	// not actionable - 19
	GovernanceStateListingWithdrawn
)

// Events that are not actionable (Check out crawler doc for details)
// NOTE(IS): We should actually still process these because we should update
// governanceevents with them.
// "GovernmentTransfered"
