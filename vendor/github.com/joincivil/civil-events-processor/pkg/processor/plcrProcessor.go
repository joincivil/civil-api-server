package processor

import (
	"fmt"
	"math/big"
	"strings"

	log "github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	commongen "github.com/joincivil/civil-events-crawler/pkg/generated/common"
	crawlermodel "github.com/joincivil/civil-events-crawler/pkg/model"

	"github.com/joincivil/civil-events-processor/pkg/model"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
	ctime "github.com/joincivil/go-common/pkg/time"
)

const (
	votesForFieldName     = "VotesFor"
	votesAgainstFieldName = "VotesAgainst"
)

// NewPlcrEventProcessor is a convenience function to init an EventProcessor
func NewPlcrEventProcessor(client bind.ContractBackend, pollPersister model.PollPersister) *PlcrEventProcessor {
	return &PlcrEventProcessor{
		client:        client,
		pollPersister: pollPersister,
	}
}

// PlcrEventProcessor handles the processing of raw events into aggregated data
// for use via the API.
type PlcrEventProcessor struct {
	client        bind.ContractBackend
	pollPersister model.PollPersister
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
	case "VoteRevealed":
		pollID, pollIDerr := p.pollIDFromEvent(event)
		if pollIDerr != nil {
			log.Infof("Error retrieving pollID: %v", err)
		}
		log.Infof("Handling VoteRevealed for pollID %v\n", pollID)
		err = p.processVoteRevealed(event, pollID)
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

func (p *PlcrEventProcessor) processVoteRevealed(event *crawlermodel.Event,
	pollID *big.Int) error {
	payload := event.EventPayload()
	choice, ok := payload["Choice"]
	if !ok {
		return errors.New("No choice found")
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
	return err
}
