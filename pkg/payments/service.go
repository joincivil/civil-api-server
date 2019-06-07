package payments

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
)

// StripeCharger defines the functions needed to create a charge with Stripe
type StripeCharger interface {
	CreateCharge(request *CreateChargeRequest) (CreateChargeResponse, error)
}

// EthereumValidator defines the functions needed to create an Ethereum payment
type EthereumValidator interface {
	ValidateTransaction(transactionID string, expectedAccount string) (*ValidateTransactionResponse, error)
}

// Service provides methods to interact with Posts
type Service struct {
	db       *gorm.DB
	stripe   StripeCharger
	ethereum EthereumValidator
}

// NewService builds an instance of posts.Service
func NewService(db *gorm.DB, stripe StripeCharger, ethereum EthereumValidator) *Service {
	return &Service{
		db,
		stripe,
		ethereum,
	}
}

// CreateEtherPayment confirm that an Ether transaction is valid and store the result as a Payment in the database
func (s *Service) CreateEtherPayment(channelID string, ownerType string, ownerID string, payment EtherPayment) (EtherPayment, error) {
	expectedAccount := getEthereumAccountForChannel(channelID)
	res, err := s.ethereum.ValidateTransaction(payment.TransactionID, expectedAccount)

	if err != nil {
		return EtherPayment{}, err
	}
	id, err := uuid.NewV4()
	if err != nil {
		return EtherPayment{}, err
	}
	data, err := json.Marshal(res)
	if err != nil {
		return EtherPayment{}, err
	}
	payment.ID = id.String()
	payment.PaymentType = payment.Type()
	payment.Data = postgres.Jsonb{RawMessage: data}
	payment.OwnerID = ownerID
	payment.OwnerType = ownerType
	payment.CurrencyCode = "ETH"
	payment.ExchangeRate = res.ExchangeRate
	payment.Amount = res.Amount
	// payment.Amount = res.Value

	if err = s.db.Create(&payment).Error; err != nil {
		fmt.Printf("An error occured: %v\n", err)
		return EtherPayment{}, err
	}
	return payment, nil
}

// CreateStripePayment will create a Stripe charge and then store the result as a Payment in the database
func (s *Service) CreateStripePayment(channelID string, ownerType string, ownerID string, payment StripePayment) (StripePayment, error) {
	res, err := s.stripe.CreateCharge(&CreateChargeRequest{
		Amount:        int64(math.Floor(payment.Amount * 100)),
		SourceToken:   payment.PaymentToken,
		StripeAccount: getStripeAccountForChannel(channelID),
		Metadata:      map[string]string{ownerType: ownerID},
	})

	if err != nil {
		return StripePayment{}, err
	}
	id, err := uuid.NewV4()
	if err != nil {
		return StripePayment{}, err
	}
	payment.ID = id.String()
	payment.PaymentType = payment.Type()
	payment.Data = postgres.Jsonb{RawMessage: json.RawMessage(res.StripeResponseJSON)}
	payment.OwnerID = ownerID
	payment.OwnerType = ownerType
	payment.ExchangeRate = 1

	if err = s.db.Create(&payment).Error; err != nil {
		fmt.Printf("An error occured: %v\n", err)
		return StripePayment{}, err
	}
	return payment, nil
}

func (s *Service) TotalPaymentsUSD(postID string) (float64, error) {
	return 100, nil
}

func getStripeAccountForChannel(channelID string) string {
	// TODO(dankins): this needs to be implemented, this is just a test account
	return "acct_1C4vupLMQdVwYica"
}

func getEthereumAccountForChannel(channelID string) string {
	// TODO(dankins): this needs to be implemented, this is just a test account
	return "0xddB9e9957452d0E39A5E43Fd1AB4aE818aecC6aD"
}
