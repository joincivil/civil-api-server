package events

import (
	"encoding/json"
	log "github.com/golang/glog"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/joincivil/civil-api-server/pkg/channels"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/processor"
)

const (
	governanceEventName = "Governance"
)

// NewGovernanceEventHandler creates a new GovernanceEventHandler
func NewGovernanceEventHandler(governanceEventPersister model.GovernanceEventPersister, listingPersister model.ListingPersister, channelService *channels.Service) *GovernanceEventHandler {
	return &GovernanceEventHandler{
		governanceEventPersister: governanceEventPersister,
		listingPersister:         listingPersister,
		channelService:           channelService,
	}
}

// GovernanceEventHandler handles Governance events from the processor
// Implements EventHandler interface
type GovernanceEventHandler struct {
	governanceEventPersister model.GovernanceEventPersister
	listingPersister         model.ListingPersister
	channelService           *channels.Service
}

// Name returns the name of this particular event handler
func (t *GovernanceEventHandler) Name() string {
	return governanceEventName
}

// Handle runs the logic to handle the event as appropriate for the event
func (t *GovernanceEventHandler) Handle(event []byte) (bool, error) {

	log.Infof("Governance handler 1")
	// Unmarshal into the processor pubsub message
	p := &processor.PubSubMessage{}
	log.Infof("Governance handler 1=2")
	err := json.Unmarshal(event, p)
	if err != nil {
		return false, err
	}
	log.Infof("Governance handler 3. try get events for p.TxHash: %s", p.TxHash)
	governanceEvents, err := t.governanceEventPersister.GovernanceEventsByTxHash(common.HexToHash(p.TxHash))
	if err != nil {
		return false, err
	}

	log.Infof("Governance handler 4")
	for _, g := range governanceEvents {
		log.Infof("Governance handler 5")
		if g.GovernanceEventType() == "ApplicationWhitelisted" {
			log.Infof("Got ApplicationWhitelisted")
			var listingAddress string
			for key, val := range g.Metadata() {
				log.Infof("Here in Metadata. key: %s", key)
				if key == "ListingAddress" {
					log.Infof("Found ListingAddress")
					listingAddress = strings.ToLower(val.(string))
				}
			}
			if listingAddress != "" {
				channel, err := t.channelService.GetChannelByReference("newsroom", listingAddress)
				if err != nil {
					return false, err
				}
				listing, err := t.listingPersister.ListingByAddress(common.HexToAddress(listingAddress))
				if err != nil {
					return false, err
				}
				name := listing.Name()
				log.Infof("Go Set Handle. name: %s", name)
				_, err = t.channelService.SetNewsroomHandleOnAccepted(channel.ID, name)
				if err != nil {
					return false, err
				}
			}
		} else if g.GovernanceEventType() == "ListingRemoved" || g.GovernanceEventType() == "ListingWithdrawn" {
			var listingAddress string
			for key, val := range g.Metadata() {
				if key == "ListingAddress" {
					listingAddress = strings.ToLower(val.(string))
				}
			}
			if listingAddress != "" {
				channel, err := t.channelService.GetChannelByReference("newsroom", listingAddress)
				if err != nil {
					return false, err
				}

				_, err = t.channelService.ClearNewsroomHandleOnRemoved(channel.ID)
				if err != nil {
					return false, err
				}
				return true, nil
			}
		}
	}

	return false, nil
}
