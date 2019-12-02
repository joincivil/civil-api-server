package newsrooms

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/pkg/errors"
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
	ethHelper         *eth.Helper
	cvl               *contract.CVLTokenContract
	newsroomService   *newsroom.Service
	txListener        *eth.TxListener
	nrsignupService   *nrsignup.Service
	channelService    *channels.Service
	applicationTokens *big.Int
	rescueAddress     common.Address
}

// ToolsConfig contains the fields needed to configure ToolDeps
type ToolsConfig struct {
	ApplicationTokens int64 // input as CVL not CVLWei
	RescueAddress     common.Address
}

// ToolDeps contains the fields needed to instatiate a new Tools instance
type ToolDeps struct {
	fx.In
	EthHelper       *eth.Helper
	CVL             *contract.CVLTokenContract
	NewsroomService *newsroom.Service
	TxListener      *eth.TxListener
	NrSignupService *nrsignup.Service
	ChannelService  *channels.Service
	Config          ToolsConfig
}

// NewTools builds a new Tools instance
func NewTools(deps ToolDeps) (*Tools, error) {

	if deps.Config.ApplicationTokens == 0 {
		return nil, errors.New("application tokens parameter not provided")
	}
	if (deps.Config.RescueAddress == common.Address{}) {
		return nil, errors.New("rescue address parameter not provided")
	}

	applicationTokens := big.NewInt(deps.Config.ApplicationTokens)
	applicationTokens = applicationTokens.Mul(applicationTokens, big.NewInt(1e18))

	return &Tools{
		ethHelper:         deps.EthHelper,
		cvl:               deps.CVL,
		channelService:    deps.ChannelService,
		newsroomService:   deps.NewsroomService,
		txListener:        deps.TxListener,
		nrsignupService:   deps.NrSignupService,
		applicationTokens: applicationTokens,
		rescueAddress:     deps.Config.RescueAddress,
	}, nil
}

const (
	// UpdatesPublishCharterStart is fired when a charter is starting to be published to IPFS
	UpdatesPublishCharterStart = "UpdatesPublishCharterStart"
	// UpdatesPublishCharterEnd is fired when the publish is complete
	UpdatesPublishCharterEnd = "UpdatesPublishCharterEnd"
	// UpdatesCreateNewsroomStart is fired when the newsroom contract creation is started
	UpdatesCreateNewsroomStart = "UpdatesCreateNewsroomStart"
	// UpdatesCreateNewsroomEnd is fired when the newsroom contract tx is mined
	UpdatesCreateNewsroomEnd = "UpdatesCreateNewsroomEnd"
	// UpdatesResumeCreateNewsroomStart is fired when resuming listening to a newsroom tx
	UpdatesResumeCreateNewsroomStart = "UpdatesResumeCreateNewsroomStart"
	// UpdatesResumeCreateNewsroomEnd is fired when the newsroom contract tx is mined
	UpdatesResumeCreateNewsroomEnd = "UpdatesResumeCreateNewsroomEnd"
	// UpdatesSavedNewsroomAddress is fired when the newsroom contract address is saved to nrsignup
	UpdatesSavedNewsroomAddress = "UpdatesSavedNewsroomAddress"
	// UpdatesSavedDeployTx is fired when the TCR deploy tx is saved to nrsignup
	UpdatesSavedDeployTx = "UpdatesSavedDeployTx"
	// UpdatesFundMultisigBalance is fired when the mulitisg has the adequate balance
	UpdatesFundMultisigBalance = "UpdatesFundMultisigBalance"
	// UpdatesFundMultisigStart is fired when starting a funding tx
	UpdatesFundMultisigStart = "UpdatesFundMultisigStart"
	// UpdatesFundMultisigEnd is fired when the multisig is funded
	UpdatesFundMultisigEnd = "UpdatesFundMultisigEnd"
	// UpdatesTCRApproveStart is fired when the tx to approve CVL token transfers is submitted
	UpdatesTCRApproveStart = "UpdatesTCRApproveStart"
	// UpdatesTCRApproveEnd is fired when the tx to approve CVL token transfers is mined
	UpdatesTCRApproveEnd = "UpdatesTCRApproveEnd"
	// UpdatesTCRApplyStart is fired when the TCR application tx si submitted
	UpdatesTCRApplyStart = "UpdatesTCRApplyStart"
	// UpdatesTCRApplyEnd is fired when the TCR application tx is mined
	UpdatesTCRApplyEnd = "UpdatesTCRApplyEnd"
	// UpdatesMultisigAddOwnerStart is fired when the tx to add an owner is submitted
	UpdatesMultisigAddOwnerStart = "UpdatesMultisigAddOwnerStart"
	// UpdatesMultisigAddOwnerEnd is fired when the tx to remove an owner is mined
	UpdatesMultisigAddOwnerEnd = "UpdatesMultisigAddOwnerEnd"
	// UpdatesMultisigRemoveOwnerStart is fired when the tx to remove an owner is submitted
	UpdatesMultisigRemoveOwnerStart = "UpdatesMultisigRemoveOwnerStart"
	// UpdatesMultisigRemoveOwnerEnd is fired when the tx to remove an owner is mined
	UpdatesMultisigRemoveOwnerEnd = "UpdatesMultisigRemoveOwnerEnd"
	// UpdatesNewsroomChannelCreated is fired when a newsromo channel is created
	UpdatesNewsroomChannelCreated = "UpdatesNewsroomChannelCreated"
	// UpdatesNewsroomDataDeleted is fired when nrsignup data is deleted
	UpdatesNewsroomDataDeleted = "UpdatesNewsroomDataDeleted"
	// UpdatesDone is fired when everything is done
	UpdatesDone = "UpdatesDone"
)

func send(updates chan string, key string, value string) {
	updates <- fmt.Sprintf("%v:%v", key, value)
}

// FastPassNewsroom will create a newsroom contract, fund the multisig, apply to TCR, add a rescue address, and create newsroom channels
// requires that the nrsignup for newsroomOwnerUID has an approved grant
func (t *Tools) FastPassNewsroom(updates chan string, newsroomOwnerUID string) error {

	newsroomData, err := t.nrsignupService.RetrieveUserJSONData(newsroomOwnerUID)
	if err != nil {
		fmt.Printf("Error getting newsroom data: %v", err)
		return err
	}

	// sanity checks before we begin
	if len(newsroomData.Charter.Roster) == 0 || newsroomData.Charter.Roster[0].EthAddress == "" {
		return errors.New("cannot start FastPass until the first member of the Roster has set their ETH address")
	}
	if newsroomData.GrantApproved != nil && !*newsroomData.GrantApproved {
		return errors.New("cannot start FastPass until the grant has been approved")
	}

	// create the newsroom if necessary
	newsroomAddress := common.HexToAddress(newsroomData.NewsroomAddress)
	if newsroomData.NewsroomAddress == "" {
		// create the newsroom contract if it doesn't exist yet, or wait for the tx to finish
		newsroomTxHash, err := t.CreateOrContinueNewsroomContract(updates, newsroomOwnerUID, newsroomData)
		if err != nil {
			fmt.Printf("Error creating or continuing newsroom contract creation: %v", err)
			return err
		}

		// save the newsroom contract address if it hasn't been saved yet
		newsroomAddress, err = t.SaveNewsroomContractAddress(updates, newsroomOwnerUID, newsroomData, newsroomTxHash)
		if err != nil {
			fmt.Printf("Error creating or continuing newsroom contract creation: %v", err)
			return err
		}
	}

	// fund the multisig if necessary
	err = t.FundMultisig(updates, newsroomAddress, newsroomData)
	if err != nil {
		return errors.Wrap(err, "Error funding multisig")
	}

	// apply to TCR
	err = t.ApplyToTCR(updates, newsroomOwnerUID, newsroomAddress, newsroomData)
	if err != nil {
		return errors.Wrap(err, "Error applying to TCR")
	}

	// add Charter Member 1 to Roster
	addr := common.HexToAddress(newsroomData.Charter.Roster[0].EthAddress)
	err = t.AddMultisigOwner(updates, newsroomAddress, addr)
	if err != nil {
		return errors.Wrap(err, "Error adding roster member to multisig")
	}

	// add foundation multisig to roster
	err = t.AddMultisigOwner(updates, newsroomAddress, t.rescueAddress)
	if err != nil {
		return errors.Wrap(err, "Error adding roster member to multisig")
	}

	// remove fastpass user from the roster
	err = t.RemoveMultisigOwner(updates, newsroomAddress, t.ethHelper.Auth.From)
	if err != nil {
		return errors.Wrap(err, "Error removing fastpass user from multisig")
	}

	// create channel for newsroom
	channel, err := t.channelService.CreateNewsroomChannel(
		newsroomOwnerUID,
		[]common.Address{addr},
		channels.CreateNewsroomChannelInput{ContractAddress: newsroomAddress.String()},
	)
	if err != nil {
		return errors.Wrap(err, "Error creating channel")
	}
	send(updates, UpdatesNewsroomChannelCreated, channel.ID)

	// delete nrsignup
	err = t.nrsignupService.DeleteNewsroomData(newsroomOwnerUID)
	if err != nil {
		return err
	}
	send(updates, UpdatesNewsroomDataDeleted, newsroomOwnerUID)

	// all done!
	send(updates, UpdatesDone, ":-)")

	return nil
}

// CreateOrContinueNewsroomContract deploys a new newsroom contract if newsroomData.NewsroomDeployTx is not set
// otherwise it will continue listening for the tx and return when complete
func (t *Tools) CreateOrContinueNewsroomContract(updates chan string, newsroomOwnerUID string, newsroomData *nrsignup.SignupUserJSONData) (common.Hash, error) {
	charter := newsroomData.Charter
	if newsroomData.NewsroomDeployTx == "" {

		// publish charter to IPFS
		send(updates, UpdatesPublishCharterStart, charter.Name)
		hash, err := t.newsroomService.PublishCharter(*charter, true)
		if err != nil {
			fmt.Printf("Error publishing charter: %v", err)
			return common.Hash{}, err
		}
		send(updates, UpdatesPublishCharterEnd, hash)

		// create newsroom from factory
		send(updates, UpdatesCreateNewsroomStart, charter.Name)
		txHash, err := t.newsroomService.CreateNewsroom(charter.Name, hash)
		if err != nil {
			return common.Hash{}, errors.Wrap(err, "could not create newsroom")
		}

		// save transaction id to nrsignup data
		err = t.nrsignupService.SaveNewsroomDeployTxHash(newsroomOwnerUID, txHash.String())
		if err != nil {
			fmt.Printf("Could not save deploy tx hash (%v): %v", txHash, err)
			return common.Hash{}, err
		}
		send(updates, UpdatesSavedDeployTx, txHash.String())

		_, err = t.ensureTransaction(txHash, err)
		if err != nil {
			fmt.Printf("Could not create newsroom: %v", err)
			return common.Hash{}, errors.Wrap(err, "Error waiting for transaction to finish")
		}
		send(updates, UpdatesCreateNewsroomEnd, txHash.String())

		newsroomData.NewsroomDeployTx = txHash.String()

		return txHash, nil
	}

	send(updates, UpdatesResumeCreateNewsroomStart, newsroomData.NewsroomDeployTx)
	txHash, err := t.ensureTransaction(common.HexToHash(newsroomData.NewsroomDeployTx), nil)
	if err != nil {
		fmt.Printf("error resuming transaction with hash (%v): %v", txHash, err)
		return common.Hash{}, err
	}
	send(updates, UpdatesResumeCreateNewsroomEnd, newsroomData.NewsroomDeployTx)
	return txHash, nil
}

// SaveNewsroomContractAddress retrieves the newsroom contract address and saves it to nrsignup
func (t *Tools) SaveNewsroomContractAddress(updates chan string, newsroomOwnerUID string, newsroomData *nrsignup.SignupUserJSONData, newsroomTxHash common.Hash) (common.Address, error) {
	// retrieve the newsroom address
	fmt.Printf("newsroomTxHash %v\n", newsroomTxHash.String())
	newsroomAddress, err := t.newsroomService.GetNewsroomAddressFromTransaction(newsroomTxHash)
	if err != nil {
		fmt.Printf("Couldn't get Newsroom Address: %v", err)

		return common.Address{}, err
	}
	// save the newsroom address
	err = t.nrsignupService.SaveNewsroomAddress(newsroomOwnerUID, newsroomAddress.String())
	if err != nil {
		fmt.Printf("error saving newsroom address %v", err)
		return common.Address{}, err
	}
	send(updates, UpdatesSavedNewsroomAddress, newsroomAddress.String())

	return newsroomAddress, nil
}

// FundMultisig transfers CVL from the API eth account into the multisig if the newsroom hasn't applied yet
func (t *Tools) FundMultisig(updates chan string, newsroomAddress common.Address, newsroomData *nrsignup.SignupUserJSONData) error {
	// do not fund if the newsroom has already applied to the TCR
	if newsroomData.TcrApplyTx != "" {
		return nil
	}

	// retrieve the multisig address
	multisigAddress, err := t.newsroomService.GetOwner(newsroomAddress)
	if err != nil {
		return errors.Wrap(err, "could not get newsroom multisig")
	}

	multisigBalance, err := t.cvl.BalanceOf(&bind.CallOpts{}, multisigAddress)
	if err != nil {
		return errors.Wrap(err, "could not get multisig balance")
	}

	send(updates, UpdatesFundMultisigBalance, multisigBalance.String())

	// fund if multisigBalance is less than tokens needed to apply
	if multisigBalance.Cmp(t.applicationTokens) == -1 {
		// fund newsroom multisig
		send(updates, UpdatesFundMultisigStart, multisigAddress.String())
		tx, err := t.cvl.Transfer(t.ethHelper.Transact(), multisigAddress, t.applicationTokens)
		if err != nil {
			return errors.Wrap(err, "error transfering funds to multisig")
		}
		txHash, err := t.ensureTransaction(tx.Hash(), err)
		if err != nil {
			return err
		}
		send(updates, UpdatesFundMultisigEnd, txHash.String())
	}

	return nil
}

// ApplyToTCR submits the newsrooms application to the TCR
func (t *Tools) ApplyToTCR(updates chan string, newsroomOwnerUID string, newsroomAddress common.Address, newsroomData *nrsignup.SignupUserJSONData) error {
	txHash := common.HexToHash(newsroomData.TcrApplyTx)
	if newsroomData.TcrApplyTx == "" {
		var err error
		// approve TCR token transfer
		send(updates, UpdatesTCRApproveStart, newsroomAddress.String())
		txHash, err = t.ensureTransaction(
			t.newsroomService.AdminApproveTCRTokenTransfer(newsroomAddress, t.applicationTokens),
		)
		if err != nil {
			return errors.Wrap(err, "error approving token transfer")
		}
		send(updates, UpdatesTCRApproveEnd, newsroomAddress.String())

		// apply to TCR
		txHash, err = t.newsroomService.AdminApplyToTCR(newsroomAddress, t.applicationTokens)
		if err != nil {
			return errors.Wrap(err, "error applying to TCR")
		}
		send(updates, UpdatesTCRApplyStart, txHash.String())
		err = t.nrsignupService.SaveNewsroomApplyTxHash(newsroomOwnerUID, txHash.String())
		if err != nil {
			return err
		}
	}

	txHash, err := t.ensureTransaction(txHash, nil)
	if err != nil {
		return errors.Wrap(err, "Error waiting for TCR application transaction")
	}
	send(updates, UpdatesTCRApplyEnd, txHash.String())

	return nil

}

// AddMultisigOwner adds an owner to the newsroom's multisig
func (t *Tools) AddMultisigOwner(updates chan string, newsroomAddress common.Address, memberAddress common.Address) error {

	txHash, err := t.newsroomService.AdminAddMultisigOwner(newsroomAddress, memberAddress)
	if err != nil {
		return errors.Wrap(err, "Error adding member to multisig")
	}
	send(updates, UpdatesMultisigAddOwnerStart, fmt.Sprintf("%v:%v:%v", newsroomAddress.String(), memberAddress.String(), txHash.String()))
	_, err = t.ensureTransaction(txHash, nil)
	if err != nil {
		return errors.Wrap(err, "Error adding member to multisig")
	}
	send(updates, UpdatesMultisigAddOwnerEnd, fmt.Sprintf("%v:%v:%v", newsroomAddress.String(), memberAddress.String(), txHash.String()))

	return nil
}

// RemoveMultisigOwner removes an owner From the newsroom's multisig
func (t *Tools) RemoveMultisigOwner(updates chan string, newsroomAddress common.Address, memberAddress common.Address) error {

	txHash, err := t.newsroomService.AdminRemoveMultisigOwner(newsroomAddress, memberAddress)
	if err != nil {
		return errors.Wrap(err, "Error removing member from multisig")
	}
	send(updates, UpdatesMultisigRemoveOwnerStart, fmt.Sprintf("%v:%v:%v", newsroomAddress.String(), memberAddress.String(), txHash.String()))
	_, err = t.ensureTransaction(txHash, nil)
	if err != nil {
		return errors.Wrap(err, "Error removing member from multisig")
	}
	send(updates, UpdatesMultisigRemoveOwnerEnd, fmt.Sprintf("%v:%v:%v", newsroomAddress.String(), memberAddress.String(), txHash.String()))

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
