package airswap

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/eth"
)

var (
	// ErrNoTokenSaleAddresses is thrown when `GRAPHQL_TOKEN_SALE_ADDRESSES` envvar is not available
	ErrNoTokenSaleAddresses = errors.New("environment variable `GRAPHQL_TOKEN_SALE_ADDRESSES` not provided")
	// ErrInvalidSupplyManager is thrown when there is an error instatiating a SupplyManager
	ErrInvalidSupplyManager = errors.New("unable to construct a SupplyManager instance")
	// ErrNoCVLTokenAddress is thrown when `GRAPHQL_CONTRACT_ADDRESSES` envvar does not contain `CVLToken`
	ErrNoCVLTokenAddress = errors.New("no CVLToken address provided in configuration")
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

// EnableAirswapRouting adds Airswap routes
func EnableAirswapRouting(router chi.Router, config *utils.GraphQLConfig, ethHelper *eth.Helper) error {

	if config.ContractAddresses == nil || config.ContractAddresses["CVLToken"] == "" {
		return ErrNoCVLTokenAddress
	}
	cvlTokenAddress := common.HexToAddress(config.ContractAddresses["CVLToken"])
	if cvlTokenAddress == common.HexToAddress("") {
		return ErrNoCVLTokenAddress
	}
	if config.TokenSaleAddresses == nil || len(config.TokenSaleAddresses) < 1 {
		return ErrNoTokenSaleAddresses
	}

	pricingManager := NewPricingManager(totalOffering, totalRaiseUSD, startingPrice)
	pairPricing := NewKrakenPairPricing(defaultKrakenPollFreqSecs)

	_, err := NewSupplyManager(
		cvlTokenAddress,
		ethHelper.Blockchain,
		pricingManager,
		config.TokenSaleAddresses,
		tokenSupplyPollFreqSecs,
	)
	if err != nil {
		return ErrInvalidSupplyManager
	}

	handlers := &Handlers{Pricing: pricingManager, Conversion: pairPricing}

	router.Group(func(r chi.Router) {
		r.Use(InternalOnlyMiddleware())
		r.Post("/airswap", handlers.GetOrder)
	})

	return nil
}

// InternalOnlyMiddleware returns 403 Forbidden if it receives traffic from the Load Balancer
func InternalOnlyMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			header := r.Header.Get("X-Forwarded-For")

			if header != "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				respBody := fmt.Sprintf("forbidden")
				_, _ = w.Write([]byte(respBody)) // nolint: gosec
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
