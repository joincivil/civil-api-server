package tokencontroller_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/go-common/pkg/generated/contract"
)

func TestAddToCivilian(t *testing.T) {
	ethHelper, err := eth.NewSimulatedBackendHelper()
	blockchain := ethHelper.Blockchain.(*backends.SimulatedBackend)
	if err != nil {
		t.Fatalf("error constructing blockchain helper: err: %v", err)
	}

	tokenControllerAddr, _, _, err := deployCivilTokenController(ethHelper)
	if err != nil {
		t.Fatalf("error deploying token controller: err: %v", err)
	}
	blockchain.Commit()

	contractAddresses := eth.DeployerContractAddresses{
		CivilTokenController: tokenControllerAddr,
	}
	service, err := tokencontroller.NewService(contractAddresses, ethHelper)
	if err != nil {
		t.Fatalf("error creating token controller service: %v", err)
	}

	alice := ethHelper.Accounts["alice"]
	_, err = service.AddToCivilians(alice.Address)
	if err != nil {
		t.Fatalf("error adding address to whitelist: %v", err)
	}
	t.Log("all done!")

}

func deployCivilTokenController(helper *eth.Helper) (common.Address, *types.Transaction, *bind.BoundContract, error) {
	libAddress, _, _, err := contract.DeployMessagesAndCodesContract(helper.Auth, helper.Blockchain)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	return eth.DeployContractWithLinks(
		helper.Auth, helper.Blockchain,
		contract.CivilTokenControllerContractABI,
		contract.CivilTokenControllerContractBin,
		map[string]common.Address{"MessagesAndCodes": libAddress},
	)

}
