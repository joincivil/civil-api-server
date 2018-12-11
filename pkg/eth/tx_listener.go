package eth

import (
	"context"
	"fmt"
	"time"

	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joincivil/civil-api-server/pkg/jobs"
)

// TxListener provides methods to interact with Ethereum transactions
type TxListener struct {
	blockchain *ethclient.Client
	jobs       jobs.JobService
}

// NewTxListener creates a new TransactionService instance
func NewTxListener(blockchain *ethclient.Client, jobs jobs.JobService) *TxListener {
	return &TxListener{blockchain, jobs}
}

// StartListener begins listening for an ethereum transaction
func (t *TxListener) StartListener(txID string) (*jobs.Subscription, error) {
	jobID := fmt.Sprintf("TxListener-%v", txID)
	job, err := t.jobs.StartJob(jobID, func(updates chan<- string) {
		t.PollForTx(txID, updates)
	})
	if err != nil && err != jobs.ErrJobAlreadyExists {
		return nil, err
	}

	if err == jobs.ErrJobAlreadyExists {
		job, err = t.jobs.GetJob(jobID)
		if err != nil {
			return nil, err
		}
	}

	subscription := job.Subscribe()

	return subscription, nil
}

// StopSubscription will stop subscribing to job updates
// this will not cancel the actual job
func (t *TxListener) StopSubscription(receipt *jobs.Subscription) error {
	return t.jobs.StopSubscription(receipt)
}

// PollForTx will continuously poll until a transaction is complete
func (t *TxListener) PollForTx(txID string, updates chan<- string) {

	hash := common.HexToHash(txID)

	ticker := time.NewTicker(time.Millisecond * 500)

	for range ticker.C {
		isPending, err := t.checkTx(hash)
		if err != nil {
			updates <- "Error! " + err.Error()
			return
		}
		if !isPending {
			updates <- "Transaction complete!"
			return
		}
		updates <- "Transaction is pending"
	}

}

func (t *TxListener) checkTx(hash common.Hash) (bool, error) {
	_, isPending, err := t.blockchain.TransactionByHash(context.Background(), hash)
	if err != nil {
		log.Printf("error: %v\n", err)
		return false, err
	}

	return isPending, nil
}
