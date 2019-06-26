package payments

import (
	"encoding/json"
	"errors"
	"math"

	log "github.com/golang/glog"
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
	// get the wallet address for this Channel
	expectedAccount := getEthereumAccountForChannel(channelID)

	// ensure the transaction is valid and goes to the correct wallet
	res, err := s.ethereum.ValidateTransaction(payment.TransactionID, expectedAccount)
	if err != nil {
		return EtherPayment{}, err
	}

	// generate a new ID
	id, err := uuid.NewV4()
	if err != nil {
		return EtherPayment{}, err
	}
	payment.ID = id.String()

	// set the `data` column to the result of ValidateTransaction
	data, err := json.Marshal(res)
	if err != nil {
		return EtherPayment{}, err
	}
	payment.Data = postgres.Jsonb{RawMessage: data}

	payment.PaymentType = payment.Type()
	payment.OwnerID = ownerID
	payment.OwnerType = ownerType
	payment.CurrencyCode = "ETH"
	payment.ExchangeRate = res.ExchangeRate
	payment.Amount = res.Amount

	if err = s.db.Create(&payment).Error; err != nil {
		log.Errorf("An error occured: %v\n", err)
		return EtherPayment{}, err
	}
	return payment, nil
}

// CreateStripePayment will create a Stripe charge and then store the result as a Payment in the database
func (s *Service) CreateStripePayment(channelID string, ownerType string, ownerID string, payment StripePayment) (StripePayment, error) {

	// generate a stripe charge
	res, err := s.stripe.CreateCharge(&CreateChargeRequest{
		Amount:        int64(math.Floor(payment.Amount * 100)),
		SourceToken:   payment.PaymentToken,
		StripeAccount: getStripeAccountForChannel(channelID),
		Metadata:      map[string]string{ownerType: ownerID},
	})
	if err != nil {
		return StripePayment{}, err
	}

	// generate a new ID for the payment model
	id, err := uuid.NewV4()
	if err != nil {
		return StripePayment{}, err
	}
	payment.ID = id.String()

	payment.PaymentType = payment.Type()

	// set the `data` column to the stripe response
	payment.Data = postgres.Jsonb{RawMessage: json.RawMessage(res.StripeResponseJSON)}
	payment.OwnerID = ownerID
	payment.OwnerType = ownerType

	// TODO(dankins): this should be set when we support currencies other than USD
	payment.ExchangeRate = 1

	if err = s.db.Create(&payment).Error; err != nil {
		log.Errorf("An error occured: %v\n", err)
		return StripePayment{}, err
	}
	return payment, nil
}

// GetPayments returns the payments associated with a Post
func (s *Service) GetPayments(postID string) ([]Payment, error) {
	var pays []PaymentModel
	if err := s.db.Where(&PaymentModel{OwnerType: "posts", OwnerID: postID}).Find(&pays).Error; err != nil {
		log.Errorf("An error occured: %v\n", err)
		return nil, err
	}

	var paymentsSlice []Payment
	for _, result := range pays {
		payment, err := ModelToInterface(&result)
		if err != nil {
			log.Errorf("An error occured: %v\n", err)
			return nil, err
		}
		paymentsSlice = append(paymentsSlice, payment)
	}

	return paymentsSlice, nil
}

// TotalPayments returns the USD equivalent of all payments associated with the post
func (s *Service) TotalPayments(postID string, currencyCode string) (float64, error) {
	if currencyCode != "USD" {
		return 0, errors.New("USD is the only `currencyCode` supported")
	}
	var totals []float64
	s.db.Table("payments").Where(&PaymentModel{OwnerType: "posts", OwnerID: postID}).Select("sum(amount * exchange_rate) as total").Pluck("total", &totals)

	return totals[0], nil
}

func getStripeAccountForChannel(channelID string) string {
	// TODO(dankins): this needs to be implemented, this is just a test account
	return "acct_1C4vupLMQdVwYica"
}

func getEthereumAccountForChannel(channelID string) string {
	// TODO(dankins): this needs to be implemented, this is just a test account
	return "0xddB9e9957452d0E39A5E43Fd1AB4aE818aecC6aD"
}
