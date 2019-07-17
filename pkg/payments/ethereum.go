package payments

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/storefront"
)

var (
	// ErrorTransactionFailed is returned when a transaction status is not 1
	ErrorTransactionFailed = fmt.Errorf("transaction failed")
	// ErrorTransactionNotFound is returned when the transaction is not found from the TransactionReader
	ErrorTransactionNotFound = fmt.Errorf("transaction not found")
	// ErrorReceiptNotFound is returned when the receipt is not found from the TransactionReader
	ErrorReceiptNotFound = fmt.Errorf("receipt not found")
	// ErrorInvalidRecipient is returned when a transaction was not sent to the expected address
	ErrorInvalidRecipient = fmt.Errorf("invalid recipient")
)

// EthereumPaymentService validates Layer1 payments
type EthereumPaymentService struct {
	chain              ethereum.TransactionReader
	currencyConversion storefront.CurrencyConversion
}

// ValidateTransactionResponse is the response of ValidateTransaction
type ValidateTransactionResponse struct {
	Amount         float64
	TransactionID  string
	ExchangeRate   float64
	PaymentAddress string
}

// NewEthereumPaymentService creates an EthereumPaymentService instance
func NewEthereumPaymentService(transactionReader ethereum.TransactionReader, currencyConversion storefront.CurrencyConversion) *EthereumPaymentService {

	return &EthereumPaymentService{
		chain:              transactionReader,
		currencyConversion: currencyConversion,
	}
}

// ValidateTransaction accepts a transaction and determines whether it is valid
func (s *EthereumPaymentService) ValidateTransaction(transactionID string, expectedReceiver common.Address) (*ValidateTransactionResponse, error) {

	// retrieve the transaction information
	data, _, err := s.chain.TransactionByHash(context.Background(), common.HexToHash(transactionID))
	if err == ethereum.NotFound {
		return nil, ErrorTransactionNotFound
	} else if err != nil {
		return nil, fmt.Errorf("error with transaction: %v", err)
	}

	// retrieve the transaction receipt
	receipt, err := s.chain.TransactionReceipt(context.Background(), common.HexToHash(transactionID))
	if err == ethereum.NotFound {
		return nil, ErrorReceiptNotFound
	} else if err != nil {
		return nil, fmt.Errorf("error with transaction: %v", err)
	}

	// confirm the tx was successful
	if receipt.Status != 1 {
		return nil, ErrorTransactionFailed
	}

	// ensure that we are actually transferring to the correct address
	// this is to prevent trying to game the system but sending a tx from a payment that wasn't actually to them
	if data.To().String() != expectedReceiver.String() {
		return nil, ErrorInvalidRecipient
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
		PaymentAddress: data.To().String(),
		Amount:         valueFloat,
		TransactionID:  data.Hash().String(),
		ExchangeRate:   exchangeRate,
	}, nil
}
