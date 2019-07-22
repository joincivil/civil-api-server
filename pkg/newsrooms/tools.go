package newsrooms

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/go-common/pkg/generated/contract"
	"github.com/joincivil/go-common/pkg/newsroom"
	"go.uber.org/fx"
)

// Tools provides utilities to create and manage newsrooms
type Tools struct {
	ethHelper       *eth.Helper
	cvl             *contract.CVLTokenContract
	newsroomService *newsroom.Service
	txListener      *eth.TxListener
}

// ToolDeps contains the fields needed to instatiate a new Tools instance
type ToolDeps struct {
	fx.In
	EthHelper       *eth.Helper
	CVL             *contract.CVLTokenContract
	NewsroomService *newsroom.Service
	TxListener      *eth.TxListener
}

// NewTools builds a new Tools instance
func NewTools(deps ToolDeps) *Tools {

	return &Tools{
		ethHelper:       deps.EthHelper,
		cvl:             deps.CVL,
		newsroomService: deps.NewsroomService,
		txListener:      deps.TxListener,
	}
}

// CreateAndApply creates a newsroom smart contract and applies to the TCR on their behalf
// This sends progress updates into the provided channel and provides some more flexibility than just writing the updates to the log
// We can create a GraphQL subscription
func (t *Tools) CreateAndApply(updates chan string, charter *newsroom.Charter, applicationTokens *big.Int) error {
	defer close(updates)

	updates <- fmt.Sprintf("Publishing Charter for %v", charter.Name)
	hash, err := t.newsroomService.PublishCharter(*charter, true)
	if err != nil {
		fmt.Printf("Error publishing charter: %v", err)
		close(updates)
		return err
	}
	updates <- fmt.Sprintf("Successfully published Charter, hash is: %v", hash)

	// create newsroom from factory
	txHash, err := t.ensureTransaction(t.newsroomService.CreateNewsroom(charter.Name, hash))
	if err != nil {
		fmt.Printf("Error! %v", err)
		fmt.Printf("Couldn't get create newsroom: %v", err)
		return err
	}

	updates <- fmt.Sprintf("Created Newsroom (%v)", txHash.String())

	// retrieve the newsroom address
	newsroomAddress, err := t.newsroomService.GetNewsroomAddressFromTransaction(txHash)
	if err != nil {
		fmt.Printf("Couldn't get Newsroom Address: %v", err)

		return err
	}
	updates <- fmt.Sprintf("Newsroom Address: %v", newsroomAddress.String())

	// retrieve the multisig address
	multisigAddress, err := t.newsroomService.GetOwner(newsroomAddress)
	if err != nil {
		fmt.Printf("Couldn't get Newsroom Multisig: %v", err)

		return err
	}
	updates <- fmt.Sprintf("Multisig Address: %v", multisigAddress.String())

	// fund newsroom multisig
	tx, err := t.cvl.Transfer(t.ethHelper.Transact(), multisigAddress, applicationTokens)
	if err != nil {
		fmt.Printf("error with transfer: %v", err)

		return err
	}
	txHash, err = t.ensureTransaction(tx.Hash(), err)
	if err != nil {
		return err
	}
	updates <- fmt.Sprintf("Funded multisig (%v)", txHash.String())

	// approve TCR token transfer
	txHash, err = t.ensureTransaction(
		t.newsroomService.AdminApproveTCRTokenTransfer(newsroomAddress, applicationTokens),
	)
	if err != nil {
		return err
	}
	updates <- fmt.Sprintf("Approved tokens for TCR application (%v)", txHash.String())

	// apply to TCR
	txHash, err = t.ensureTransaction(
		t.newsroomService.AdminApplyToTCR(newsroomAddress, applicationTokens),
	)
	if err != nil {
		return err
	}
	updates <- fmt.Sprintf("Application to TCR successful (%v)", txHash.String())

	updates <- fmt.Sprintf("JOB COMPLETE")

	return nil
}

// HandoffNewsroom adds a new owner to the newsroom multisig, removes the hot wallet, and optionally adds the foundation backup multisig for recovery
func (t *Tools) HandoffNewsroom(updates chan string, newsroomAddress common.Address, newOwners []common.Address) error {
	defer close(updates)

	for _, newOwner := range newOwners {
		updates <- fmt.Sprintf("Adding new owner to multisig %v", newOwner.String())
		txHash, err := t.ensureTransaction(
			t.newsroomService.AdminAddMultisigOwner(newsroomAddress, newOwner),
		)
		if err != nil {
			return err
		}
		updates <- fmt.Sprintf("Added owner to multisig (%v)", txHash.String())
	}

	// remove hot wallet
	updates <- fmt.Sprintf("Removing hot wallet from multisig %v", t.ethHelper.Auth.From)
	txHash, err := t.ensureTransaction(
		t.newsroomService.AdminRemoveMultisigOwner(newsroomAddress, t.ethHelper.Auth.From),
	)
	if err != nil {
		return err
	}
	updates <- fmt.Sprintf("Removed hot wallet from multisig (%v)", txHash.String())

	return nil
}

func (t *Tools) ensureTransaction(txHash common.Hash, err error) (common.Hash, error) {
	if err != nil {
		fmt.Printf("transaction returned an error: %v\n", err)
		return common.Hash{}, err
	}

	sub, err := t.txListener.StartListener(txHash.String())
	if err != nil {
		fmt.Printf("error starting tx listener: %v\n", err)
		return common.Hash{}, err
	}

	for range sub.Updates {
	}

	receipt, err := t.ethHelper.Blockchain.(ethereum.TransactionReader).TransactionReceipt(context.Background(), txHash)
	if err != nil {
		fmt.Printf("error getting tx receipt: %v\n", err)
		return common.Hash{}, err
	}

	if receipt.Status != 1 {
		fmt.Printf("transaction failed\n")
		return common.Hash{}, err
	}

	return txHash, nil
}
