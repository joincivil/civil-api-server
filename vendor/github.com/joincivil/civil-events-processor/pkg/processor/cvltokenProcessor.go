package processor

import (
	"math/big"
	"strings"

	log "github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	commongen "github.com/joincivil/civil-events-crawler/pkg/generated/common"
	crawlermodel "github.com/joincivil/civil-events-crawler/pkg/model"
	cpersist "github.com/joincivil/go-common/pkg/persistence"

	"github.com/joincivil/civil-events-processor/pkg/model"
)

// NewCvlTokenEventProcessor is a convenience function to init an Event processor
func NewCvlTokenEventProcessor(client bind.ContractBackend,
	transferPersister model.TokenTransferPersister) *CvlTokenEventProcessor {
	return &CvlTokenEventProcessor{
		client:            client,
		transferPersister: transferPersister,
	}
}

// CvlTokenEventProcessor handles the processing of raw CvlToken events into aggregated data
// for use via the API.
type CvlTokenEventProcessor struct {
	client            bind.ContractBackend
	transferPersister model.TokenTransferPersister
}

// Process processes Newsroom Events into aggregated data
func (c *CvlTokenEventProcessor) Process(event *crawlermodel.Event) (bool, error) {
	if !c.isValidCvlTokenEventName(event.EventType()) {
		return false, nil
	}

	var err error
	ran := true
	eventName := strings.Trim(event.EventType(), " _")

	// Handling all the actionable events from the cvl token addressses
	switch eventName {
	// When a token transfer has occurred
	case "Transfer":
		log.Infof("Handling Token Transfer for %v\n", event.ContractAddress().Hex())
		err = c.processCvlTokenTransfer(event)

	default:
		ran = false
	}
	return ran, err
}

func (c *CvlTokenEventProcessor) isValidCvlTokenEventName(name string) bool {
	name = strings.Trim(name, " _")
	eventNames := commongen.EventTypesCVLTokenContract()
	return isStringInSlice(eventNames, name)
}

func (c *CvlTokenEventProcessor) processCvlTokenTransfer(event *crawlermodel.Event) error {
	payload := event.EventPayload()

	toAddress, ok := payload["To"]
	if !ok {
		return errors.New("No purchaser address found")
	}
	fromAddress, ok := payload["From"]
	if !ok {
		return errors.New("No source address found")
	}
	amount, ok := payload["Value"]
	if !ok {
		return errors.New("No amount found")
	}
	transferDate := event.Timestamp()

	paddr := toAddress.(common.Address)
	caddr := fromAddress.(common.Address)

	params := &model.TokenTransferParams{
		ToAddress:    paddr,
		FromAddress:  caddr,
		Amount:       amount.(*big.Int),
		TransferDate: transferDate,
		BlockNumber:  event.BlockNumber(),
		TxHash:       event.TxHash(),
		TxIndex:      event.TxIndex(),
		BlockHash:    event.BlockHash(),
		Index:        event.LogIndex(),
	}
	newPurchase := model.NewTokenTransfer(params)

	purchases, err := c.transferPersister.TokenTransfersByToAddress(paddr)
	if err != nil {
		if err != cpersist.ErrPersisterNoResults {
			return errors.WithMessage(err, "error retrieving token transfer")
		}
	}
	if len(purchases) > 0 {
		for _, purchase := range purchases {
			if purchase.Equals(newPurchase) {
				return errors.Errorf(
					"Token transfer already exists: %v, %v, %v, %v",
					purchase.ToAddress().Hex(),
					purchase.FromAddress().Hex(),
					purchase.Amount().Int64(),
					purchase.TransferDate(),
				)
			}
		}
	}

	return c.transferPersister.CreateTokenTransfer(newPurchase)
}
