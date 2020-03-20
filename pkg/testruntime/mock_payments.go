package testruntime

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/channels"
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
func (p *MockPaymentHelper) CreateCharge(request payments.CreateChargeRequest) (payments.CreateChargeResponse, error) {
	return payments.CreateChargeResponse{}, nil
}

// CreateCustomer is a mock to create a stripe customer
func (p *MockPaymentHelper) CreateCustomer(request payments.CreateCustomerRequest) (payments.CreateCustomerResponse, error) {
	return payments.CreateCustomerResponse{}, nil
}

// AddCustomerCard is a mock to add a card to a stripe customer
func (p *MockPaymentHelper) AddCustomerCard(request payments.AddCustomerCardRequest) (payments.AddCustomerCardResponse, error) {
	return payments.AddCustomerCardResponse{}, nil
}

// SetStripeCustomerID is a mock to set the stripe customer ID of a channel
func (p *MockPaymentHelper) SetStripeCustomerID(channelID string, stripeCustomerID string) (*channels.Channel, error) {
	return nil, nil
}

// CreateStripePaymentIntent is a mock to create a payment intent
func (p *MockPaymentHelper) CreateStripePaymentIntent(request payments.CreatePaymentIntentRequest) (payments.StripePaymentIntent, error) {
	return payments.StripePaymentIntent{}, nil
}

// ClonePaymentMethod is a mock to clone payment methods
func (p *MockPaymentHelper) ClonePaymentMethod(request payments.ClonePaymentMethodRequest) (payments.ClonePaymentMethodResponse, error) {
	return payments.ClonePaymentMethodResponse{}, nil
}

// RemovePaymentMethod is a mock to remove a payment method
func (p *MockPaymentHelper) RemovePaymentMethod(paymentMethodID string) error {
	return nil
}

// GetEthereumPaymentAddress returns a mock eth account for the channel address
func (p *MockPaymentHelper) GetEthereumPaymentAddress(channelID string) (common.Address, error) {

	return common.HexToAddress("101"), nil
}

// GetStripePaymentAccount returns a mock stripe payment account for the channel
func (p *MockPaymentHelper) GetStripePaymentAccount(channelID string) (string, error) {
	return "stripe" + channelID, nil
}

// GetStripeCustomerID returns a mock stripe customer id for the channel
func (p *MockPaymentHelper) GetStripeCustomerID(channelID string) (string, error) {
	return "stripe" + channelID, nil
}

// GetCustomerInfo returns a mock stripe customer info for the channel
func (p *MockPaymentHelper) GetCustomerInfo(channelID string) (payments.StripeCustomerInfo, error) {
	return payments.StripeCustomerInfo{}, nil
}

// GetStripeApplyPayDomains returns a mock list of enabled apply pay domains
func (p *MockPaymentHelper) GetStripeApplyPayDomains(channelID string) ([]string, error) {
	return []string{}, nil
}

// GetChannelAdminUserChannels returns a mock list of a channel's admins' user channels
func (p *MockPaymentHelper) GetChannelAdminUserChannels(channelID string) ([]*channels.Channel, error) {
	return []*channels.Channel{}, nil
}
