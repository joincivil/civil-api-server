package storefront_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"

	"github.com/joincivil/go-common/pkg/eth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/go-common/pkg/generated/contract"
)

func setupContracts(t *testing.T, helper *eth.Helper) (common.Address, *contract.CVLTokenContract) {
	// cast the backend interface to SimulatedBackend so we can use backend.Commit()
	backend := helper.Blockchain.(*backends.SimulatedBackend)

	wei := big.NewInt(1e18)

	// deploy NoOpController
	controllerAddress, _, _, err := contract.DeployNoOpTokenControllerContract(helper.Auth, backend)
	if err != nil {
		t.Fatalf("error deploying NoOpTokenControllerContract %v", err)
	}

	// deploy the CVLToken with 2m tokens in circulation
	supply := big.NewInt(2000000)
	supply.Mul(supply, wei)
	tokenAddress, _, tokenContract, err := contract.DeployCVLTokenContract(helper.Auth, backend, supply, "CVL", 18, "CVL", controllerAddress)
	if err != nil {
		t.Fatalf("error deploying CVLTokenContract %v", err)
	}
	backend.Commit()

	// distribute 500k tokens to "multisig", "user", "foundation"
	// this leaves 500k in the "genesis" account
	tokens := big.NewInt(500000)
	tokens.Mul(tokens, wei)
	tokenContract.Transfer(helper.Auth, helper.Accounts["alice"].Address, tokens)
	tokenContract.Transfer(helper.Auth, helper.Accounts["bob"].Address, tokens)
	tokenContract.Transfer(helper.Auth, helper.Accounts["carol"].Address, tokens)
	backend.Commit()

	return tokenAddress, tokenContract
}

func TestSupply(t *testing.T) {
	// start up a simulated backend
	helper, err := eth.NewSimulatedBackendHelper()

	if err != nil {
		t.Fatalf("error setting up SimulatedBackend %v", err)
	}
	backend := helper.Blockchain.(*backends.SimulatedBackend)

	// deploy CVLToken and distribute CVL to accounts
	cvlTokenAddress, cvlTokenContract := setupContracts(t, helper)

	totalOffering := 1000000.0 // one million tokens for sale
	totalRaiseUSD := 10.0
	startingPrice := 0.2
	pricing := storefront.NewPricingManager(totalOffering, totalRaiseUSD, startingPrice)

	// configure the wallets that will hold the CVL for the token sale
	tokenSaleAddresses := []common.Address{
		helper.Accounts["genesis"].Address,
		helper.Accounts["alice"].Address,
	}

	// initialize a SupplyManager
	supplyManager, err := storefront.NewSupplyManager(cvlTokenAddress, helper.Blockchain, pricing, tokenSaleAddresses, 0)
	if err != nil {
		t.Fatalf("error getting SupplyManager %v", err)
	}

	tokensSold, err := supplyManager.UpdateTokensSold()
	if err != nil {
		t.Fatalf("received an error with UpdateTokensSold %v", err)
	}

	// no tokens have been sold yet (ie, 1m in genesis+multisig)
	expected := 0.0
	if tokensSold != expected {
		t.Fatalf("expecting tokensSold to be %v but it is %v", expected, fmt.Sprintf("%2f", tokensSold))
	}

	// simulate a sale
	cvlTokenContract.Transfer(helper.Accounts["genesis"].Auth, helper.Accounts["bob"].Address, big.NewInt(1*1e18))
	backend.Commit()

	tokensSold, err = supplyManager.UpdateTokensSold()
	if err != nil {
		t.Fatalf("received an error with UpdateTokensSold %v", err)
	}

	expected = 1
	if tokensSold != expected {
		t.Fatalf("expecting tokensSold to be %v but it is %v", expected, tokensSold)
	}

	// simulate a sale
	cvlTokenContract.Transfer(helper.Accounts["genesis"].Auth, helper.Accounts["bob"].Address, big.NewInt(5*1e17))
	backend.Commit()

	tokensSold, err = supplyManager.UpdateTokensSold()
	if err != nil {
		t.Fatalf("received an error with UpdateTokensSold %v", err)
	}

	expected = 1.5
	if tokensSold != expected {
		t.Fatalf("expecting tokensSold to be %v but it is %v", expected, tokensSold)
	}

}
