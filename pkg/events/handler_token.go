package events

import (
	"encoding/json"

	log "github.com/golang/glog"

	"github.com/ethereum/go-ethereum/common"

	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/processor"
)

const (
	transferEventName = "Transfer"
)

// NewCvlTokenTransferEventHandler creates a new CvlTokenTransferEventHandler
func NewCvlTokenTransferEventHandler(tokenPersister model.TokenTransferPersister,
	userService *users.UserService, tokenSaleAddresses []common.Address,
	registryListID string) *CvlTokenTransferEventHandler {
	return &CvlTokenTransferEventHandler{
		tokenPersister:     tokenPersister,
		userService:        userService,
		tokenSaleAddresses: tokenSaleAddresses,
		registryListID:     registryListID,
	}
}

// CvlTokenTransferEventHandler handles TokenTransfer events from the processor
// Implements EventHandler interface
type CvlTokenTransferEventHandler struct {
	tokenPersister     model.TokenTransferPersister
	userService        *users.UserService
	tokenSaleAddresses []common.Address
	registryListID     string
}

// Name returns the name of this particular event handler
func (t *CvlTokenTransferEventHandler) Name() string {
	return transferEventName
}

// Handle runs the logic to handle the event as appropriate for the event
func (t *CvlTokenTransferEventHandler) Handle(event []byte) (bool, error) {

	log.Warningf("Handle()")
	// Unmarshal into the processor pubsub message
	p := &processor.PubSubMessage{}
	err := json.Unmarshal(event, p)
	if err != nil {
		return false, err
	}
	transfers, err := t.tokenPersister.TokenTransfersByTxHash(common.HexToHash(p.TxHash))
	if err != nil {
		return false, err
	}

	ran := false

Loop:
	for _, transfer := range transfers {
		log.Warningf("Handle() 2")
		addr := transfer.ToAddress().Hex()

		if !t.isTokenSaleAddr(transfer.FromAddress()) {
			continue
		}

		if addr != "" {
			log.Infof("transfer = %v", transfer)
			// Try to find the user
			user, err := t.userService.MaybeGetUser(users.UserCriteria{
				EthAddress: addr,
			})
			if err != nil {
				log.Errorf("Error retrieving user: err: %v", err)
				continue
			}
			if user == nil {
				log.Errorf("User was not found: %v", addr)
				continue
			}

			// User found, add them to the registry email list
			if user.Email != "" {
				err = t.addToRegistryAlertsList(user.Email)
				if err != nil {
					log.Errorf("Error adding to registry list %v: err: %v", t.registryListID, err)
				}
			}

			ran = true
			break Loop
		}
	}
	log.Infof("Done handling token transfer event: ran: %v", ran)
	return ran, nil
}

func (t *CvlTokenTransferEventHandler) isTokenSaleAddr(addr common.Address) bool {
	for _, taddr := range t.tokenSaleAddresses {
		if addr.Hex() == taddr.Hex() {
			return true
		}
	}
	return false
}

func (t *CvlTokenTransferEventHandler) addToRegistryAlertsList(emailAddress string) error {
	if t.registryListID != "" {
		err := utils.AddToRegistryAlertsList(emailAddress, t.registryListID)
		if err != nil {
			return err
		}
	}
	return nil
}
