package payments

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/storefront"
)

const (
	// defaultKrakenPollFreqSecs is the how often the price will get updated in seconds
	defaultKrakenPollFreqSecs = 30
)

// EthereumService validates Layer1 payments
type EthereumService struct {
	chain              ethereum.TransactionReader
	currencyConversion storefront.CurrencyConversion
}

// ValidateTransactionResponse is the response of ValidateTransaction
type ValidateTransactionResponse struct {
	Amount        float64
	TransactionID string
	ExchangeRate  float64
}

// NewEthereumService creates an EthereumService instance
func NewEthereumService(chain ethereum.TransactionReader) *EthereumService {
	// TODO(dankins): this should be injected, but holding off until potential uber.fx changes
	currencyConversion := storefront.NewKrakenCurrencyConversion(defaultKrakenPollFreqSecs)
	return &EthereumService{
		chain,
		currencyConversion,
	}
}

// ValidateTransaction accepts a transaction and determines whether it is valid
func (s *EthereumService) ValidateTransaction(transactionID string, expectedReceiver string) (*ValidateTransactionResponse, error) {

	// parse expectedReceiver into an ETH address
	receiverAddr := common.HexToAddress(expectedReceiver)
	if receiverAddr == common.HexToAddress("0x") {
		return nil, errors.New("invalid expectedReceiver address")
	}

	// retrieve the transaction information
	data, _, err := s.chain.TransactionByHash(context.Background(), common.HexToHash(transactionID))
	if err != nil {
		return nil, fmt.Errorf("error with transaction: %v", err)
	}

	// ensure that we are actually transferring to the correct address
	// this is to prevent trying to game the system but sending a tx from a payment that wasn't actually to them
	if data.To().String() != receiverAddr.String() {

		// TODO(dankins): don't actually return an error until this is fully implemented
		log.Errorf("transaction sent to %v but was expecting %v. continuing on until the expectedReceiver logic is implemented.", data.To().String(), receiverAddr.String())
		// return nil, fmt.Errorf("transaction sent to %v but was expecting %v", data.To().String(), receiverAddr.String())
	}
	// convert the transction amount from wei to ether
	var ether = new(big.Float).SetInt(data.Value())
	ether = ether.Quo(ether, big.NewFloat(1e18))
	valueFloat, _ := ether.Float64()

	// retrieve the current exchange rate to USD
	exchangeRate, err := s.currencyConversion.ETHToUSD()
	if err != nil {
		return nil, fmt.Errorf("error getting exchange rate: err: %v", err)
	}

	return &ValidateTransactionResponse{
		Amount:        valueFloat,
		TransactionID: data.Hash().String(),
		ExchangeRate:  exchangeRate,
	}, nil
}
