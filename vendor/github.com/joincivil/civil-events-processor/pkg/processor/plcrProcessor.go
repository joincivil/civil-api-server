package processor

import (
	"fmt"
	log "github.com/golang/glog"
	"github.com/pkg/errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	commongen "github.com/joincivil/civil-events-crawler/pkg/generated/common"
	crawlermodel "github.com/joincivil/civil-events-crawler/pkg/model"

	"github.com/joincivil/civil-events-processor/pkg/model"

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
)

// NewPlcrEventProcessor is a convenience function to init an EventProcessor
func NewPlcrEventProcessor(client bind.ContractBackend, pollPersister model.PollPersister,
	userChallengeDataPersister model.UserChallengeDataPersister,
	challengePersister model.ChallengePersister,
	appealPersister model.AppealPersister) *PlcrEventProcessor {
	return &PlcrEventProcessor{
		client:                     client,
		pollPersister:              pollPersister,
		challengePersister:         challengePersister,
		userChallengeDataPersister: userChallengeDataPersister,
		appealPersister:            appealPersister,
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
		}
		log.Infof("Handling PollCreated for pollID %v\n", pollID)
		err = p.processPollCreated(event, pollID)
	case "VoteCommitted":
		pollID, pollIDerr := p.pollIDFromEvent(event)
		if pollIDerr != nil {
			log.Infof("Error retrieving pollID: %v", err)
		}
		log.Infof("Handling VoteCommitted for pollID %v\n", pollID)
		err = p.processVoteCommitted(event, pollID)
	case "VoteRevealed":
		pollID, pollIDerr := p.pollIDFromEvent(event)
		if pollIDerr != nil {
			log.Infof("Error retrieving pollID: %v", err)
		}
		log.Infof("Handling VoteRevealed for pollID %v\n", pollID)
		err = p.processVoteRevealed(event, pollID)
	case "TokensRescued":
		pollID, pollIDerr := p.pollIDFromEvent(event)
		if pollIDerr != nil {
			log.Infof("Error retrieving pollID: %v", err)
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

	poll, err := p.pollPersister.PollByPollID(int(pollID.Int64()))
	if err != nil {
		// TOODO: We should do contract call and save to DB bc it should exist in DB
		return err
	}

	pollType := poll.PollType()

	// NOTE: get poll type from challenge and set it in poll.
	if pollType == "" {
		challenge, err := p.challengePersister.ChallengeByChallengeID(int(pollID.Int64()))
		if err != nil && err != cpersist.ErrPersisterNoResults {
			return err
		}
		// NOTE(IS): this will return errpersisternoresults upon gov param challenges
		// Once processing gov contract, this will be fixed, for now manually add this
		if err == cpersist.ErrPersisterNoResults {
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
		// NOTE(IS): set parentchallengeid here from appeal table, but could add an additional field
		// for parentchallengeID in challenge model
		appeal, err := p.appealPersister.AppealByAppealChallengeID(int(pollID.Int64()))
		if err != nil {
			return err
		}
		parentChallengeID = appeal.OriginalChallengeID()
	}

	userChallengeData := model.NewUserChallengeData(
		voterAddress.(common.Address),
		pollID,
		numTokens.(*big.Int),
		userDidCommit,
		poll.RevealEndDate(),
		pollType,
		ctime.CurrentEpochSecsInInt64(),
	)
	if parentChallengeID != nil {
		userChallengeData.SetParentChallengeID(parentChallengeID)
	}
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
	userChallengeData[0].SetUserDidReveal(didReveal)
	userChallengeData[0].SetSalt(salt.(*big.Int))
	userChallengeData[0].SetChoice(choice.(*big.Int))

	updatedUserFields := []string{didUserRevealFieldName, saltFieldName, choiceFieldName}
	updateWithUserAddress := true
	err = p.userChallengeDataPersister.UpdateUserChallengeData(userChallengeData[0],
		updatedUserFields, updateWithUserAddress)
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

	updatedUserFields := []string{didUserRescueFieldName}
	updateWithUserAddress := true
	// NOTE: At this point, userChallengeData can only be length 1
	err = p.userChallengeDataPersister.UpdateUserChallengeData(userChallengeData[0],
		updatedUserFields, updateWithUserAddress)
	if err != nil {
		return fmt.Errorf("Error updating UserChallengeData, err: %v", err)
	}

	return nil
}
