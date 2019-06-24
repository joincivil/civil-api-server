package processor

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/pkg/errors"

	log "github.com/golang/glog"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	commongen "github.com/joincivil/civil-events-crawler/pkg/generated/common"
	crawlermodel "github.com/joincivil/civil-events-crawler/pkg/model"

	cerrors "github.com/joincivil/go-common/pkg/errors"
	"github.com/joincivil/go-common/pkg/generated/contract"
	cpersist "github.com/joincivil/go-common/pkg/persistence"
	ctime "github.com/joincivil/go-common/pkg/time"

	"github.com/joincivil/civil-events-processor/pkg/model"
)

var paramChallengeEventNames = []string{"ChallengeFailed", "ChallengeSucceeded", "NewChallenge"}

const (
	proposalAcceptedFieldName      = "Accepted"
	proposalExpiredFieldName       = "Expired"
	userChallengeIsPassedFieldName = "PollIsPassed"
)

// NewParameterizerEventProcessor is a convenience function to init a parameterizer processor
func NewParameterizerEventProcessor(client bind.ContractBackend, challengePersister model.ChallengePersister,
	paramProposalPersister model.ParamProposalPersister, pollPersister model.PollPersister,
	userChallengeDataPersister model.UserChallengeDataPersister, errRep cerrors.ErrorReporter) *ParameterizerEventProcessor {
	return &ParameterizerEventProcessor{
		client:                     client,
		challengePersister:         challengePersister,
		paramProposalPersister:     paramProposalPersister,
		pollPersister:              pollPersister,
		userChallengeDataPersister: userChallengeDataPersister,
		errRep:                     errRep,
	}
}

// ParameterizerEventProcessor handles the processing of raw events into aggregated data
type ParameterizerEventProcessor struct {
	client                     bind.ContractBackend
	challengePersister         model.ChallengePersister
	paramProposalPersister     model.ParamProposalPersister
	pollPersister              model.PollPersister
	userChallengeDataPersister model.UserChallengeDataPersister
	errRep                     cerrors.ErrorReporter
}

func (p *ParameterizerEventProcessor) isValidParameterizerContractEventName(name string) bool {
	name = strings.Trim(name, " _")
	eventNames := commongen.EventTypesParameterizerContract()
	return isStringInSlice(eventNames, name)
}

func (p *ParameterizerEventProcessor) challengeIDFromEvent(event *crawlermodel.Event) (*big.Int, error) {
	payload := event.EventPayload()
	challengeIDInterface, ok := payload["ChallengeID"]
	if !ok {
		return nil, errors.New("Unable to find the challenge ID in the payload")
	}
	return challengeIDInterface.(*big.Int), nil
}

// TODO: Move to go-common?
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Process processes Parameterizer Events into aggregated data
func (p *ParameterizerEventProcessor) Process(event *crawlermodel.Event) (bool, error) {
	if !p.isValidParameterizerContractEventName(event.EventType()) {
		return false, nil
	}

	var err error
	ran := true
	eventName := strings.Trim(event.EventType(), " _")

	var challengeID *big.Int
	if stringInSlice(eventName, paramChallengeEventNames) {
		challengeID, err = p.challengeIDFromEvent(event)
		if err != nil {
			return false, err
		}
	}

	// NOTE(IS): Only tracking challenge related data for Parameterizer contract for now.
	switch eventName {
	case "NewChallenge":
		log.Infof("Handling challenge %v\n", *challengeID)
		err = p.processParameterizerChallenge(event, challengeID)
	case "ChallengeFailed":
		log.Infof("Handling challenge %v\n", *challengeID)
		err = p.processChallengeFailed(event, challengeID)
	case "ChallengeSucceeded":
		log.Infof("Handling challenge %v\n", *challengeID)
		err = p.processChallengeSucceeded(event, challengeID)
	case "ReparameterizationProposal":
		log.Infof("Handling %v\n", eventName)
		err = p.processReparameterizationProposal(event)
	case "ProposalAccepted":
		log.Infof("Handling %v\n", eventName)
		err = p.processProposalAccepted(event)
	case "ProposalExpired":
		log.Infof("Handling %v\n", eventName)
		err = p.processProposalExpired(event)
	default:
		ran = false
	}
	return ran, err
}

func (p *ParameterizerEventProcessor) processReparameterizationProposal(event *crawlermodel.Event) error {
	return p.newParameterizationFromProposal(event)

}

func (p *ParameterizerEventProcessor) getPropIDFromEvent(event *crawlermodel.Event) ([32]byte, error) {
	payload := event.EventPayload()
	propID, ok := payload["PropID"]
	if !ok {
		return [32]byte{}, errors.New("Unable to get PropID in the payload")
	}
	return propID.([32]byte), nil
}

func (p *ParameterizerEventProcessor) processProposalAccepted(event *crawlermodel.Event) error {
	paramProposal, err := p.getExistingParameterProposal(event)
	if err != nil {
		return err
	}
	paramProposal.SetAccepted(true)
	return p.paramProposalPersister.UpdateParamProposal(paramProposal, []string{proposalAcceptedFieldName})
}

func (p *ParameterizerEventProcessor) processProposalExpired(event *crawlermodel.Event) error {
	paramProposal, err := p.getExistingParameterProposal(event)
	if err != nil {
		return err
	}
	paramProposal.SetExpired(true)
	return p.paramProposalPersister.UpdateParamProposal(paramProposal, []string{proposalExpiredFieldName})
}

func (p *ParameterizerEventProcessor) processParameterizerChallenge(event *crawlermodel.Event,
	challengeID *big.Int) error {
	_, err := p.newChallenge(event.ContractAddress(), challengeID)
	return err
}

func (p *ParameterizerEventProcessor) processChallengeFailed(event *crawlermodel.Event,
	challengeID *big.Int) error {

	pollIsPassed := true
	err := p.setPollIsPassedInPoll(challengeID, pollIsPassed)
	if err != nil {
		return fmt.Errorf("Error setting isPassed field in poll, err: %v", err)
	}
	return p.processChallengeResolution(event, challengeID, pollIsPassed)
}

func (p *ParameterizerEventProcessor) processChallengeSucceeded(event *crawlermodel.Event,
	challengeID *big.Int) error {

	pollIsPassed := false
	err := p.setPollIsPassedInPoll(challengeID, pollIsPassed)
	if err != nil {
		return fmt.Errorf("Error setting isPassed field in poll, err: %v", err)
	}
	return p.processChallengeResolution(event, challengeID, pollIsPassed)
}

func (p *ParameterizerEventProcessor) processChallengeResolution(event *crawlermodel.Event,
	challengeID *big.Int, pollIsPassed bool) error {
	payload := event.EventPayload()
	resolved := true
	totalTokens, ok := payload["TotalTokens"]
	if !ok {
		return errors.New("No totalTokens found")
	}
	pAddress := event.ContractAddress()

	existingChallenge, err := p.getExistingChallenge(challengeID, pAddress)
	if err != nil {
		return fmt.Errorf("Error getting existing challenge with id: %v. Err: %v",
			existingChallenge.ChallengeID(), err)
	}
	existingChallenge.SetResolved(resolved)
	existingChallenge.SetTotalTokens(totalTokens.(*big.Int))

	updatedFields := []string{resolvedFieldName, totalTokensFieldName}
	err = p.challengePersister.UpdateChallenge(existingChallenge, updatedFields)
	if err != nil {
		return fmt.Errorf("Error updating challenge %v, err: %v", existingChallenge.ChallengeID(), err)
	}

	return p.updateUserChallengeDataForChallengeRes(challengeID, pAddress, pollIsPassed)
}

func (p *ParameterizerEventProcessor) updateUserChallengeDataForChallengeRes(pollID *big.Int,
	pAddress common.Address, pollIsPassed bool) error {

	paramContract, err := contract.NewParameterizerContract(pAddress, p.client)
	if err != nil {
		return fmt.Errorf("Error calling parameterizer contract: %v", err)
	}

	userChallengeDataVotes, err := p.userChallengeDataPersister.UserChallengeDataByCriteria(
		&model.UserChallengeDataCriteria{
			PollID: pollID.Uint64(),
		},
	)

	if err != nil {
		if err == cpersist.ErrPersisterNoResults {
			log.Infof("No userChallengeData for %v", pollID)
			return nil
		}
		return fmt.Errorf("Error getting userchallengedata %v", err)
	}

	for _, userChallengeData := range userChallengeDataVotes {
		voter := userChallengeData.UserAddress()
		salt := userChallengeData.Salt()
		voterReward, err := paramContract.VoterReward(&bind.CallOpts{}, voter, pollID, salt)
		if err != nil {
			log.Errorf("Error getting voter reward %v", err)
			p.errRep.Error(errors.Wrap(err, "error getting voter reward"), nil)
		}
		var isVoterWinner bool
		if (pollIsPassed && userChallengeData.Choice().Int64() == 1) ||
			(!pollIsPassed && userChallengeData.Choice().Int64() == 0) {
			isVoterWinner = true
		} else {
			isVoterWinner = false
		}
		userChallengeData.SetVoterReward(voterReward)
		userChallengeData.SetIsVoterWinner(isVoterWinner)
		userChallengeData.SetPollIsPassed(pollIsPassed)
		updatedFields := []string{voterRewardFieldName, userChallengeIsPassedFieldName,
			isVoterWinnerFieldName}
		updateWithUserAddress := true
		latestVote := true

		err = p.userChallengeDataPersister.UpdateUserChallengeData(userChallengeData, updatedFields,
			updateWithUserAddress, latestVote)
		if err != nil {
			log.Errorf("Error updating poll in persistence: %v", err)
			p.errRep.Error(errors.Wrap(err, "error updating poll"), nil)
		}
	}
	return nil
}

func (p *ParameterizerEventProcessor) getExistingParameterProposal(event *crawlermodel.Event) (*model.ParameterProposal, error) {
	propID, err := p.getPropIDFromEvent(event)
	if err != nil {
		return nil, err
	}
	// get parameterization from db
	paramProposal, err := p.paramProposalPersister.ParamProposalByPropID(propID)
	if err != nil && err != cpersist.ErrPersisterNoResults {
		return nil, err
	}
	if err == cpersist.ErrPersisterNoResults {
		paramProposal, err = p.newParameterizationFromContract(event)
		if err != nil {
			return nil, err
		}
	}
	return paramProposal, nil
}

func (p *ParameterizerEventProcessor) newParameterizationFromProposal(event *crawlermodel.Event) error {
	payload := event.EventPayload()
	name, ok := payload["Name"]
	if !ok {
		return errors.New("No Name found")
	}
	value, ok := payload["Value"]
	if !ok {
		return errors.New("No Value found")
	}
	propID, ok := payload["PropID"]
	if !ok {
		return errors.New("No PropID found")
	}
	deposit, ok := payload["Deposit"]
	if !ok {
		return errors.New("No Deposit found")
	}
	appExpiry, ok := payload["AppEndDate"]
	if !ok {
		return errors.New("No AppEndDate found")
	}
	proposer, ok := payload["Proposer"]
	if !ok {
		return errors.New("No Proposer found")
	}
	// IF events are out of order this could be true
	accepted := false
	currentTime := ctime.CurrentEpochSecsInInt64()

	// calculate if expired
	var expired bool
	if currentTime < appExpiry.(*big.Int).Int64() {
		expired = false
	} else {
		expired = true
	}

	paramProposal := model.NewParameterProposal(&model.ParameterProposalParams{
		Name:              name.(string),
		Value:             value.(*big.Int),
		PropID:            propID.([32]byte),
		Deposit:           deposit.(*big.Int),
		AppExpiry:         appExpiry.(*big.Int),
		ChallengeID:       big.NewInt(0),
		Proposer:          proposer.(common.Address),
		Accepted:          accepted,
		Expired:           expired,
		LastUpdatedDateTs: currentTime,
	})

	// newParamProposal
	err := p.paramProposalPersister.CreateParameterProposal(paramProposal)
	return err
}

func (p *ParameterizerEventProcessor) newParameterizationFromContract(event *crawlermodel.Event) (*model.ParameterProposal, error) {
	payload := event.EventPayload()
	propID, ok := payload["PropID"]
	if !ok {
		return nil, errors.New("No PropID field found")
	}
	paramContract, err := contract.NewParameterizerContract(event.ContractAddress(), p.client)
	if err != nil {
		return nil, fmt.Errorf("Error calling parameterizer contract: %v", err)
	}
	prop, err := paramContract.Proposals(&bind.CallOpts{}, propID.([32]byte))
	if err != nil {
		return nil, fmt.Errorf("Error calling parameterizer contract: %v", err)
	}

	currentTime := ctime.CurrentEpochSecsInInt64()

	// calculate if expired
	var expired bool
	if currentTime < prop.AppExpiry.Int64() {
		expired = false
	} else {
		expired = true
	}
	// setting accepted to false for now
	accepted := false

	paramProposal := model.NewParameterProposal(&model.ParameterProposalParams{
		Name:              prop.Name,
		Value:             prop.Value,
		PropID:            propID.([32]byte),
		Deposit:           prop.Deposit,
		AppExpiry:         prop.AppExpiry,
		ChallengeID:       prop.ChallengeID,
		Proposer:          prop.Owner,
		Accepted:          accepted,
		Expired:           expired,
		LastUpdatedDateTs: currentTime,
	})
	return paramProposal, nil
}

func (p *ParameterizerEventProcessor) getExistingChallenge(challengeID *big.Int,
	pAddress common.Address) (*model.Challenge, error) {

	existingChallenge, err := p.challengePersister.ChallengeByChallengeID(int(challengeID.Int64()))
	if err != nil && err != cpersist.ErrPersisterNoResults {
		return nil, err
	}

	if existingChallenge == nil {
		existingChallenge, err = p.newChallenge(pAddress, challengeID)
		if err != nil {
			return nil, fmt.Errorf("Error persisting challenge: %v", err)
		}
	}
	return existingChallenge, nil
}

func (p *ParameterizerEventProcessor) newChallenge(pAddress common.Address,
	challengeID *big.Int) (*model.Challenge, error) {
	// NOTE(IS): Create a new challenge using data obtained by calling the smart contract.
	// For now, get all data from contract

	// NOTE(IS): No statement or requestAppealExpiry
	statement := ""
	requestAppealExpiry := big.NewInt(0)

	paramContract, err := contract.NewParameterizerContract(pAddress, p.client)
	if err != nil {
		return nil, fmt.Errorf("Error creating Parameterizer contract: err: %v", err)
	}
	challengeRes, err := paramContract.Challenges(&bind.CallOpts{}, challengeID)
	if err != nil {
		return nil, fmt.Errorf("Error calling function in TCR contract: err: %v", err)
	}

	challengeType := model.ParamProposalPollType
	// NOTE(IS): In parameterizer contract, there's no TotalTokens, but WinningTokens
	challenge := model.NewChallenge(
		challengeID,
		pAddress,
		statement,
		challengeRes.RewardPool,
		challengeRes.Challenger,
		challengeRes.Resolved,
		challengeRes.Stake,
		challengeRes.WinningTokens,
		requestAppealExpiry,
		challengeType,
		ctime.CurrentEpochSecsInInt64())

	err = p.challengePersister.CreateChallenge(challenge)
	return challenge, err
}

func (p *ParameterizerEventProcessor) setPollIsPassedInPoll(pollID *big.Int, isPassed bool) error {
	poll, err := p.pollPersister.PollByPollID(int(pollID.Int64()))
	if err != nil {
		return fmt.Errorf("Error getting poll from persistence: %v", err)
	}
	// TODO(IS): Shouldn't happen if all events are processed and in order, but create new poll if DNE
	poll.SetIsPassed(isPassed)
	updatedFields := []string{isPassedFieldName}

	err = p.pollPersister.UpdatePoll(poll, updatedFields)
	if err != nil {
		return fmt.Errorf("Error updating poll in persistence: %v", err)
	}

	return nil

}
