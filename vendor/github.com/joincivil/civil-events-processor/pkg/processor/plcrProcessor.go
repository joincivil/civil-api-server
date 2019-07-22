package processor

import (
	"fmt"
	"math/big"
	"strings"

	log "github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	commongen "github.com/joincivil/civil-events-crawler/pkg/generated/common"
	crawlermodel "github.com/joincivil/civil-events-crawler/pkg/model"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/persistence"

	cerrors "github.com/joincivil/go-common/pkg/errors"
	cpersist "github.com/joincivil/go-common/pkg/persistence"
	ctime "github.com/joincivil/go-common/pkg/time"
)

const (
	votesForFieldName     = "VotesFor"
	votesAgainstFieldName = "VotesAgainst"
	pollTypeFieldName     = "PollType"

	didUserRevealFieldName = "UserDidReveal"
	choiceFieldName        = "Choice"
	saltFieldName          = "Salt"
	didUserRescueFieldName = "DidUserRescue"
	latestVoteFieldName    = "LatestVote"
)

// NewPlcrEventProcessor is a convenience function to init an EventProcessor
func NewPlcrEventProcessor(client bind.ContractBackend, pollPersister model.PollPersister,
	userChallengeDataPersister model.UserChallengeDataPersister,
	challengePersister model.ChallengePersister,
	appealPersister model.AppealPersister, errRep cerrors.ErrorReporter) *PlcrEventProcessor {
	return &PlcrEventProcessor{
		client:                     client,
		pollPersister:              pollPersister,
		challengePersister:         challengePersister,
		userChallengeDataPersister: userChallengeDataPersister,
		appealPersister:            appealPersister,
		errRep:                     errRep,
	}
}

// PlcrEventProcessor handles the processing of raw events into aggregated data
// for use via the API.
type PlcrEventProcessor struct {
	client                     bind.ContractBackend
	pollPersister              model.PollPersister
	challengePersister         model.ChallengePersister
	userChallengeDataPersister model.UserChallengeDataPersister
	appealPersister            model.AppealPersister
	errRep                     cerrors.ErrorReporter
}

func (p *PlcrEventProcessor) isValidPLCRContractEventName(name string) bool {
	name = strings.Trim(name, " _")
	eventNames := commongen.EventTypesCivilPLCRVotingContract()
	return isStringInSlice(eventNames, name)
}

func (p *PlcrEventProcessor) pollIDFromEvent(event *crawlermodel.Event) (*big.Int, error) {
	payload := event.EventPayload()
	pollID, ok := payload["PollID"]
	if !ok {
		return nil, errors.New("Unable to find the poll ID in the payload")
	}
	return pollID.(*big.Int), nil
}

// Process processes Plcr Events into aggregated data
func (p *PlcrEventProcessor) Process(event *crawlermodel.Event) (bool, error) {
	if !p.isValidPLCRContractEventName(event.EventType()) {
		return false, nil
	}

	var err error
	ran := true
	eventName := strings.Trim(event.EventType(), " _")
	// Handling all the actionable events from PLCR Contract
	switch eventName {
	case "PollCreated":
		pollID, pollIDerr := p.pollIDFromEvent(event)
		if pollIDerr != nil {
			log.Infof("Error retrieving pollID: %v", err)
			p.errRep.Error(err, nil)
		}
		log.Infof("Handling PollCreated for pollID %v\n", pollID)
		err = p.processPollCreated(event, pollID)
	case "VoteCommitted":
		pollID, pollIDerr := p.pollIDFromEvent(event)
		if pollIDerr != nil {
			log.Infof("Error retrieving pollID: %v", err)
			p.errRep.Error(err, nil)
		}
		log.Infof("Handling VoteCommitted for pollID %v\n", pollID)
		err = p.processVoteCommitted(event, pollID)
	case "VoteRevealed":
		pollID, pollIDerr := p.pollIDFromEvent(event)
		if pollIDerr != nil {
			log.Infof("Error retrieving pollID: %v", err)
			p.errRep.Error(err, nil)
		}
		log.Infof("Handling VoteRevealed for pollID %v\n", pollID)
		err = p.processVoteRevealed(event, pollID)
	case "TokensRescued":
		pollID, pollIDerr := p.pollIDFromEvent(event)
		if pollIDerr != nil {
			log.Infof("Error retrieving pollID: %v", err)
			p.errRep.Error(err, nil)
		}
		log.Infof("Handling TokensRescued for pollID %v\n", pollID)
		err = p.processTokensRescued(event, pollID)

	}

	return ran, err
}

func (p *PlcrEventProcessor) processPollCreated(event *crawlermodel.Event,
	pollID *big.Int) error {
	payload := event.EventPayload()
	voteQuorum, ok := payload["VoteQuorum"]
	if !ok {
		return errors.New("No voteQuorum found")
	}

	commitEndDate, ok := payload["CommitEndDate"]
	if !ok {
		return errors.New("No commitEndDate found")
	}

	revealEndDate, ok := payload["RevealEndDate"]
	if !ok {
		return errors.New("No revealEndDate found")
	}
	votesFor := big.NewInt(0)
	votesAgainst := big.NewInt(0)

	poll := model.NewPoll(
		pollID,
		commitEndDate.(*big.Int),
		revealEndDate.(*big.Int),
		voteQuorum.(*big.Int),
		votesFor,
		votesAgainst,
		ctime.CurrentEpochSecsInInt64(),
	)

	return p.pollPersister.CreatePoll(poll)
}

func (p *PlcrEventProcessor) processVoteCommitted(event *crawlermodel.Event,
	pollID *big.Int) error {
	payload := event.EventPayload()
	numTokens := payload["NumTokens"]
	voterAddress := payload["Voter"]
	userDidCommit := true

	voteCommittedTimestamp := event.Timestamp()

	poll, err := p.pollPersister.PollByPollID(int(pollID.Int64()))
	if err != nil {
		// TODO: We should do contract call and save to DB bc it should exist in DB
		return err
	}

	pollType := poll.PollType()

	// get poll type from challenge and set it in poll.
	if pollType == "" {
		challenge, cErr := p.challengePersister.ChallengeByChallengeID(int(pollID.Int64()))
		if cErr != nil && cErr != cpersist.ErrPersisterNoResults {
			return cErr
		}
		// this will return errpersisternoresults upon gov param challenges
		// Once processing gov contract, this will be fixed, for now manually add this
		if cErr == cpersist.ErrPersisterNoResults {
			pollType = model.GovProposalPollType
		} else {
			pollType = challenge.ChallengeType()
		}

		poll.SetPollType(pollType)
		err = p.pollPersister.UpdatePoll(poll, []string{pollTypeFieldName})
		if err != nil {
			return err
		}
	}
	var parentChallengeID *big.Int
	if pollType == model.AppealChallengePollType {
		// set parentchallengeid here from appeal table, but could add an additional field
		// for parentchallengeID in challenge model
		appeal, aErr := p.appealPersister.AppealByAppealChallengeID(int(pollID.Int64()))
		if aErr != nil {
			return err
		}
		parentChallengeID = appeal.OriginalChallengeID()
	}

	// If there are any existing votes for this user for this poll, mark it as
	// NOT the latest vote and continue adding a new row for the latest vote.
	existingUserChallengeData := &model.UserChallengeData{}
	existingUserChallengeData.SetLatestVote(false)
	existingUserChallengeData.SetPollID(pollID)
	existingUserChallengeData.SetUserAddress(voterAddress.(common.Address))
	updatedFields := []string{latestVoteFieldName}
	// This is false because we want to update all existing committed votes
	latestVote := false
	updateWithUserAddress := true

	err = p.userChallengeDataPersister.UpdateUserChallengeData(existingUserChallengeData,
		updatedFields, updateWithUserAddress, latestVote)
	// If no rows affected, that means there is no existing vote for a user for this poll,
	// so continue to save. If there is an error, then return.
	if err != nil && err != persistence.ErrNoRowsAffected {
		return err
	}

	// Create a new row with the new/updated committed vote value for this user for this poll
	userChallengeData := model.NewUserChallengeData(
		voterAddress.(common.Address),
		pollID,
		numTokens.(*big.Int),
		userDidCommit,
		poll.RevealEndDate(),
		pollType,
		voteCommittedTimestamp,
		ctime.CurrentEpochSecsInInt64(),
	)
	if parentChallengeID != nil {
		userChallengeData.SetParentChallengeID(parentChallengeID)
	}
	userChallengeData.SetLatestVote(true)
	return p.userChallengeDataPersister.CreateUserChallengeData(userChallengeData)
}

func (p *PlcrEventProcessor) processVoteRevealed(event *crawlermodel.Event,
	pollID *big.Int) error {

	payload := event.EventPayload()

	choice, ok := payload["Choice"]
	if !ok {
		return errors.New("No choice found")
	}
	voter, ok := payload["Voter"]
	if !ok {
		return errors.New("No voter found")
	}
	salt, ok := payload["Salt"]
	if !ok {
		return errors.New("No salt found")
	}

	poll, err := p.pollPersister.PollByPollID(int(pollID.Int64()))
	if err != nil && err != cpersist.ErrPersisterNoResults {
		return err
	}
	if poll == nil {
		// TODO(IS): create new poll. If getting events in order, this shouldn't happen.
		return fmt.Errorf("Error persisting poll: no poll with ID: %v", pollID)
	}

	var updatedFields []string
	if choice.(*big.Int).Int64() == 1 {
		votesFor, ok := payload["VotesFor"]
		if !ok {
			return errors.New("No votesFor found")
		}
		poll.UpdateVotesFor(votesFor.(*big.Int))
		updatedFields = []string{votesForFieldName}
	} else {
		votesAgainst, ok := payload["VotesAgainst"]
		if !ok {
			return errors.New("No votesAgainst found")
		}
		poll.UpdateVotesAgainst(votesAgainst.(*big.Int))
		updatedFields = []string{votesAgainstFieldName}
	}

	err = p.pollPersister.UpdatePoll(poll, updatedFields)
	if err != nil {
		return fmt.Errorf("Error updating poll, err: %v", err)
	}

	userChallengeData, err := p.userChallengeDataPersister.UserChallengeDataByCriteria(
		&model.UserChallengeDataCriteria{
			UserAddress: voter.(common.Address).Hex(),
			PollID:      pollID.Uint64(),
		},
	)

	if err != nil || len(userChallengeData) > 1 {
		return fmt.Errorf("Error getting userChallengedata to update, err: %v", err)
	}

	didReveal := true

	// NOTE: At this point, userChallengeData can only be length 1
	existingUserChallengeData := userChallengeData[0]
	existingUserChallengeData.SetUserDidReveal(didReveal)
	existingUserChallengeData.SetSalt(salt.(*big.Int))
	existingUserChallengeData.SetChoice(choice.(*big.Int))

	updatedUserFields := []string{didUserRevealFieldName, saltFieldName, choiceFieldName}
	updateWithUserAddress := true
	latestVote := true
	err = p.userChallengeDataPersister.UpdateUserChallengeData(existingUserChallengeData,
		updatedUserFields, updateWithUserAddress, latestVote)
	if err != nil {
		return fmt.Errorf("Error updating UserChallengeData, err: %v", err)
	}

	return nil
}

func (p *PlcrEventProcessor) processTokensRescued(event *crawlermodel.Event,
	pollID *big.Int) error {

	payload := event.EventPayload()
	voter, ok := payload["Voter"]
	if !ok {
		return errors.New("No voter found")
	}

	// TODO(IS): Update user challengedata object with didUserRescue True,
	userChallengeData, err := p.userChallengeDataPersister.UserChallengeDataByCriteria(
		&model.UserChallengeDataCriteria{
			UserAddress: voter.(common.Address).Hex(),
			PollID:      pollID.Uint64(),
		},
	)
	if err != nil || len(userChallengeData) > 1 {
		return fmt.Errorf("Error getting userChallengedata to update, err: %v", err)
	}

	// NOTE: At this point, userChallengeData can only be length 1
	existingUserChallengeData := userChallengeData[0]
	existingUserChallengeData.SetDidUserRescue(true)
	updatedUserFields := []string{didUserRescueFieldName}
	updateWithUserAddress := true
	latestVote := true

	err = p.userChallengeDataPersister.UpdateUserChallengeData(existingUserChallengeData,
		updatedUserFields, updateWithUserAddress, latestVote)
	if err != nil {
		return fmt.Errorf("Error updating UserChallengeData, err: %v", err)
	}

	return nil
}
