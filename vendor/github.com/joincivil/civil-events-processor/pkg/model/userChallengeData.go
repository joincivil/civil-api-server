// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// NewUserChallengeData creates a new userChallengeData
func NewUserChallengeData(address common.Address, pollID *big.Int,
	numTokens *big.Int, userDidCommit bool, pollRevealEndDate *big.Int,
	pollType string, voteCommittedTs int64, lastUpdatedDateTs int64) *UserChallengeData {
	return &UserChallengeData{
		pollID:            pollID,
		userAddress:       address,
		numTokens:         numTokens,
		userDidCommit:     userDidCommit,
		pollRevealEndDate: pollRevealEndDate,
		pollType:          pollType,
		voteCommittedTs:   voteCommittedTs,
		lastUpdatedDateTs: lastUpdatedDateTs,
	}
}

// UserChallengeData is data related to the user challenge as defined in the dapp
type UserChallengeData struct {
	pollID            *big.Int
	pollRevealEndDate *big.Int
	pollType          string
	userAddress       common.Address
	userDidCommit     bool
	voteCommittedTs   int64
	userDidReveal     bool
	didUserCollect    bool
	didUserRescue     bool
	didCollectAmount  *big.Int
	isVoterWinner     bool
	pollIsPassed      bool
	salt              *big.Int
	choice            *big.Int
	numTokens         *big.Int
	voterReward       *big.Int
	parentChallengeID *big.Int
	latestVote        bool
	lastUpdatedDateTs int64
}

// PollID is the pollID of this vote
func (u *UserChallengeData) PollID() *big.Int {
	return u.pollID
}

// SetPollID sets the pollID of this vote
func (u *UserChallengeData) SetPollID(pollID *big.Int) {
	u.pollID = pollID
}

// PollRevealEndDate is the reveal end date of this poll
func (u *UserChallengeData) PollRevealEndDate() *big.Int {
	return u.pollRevealEndDate
}

// PollType is the type of poll
func (u *UserChallengeData) PollType() string {
	return u.pollType
}

// SetPollType sets the poll type
func (u *UserChallengeData) SetPollType(pollType string) {
	u.pollType = pollType
}

// UserAddress is the address of this user
func (u *UserChallengeData) UserAddress() common.Address {
	return u.userAddress
}

// SetUserAddress sets the address of this user
func (u *UserChallengeData) SetUserAddress(address common.Address) {
	u.userAddress = address
}

// UserDidCommit is whether this user committed a vote
func (u *UserChallengeData) UserDidCommit() bool {
	return u.userDidCommit
}

// VoteCommittedTs is the timestamp the user committed a vote
func (u *UserChallengeData) VoteCommittedTs() int64 {
	return u.voteCommittedTs
}

// UserDidReveal is whether this user revealed a vote
func (u *UserChallengeData) UserDidReveal() bool {
	return u.userDidReveal
}

// SetUserDidReveal sets whether this user revealed a vote
func (u *UserChallengeData) SetUserDidReveal(userDidReveal bool) {
	u.userDidReveal = userDidReveal
}

// DidUserCollect is whether this user has reward available and has collected rewards
func (u *UserChallengeData) DidUserCollect() bool {
	return u.didUserCollect
}

// SetDidUserCollect is whether this user has reward available and has collected rewards
func (u *UserChallengeData) SetDidUserCollect(didUserCollect bool) {
	u.didUserCollect = didUserCollect
}

// DidUserRescue is whether this user rescued: user committed but did not reveal or rescue
func (u *UserChallengeData) DidUserRescue() bool {
	return u.didUserRescue
}

// SetDidUserRescue is whether this user rescued: user committed but did not reveal or rescue
func (u *UserChallengeData) SetDidUserRescue(didUserRescue bool) {
	u.didUserRescue = didUserRescue
}

// DidCollectAmount is the reward this user claimed
func (u *UserChallengeData) DidCollectAmount() *big.Int {
	return u.didCollectAmount
}

// SetDidCollectAmount sets the reward this user claimed
func (u *UserChallengeData) SetDidCollectAmount(didCollectAmount *big.Int) {
	u.didCollectAmount = didCollectAmount
}

// IsVoterWinner is whether this vote won
func (u *UserChallengeData) IsVoterWinner() bool {
	return u.isVoterWinner
}

// SetIsVoterWinner sets whether this vote won
func (u *UserChallengeData) SetIsVoterWinner(isVoterWinner bool) {
	u.isVoterWinner = isVoterWinner
}

// PollIsPassed is whether this poll is passed
func (u *UserChallengeData) PollIsPassed() bool {
	return u.pollIsPassed
}

// SetPollIsPassed sets whether this poll is passed
func (u *UserChallengeData) SetPollIsPassed(pollIsPassed bool) {
	u.pollIsPassed = pollIsPassed
}

// Salt is the user's salt
func (u *UserChallengeData) Salt() *big.Int {
	return u.salt
}

// SetSalt sets the user's salt
func (u *UserChallengeData) SetSalt(salt *big.Int) {
	u.salt = salt
}

// Choice is what the user voted
func (u *UserChallengeData) Choice() *big.Int {
	return u.choice
}

// SetChoice sets what the user voted
func (u *UserChallengeData) SetChoice(choice *big.Int) {
	u.choice = choice
}

// NumTokens is the number of tokens the user put with the vote
func (u *UserChallengeData) NumTokens() *big.Int {
	return u.numTokens
}

// SetNumTokens is the number of tokens the user put with the vote
func (u *UserChallengeData) SetNumTokens(numTokens *big.Int) {
	u.numTokens = numTokens
}

// VoterReward is the voter reward
func (u *UserChallengeData) VoterReward() *big.Int {
	return u.voterReward
}

// SetVoterReward is the voter reward
func (u *UserChallengeData) SetVoterReward(voterReward *big.Int) {
	u.voterReward = voterReward
}

// ParentChallengeID is the parent challenge ID if this is a vote for appeal challenge
func (u *UserChallengeData) ParentChallengeID() *big.Int {
	return u.parentChallengeID
}

// SetParentChallengeID sets the parent challenge ID if this is a vote for appeal challenge
func (u *UserChallengeData) SetParentChallengeID(pChallengeID *big.Int) {
	u.parentChallengeID = pChallengeID
}

// LatestVote returns true if this was the latest vote commit for user/poll, else false
func (u *UserChallengeData) LatestVote() bool {
	return u.latestVote
}

// SetLatestVote sets the latest vote flag
func (u *UserChallengeData) SetLatestVote(latestVote bool) {
	u.latestVote = latestVote
}

// LastUpdatedDateTs returns the ts of last update
func (u *UserChallengeData) LastUpdatedDateTs() int64 {
	return u.lastUpdatedDateTs
}

// SetLastUpdatedDateTs sets the date of last update
func (u *UserChallengeData) SetLastUpdatedDateTs(ts int64) {
	u.lastUpdatedDateTs = ts
}
