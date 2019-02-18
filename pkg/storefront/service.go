package storefront

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/users"
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
	currencyConversion CurrencyConversion
	pricing            *PricingManager
	userService        *users.UserService
}

// NewService constructs a new Service instance
func NewService(cvlTokenAddr string, tokenSaleAddrs []common.Address, ethHelper *eth.Helper,
	userService *users.UserService) (*Service, error) {

	initSupplyManager := true
	cvlTokenAddress := common.HexToAddress(cvlTokenAddr)
	if cvlTokenAddress == common.HexToAddress("") {
		initSupplyManager = false
		glog.Infof("Not initializing supply manager, err: %v", ErrNoCVLTokenAddress)

	} else if tokenSaleAddrs == nil || len(tokenSaleAddrs) < 1 {
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
			tokenSaleAddrs,
			tokenSupplyPollFreqSecs,
		)
		if err != nil {
			return nil, ErrInvalidSupplyManager
		}
	}

	return &Service{
		pricing:            pricingManager,
		currencyConversion: currencyConversion,
		userService:        userService,
	}, nil
}

// BuildService makes a Service with the specified parameters
func BuildService(pricingManager *PricingManager, currencyConversion CurrencyConversion) *Service {
	return &Service{
		pricing:            pricingManager,
		currencyConversion: currencyConversion,
	}
}

// GetQuote returns the price in USD to buy `numTokens` of CVL
func (s *Service) GetQuote(numTokens float64) float64 {
	return s.pricing.GetQuote(numTokens)
}

// GetTokensToBuy returns the number of tokens you will receive if you buy `usdToSpend` worth of CVL
func (s *Service) GetTokensToBuy(usdToSpend float64) float64 {
	return s.pricing.GetTokensToBuy(usdToSpend)
}

// ConvertUSDToETH returns the price of 1 USD in ETH
func (s *Service) ConvertUSDToETH() (float64, error) {
	return s.currencyConversion.USDToETH()
}

// ConvertETHToUSD returns the price of 1 ETH in USD
func (s *Service) ConvertETHToUSD() (float64, error) {
	return s.currencyConversion.ETHToUSD()
}

// AirswapOnComplete handles the transaction hash from the Airswap onComplete when a sale
// is completed
func (s *Service) AirswapOnComplete(buyerUID string, txHash string) error {
	user, err := s.userService.MaybeGetUser(users.UserCriteria{
		UID: buyerUID,
	})
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("No user found: uid: %v", buyerUID)
	}

	err = user.AddPurchaseTxHash(txHash)
	if err != nil {
		return err
	}

	update := &users.UserUpdateInput{
		PurchaseTxHashesStr: user.PurchaseTxHashesStr,
	}
	_, err = s.userService.UpdateUser(buyerUID, update)
	if err != nil {
		return err
	}

	// Add user to the Mailchimp members list for marketing/growth team

	return nil
}

// AirswapOnCancel handles the Airswap onCancel when a sale is cancelled
func (s *Service) AirswapOnCancel(buyerUID string) error {
	// Add user to the Mailchimp abandoned list for marketing/growth team
	return nil
}
