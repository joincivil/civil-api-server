package storefront

import (
	"errors"
	"math/big"
	"time"

	log "github.com/golang/glog"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/go-common/pkg/generated/contract"
)

var (
	// ErrInvalidTokensSold thrown when there is an invalid number of tokens sold
	ErrInvalidTokensSold = errors.New("sanity check failed: tokens sold < 0")
)

// SupplyManager updates the pricing.TokensSold based on available supply
type SupplyManager struct {
	pricing            *PricingManager
	cvlTokenContract   *contract.CVLTokenContract
	tokenSaleAddresses []common.Address // Ethereum addresses that are selling tokens
	UpdateTicker       *time.Ticker     // used to control periodic calling of UpdateTokensSold
}

// NewSupplyManager returns a new SupplyManager
func NewSupplyManager(
	cvlTokenAddress common.Address,
	blockchain bind.ContractBackend,
	pricing *PricingManager,
	tokenSaleAddresses []common.Address,
	pollingFrequencySeconds uint,
) (*SupplyManager, error) {

	log.Infof("Initializing CVLToken at address %v", cvlTokenAddress.String())
	cvlToken, err := contract.NewCVLTokenContract(cvlTokenAddress, blockchain)
	if err != nil {
		return nil, err
	}

	manager := &SupplyManager{
		cvlTokenContract:   cvlToken,
		pricing:            pricing,
		tokenSaleAddresses: tokenSaleAddresses,
	}

	_, err = manager.UpdateTokensSold()
	if err != nil {
		return nil, err
	}

	if pollingFrequencySeconds > 0 {
		manager.SupplyPolling(pollingFrequencySeconds)
	}

	return manager, nil
}

// SupplyPolling calls UpdateTokensSold at the specified interval
func (s *SupplyManager) SupplyPolling(frequencySeconds uint) {
	s.UpdateTicker = time.NewTicker(time.Duration(frequencySeconds) * time.Second)
	go func() {
		for range s.UpdateTicker.C {
			_, err := s.UpdateTokensSold()
			if err != nil {
				log.Errorf("Error with SupplyPolling %v", err)
			}
		}
	}()
}

// UpdateTokensSold sets the supply based on whats left in the hot wallet and multisig
func (s *SupplyManager) UpdateTokensSold() (float64, error) {

	// sum up all of the address balances that hold CVL to be sold
	availableTokens := big.NewInt(0)
	for _, addr := range s.tokenSaleAddresses {
		balance, err := s.cvlTokenContract.BalanceOf(&bind.CallOpts{}, addr)
		if err != nil {
			log.Errorf("Error receiving balance for %v: err: %v", addr.String(), err)
			return 0, err
		}
		availableTokens.Add(availableTokens, balance)
		if log.V(3) {
			log.Infof("%v has %v tokens\n", addr.String(), balance.Quo(balance, big.NewInt(1e18)).String())
		}

	}

	// store the total number of tokens to be sold in a big.Int
	offering := big.NewInt(int64(s.pricing.TotalOffering))
	// and convert it into wei
	offering = offering.Mul(offering, big.NewInt(1e18))

	// tokensSold is (TotalOffering - availableTokens) and is in wei
	tokensSold := new(big.Float).SetInt(offering.Sub(offering, availableTokens))

	// and then convert from wei
	tokensSold.Quo(tokensSold, big.NewFloat(1e18))
	// unwrap from big.Float to float
	tokensSoldFloat, _ := tokensSold.Float64()

	// this should never happen unless you did something screwy with your parameters
	if tokensSoldFloat < 0.0 {
		log.Errorf("sanity check failed - tokens sold < 0")
		return 0, ErrInvalidTokensSold
	}

	if s.pricing.TokensSold != tokensSoldFloat {
		log.Infof("Updating tokens sold to %v\n", tokensSold.String())
		s.pricing.TokensSold = tokensSoldFloat
	}

	return tokensSoldFloat, nil
}
