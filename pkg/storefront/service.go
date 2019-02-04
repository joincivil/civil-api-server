package storefront

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/eth"
)

const (
	// defaultKrakenPollFreqSecs is the how often the price will get updated in seconds
	defaultKrakenPollFreqSecs = 30
	// tokenSupplyPollFreqSecs is how often we inspect token balances of Token Sale wallets to calculate tokens sold
	tokenSupplyPollFreqSecs = 15
	// totalOffering is the number of tokens that will be sold in the token sale
	totalOffering = 34000000.0
	// totalRaiseUSD is the amount of $ we will raise from selling `totalOffering` tokens
	totalRaiseUSD = 19400000.0
	// startingPrice is the initialize price in USD of the first token we sell
	startingPrice = 0.2
)

var (
	// ErrNoTokenSaleAddresses is thrown when `GRAPHQL_TOKEN_SALE_ADDRESSES` envvar is not available
	ErrNoTokenSaleAddresses = errors.New("environment variable `GRAPHQL_TOKEN_SALE_ADDRESSES` not provided")
	// ErrInvalidSupplyManager is thrown when there is an error instatiating a SupplyManager
	ErrInvalidSupplyManager = errors.New("unable to construct a SupplyManager instance")
	// ErrNoCVLTokenAddress is thrown when `GRAPHQL_CONTRACT_ADDRESSES` envvar does not contain `CVLToken`
	ErrNoCVLTokenAddress = errors.New("no CVLToken address provided in configuration")
)

// Service defines methods to operate on the storefront
type Service struct {
	CurrencyConversion CurrencyConversion
	Pricing            *PricingManager
}

// NewService constructs a new Service instance
func NewService(config *utils.GraphQLConfig, ethHelper *eth.Helper) (*Service, error) {

	initSupplyManager := true
	cvlTokenAddress := common.HexToAddress(config.ContractAddresses["CVLToken"])
	if cvlTokenAddress == common.HexToAddress("") {
		initSupplyManager = false
		glog.Infof("Not initializing supply manager, err: %v", ErrNoCVLTokenAddress)

	} else if config.TokenSaleAddresses == nil || len(config.TokenSaleAddresses) < 1 {
		initSupplyManager = false
		glog.Infof("Not initializing supply manager, err: %v", ErrNoTokenSaleAddresses)
	}

	pricingManager := NewPricingManager(totalOffering, totalRaiseUSD, startingPrice)
	currencyConversion := NewKrakenCurrencyConversion(defaultKrakenPollFreqSecs)

	if initSupplyManager {
		_, err := NewSupplyManager(
			cvlTokenAddress,
			ethHelper.Blockchain,
			pricingManager,
			config.TokenSaleAddresses,
			tokenSupplyPollFreqSecs,
		)
		if err != nil {
			return nil, ErrInvalidSupplyManager
		}

	}

	return &Service{
		Pricing:            pricingManager,
		CurrencyConversion: currencyConversion,
	}, nil
}
