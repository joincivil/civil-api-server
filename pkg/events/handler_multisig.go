package events

import (
	"encoding/json"
	log "github.com/golang/glog"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/users"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/processor"
)

const (
	multiSigEventName = "MultiSig"
)

// NewMultiSigEventHandler creates a new MultiSigEventHandler
func NewMultiSigEventHandler(listingPersister model.ListingPersister, userService *users.UserService, channelService *channels.Service) *MultiSigEventHandler {
	return &MultiSigEventHandler{
		listingPersister: listingPersister,
		userService:      userService,
		channelService:   channelService,
	}
}

// MultiSigEventHandler handles Multi Sig events from the processor
// Implements EventHandler interface
type MultiSigEventHandler struct {
	listingPersister model.ListingPersister
	userService      *users.UserService
	channelService   *channels.Service
}

// Name returns the name of this particular event handler
func (t *MultiSigEventHandler) Name() string {
	return multiSigEventName
}

// Handle runs the logic to handle the event as appropriate for the event
func (t *MultiSigEventHandler) Handle(event []byte) (bool, error) {

	// Unmarshal into the processor pubsub message
	p := &processor.PubSubMultiSigMessage{}
	err := json.Unmarshal(event, p)
	if err != nil {
		return false, err
	}

	// get user
	user, err := t.userService.MaybeGetUser(users.UserCriteria{
		EthAddress: strings.ToLower(p.OwnerAddr),
	})

	// create user if necessary
	if err != nil || user == nil {
		user, err = t.userService.CreateUser(users.UserCriteria{
			EthAddress: strings.ToLower(p.OwnerAddr),
		})
		if err != nil {
			log.Errorf("Error creating user")
			return false, err
		}
	}

	// get listings
	listings, err := t.listingPersister.ListingsByOwnerAddress(common.HexToAddress(p.MultiSigAddr))
	if err != nil {
		log.Errorf("Error retrieving listings: %s", err)
		return false, err
	}

	for _, listing := range listings {
		// get channel
		channel, err := t.channelService.GetChannelByReference("newsroom", strings.ToLower(listing.ContractAddress().String()))

		if p.Action == processor.MultiSigOwnerAdded {
			if err != nil || channel == nil {
				// create channel with member
				_, err = t.channelService.CreateNewsroomChannel(
					user.UID,
					[]common.Address{common.HexToAddress(strings.ToLower(p.OwnerAddr))},
					channels.CreateNewsroomChannelInput{
						ContractAddress: strings.ToLower(listing.ContractAddress().String()),
					},
				)
				if err != nil {
					log.Errorf("Error creating channel")
					return false, err
				}
			} else { // channel already exists, add user as admin
				_, err := t.channelService.CreateChannelMember(user.UID, channel.ID)
				if err != nil {
					log.Errorf("Error creating channel member")
					return false, err
				}
			}
		} else if p.Action == processor.MultiSigOwnerRemoved {
			if err != nil || channel == nil {
				log.Errorf("No channel found when attempting to remove channel member")
				return false, err
			}
			err := t.channelService.DeleteChannelMember(user.UID, channel.ID)
			if err != nil {
				log.Errorf("Error deleting channel member")
				return false, err
			}
		}
	}

	return true, nil

}
