package testruntime

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/payments"
)

// MockPaymentHelper implements payments.StripeCharger, and ChannelHelper interface
type MockPaymentHelper struct {
	txs *MockTransactionReader
}

// NewMockPaymentHelper creates a new NewMockPaymentHelper
func NewMockPaymentHelper(txs *MockTransactionReader) *MockPaymentHelper {
	return &MockPaymentHelper{txs}
}

// CreateCharge is a mock to create a stripe charge
func (p *MockPaymentHelper) CreateCharge(request *payments.CreateChargeRequest) (payments.CreateChargeResponse, error) {
	return payments.CreateChargeResponse{}, nil
}

// GetEthereumPaymentAddress returns a mock eth account for the channel address
func (p *MockPaymentHelper) GetEthereumPaymentAddress(channelID string) (common.Address, error) {

	return common.HexToAddress("101"), nil
}

// GetStripePaymentAccount returns a mock stripe payment account for the channel
func (p *MockPaymentHelper) GetStripePaymentAccount(channelID string) (string, error) {
	return "stripe" + channelID, nil
}
