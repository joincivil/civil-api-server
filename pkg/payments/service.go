package payments

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/common"

	log "github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joincivil/go-common/pkg/email"
	uuid "github.com/satori/go.uuid"
)

const (
	ethPaymentStartedEmailTemplateID  = "d-a4595d0eb8c941ab897b9414ac846aff"
	ethPaymentFinishedEmailTemplateID = "d-a4595d0eb8c941ab897b9414ac846aff"
	ccPaymentReceiptEmailTemplateID   = "d-a4595d0eb8c941ab897b9414ac846aff"

	civilEmailName      = "Civil"
	supportEmailAddress = "support@civil.co"

	defaultFromEmailName    = civilEmailName
	defaultFromEmailAddress = supportEmailAddress

	// TODO: get correct group ID for payments
	defaultAsmGroupID = 8328 // Civil Registry Alerts
)

// StripeCharger defines the functions needed to create a charge with Stripe
type StripeCharger interface {
	CreateCharge(request *CreateChargeRequest) (CreateChargeResponse, error)
}

// EthereumValidator defines the functions needed to create an Ethereum payment
type EthereumValidator interface {
	ValidateTransaction(transactionID string, expectedAccount common.Address) (*ValidateTransactionResponse, error)
}

// ChannelHelper defines the methods needed to interact with a channel
type ChannelHelper interface {
	GetEthereumPaymentAddress(channelID string) (common.Address, error)
	GetStripePaymentAccount(channelID string) (string, error)
}

// Service provides methods to interact with Posts
type Service struct {
	db       *gorm.DB
	stripe   StripeCharger
	ethereum EthereumValidator
	channel  ChannelHelper
	emailer  *email.Emailer
}

// NewService builds an instance of posts.Service
func NewService(db *gorm.DB, stripe StripeCharger, ethereum EthereumValidator, channel ChannelHelper, emailer *email.Emailer) *Service {
	return &Service{
		db,
		stripe,
		ethereum,
		channel,
		emailer,
	}
}

func (s *Service) getTemplateRequest(templateID string, emailAddress string) (req *email.SendTemplateEmailRequest) {
	tmplData := email.TemplateData{
		"newsroom_name": "test1",
	}
	return &email.SendTemplateEmailRequest{
		ToName:       emailAddress,
		ToEmail:      emailAddress,
		FromName:     defaultFromEmailName,
		FromEmail:    defaultFromEmailAddress,
		TemplateID:   templateID,
		TemplateData: tmplData,
		AsmGroupID:   defaultAsmGroupID,
	}
}

func (s *Service) getTemplateRequest2(templateID string, emailAddress string, tmplData email.TemplateData) (req *email.SendTemplateEmailRequest) {
	return &email.SendTemplateEmailRequest{
		ToName:       emailAddress,
		ToEmail:      emailAddress,
		FromName:     defaultFromEmailName,
		FromEmail:    defaultFromEmailAddress,
		TemplateID:   templateID,
		TemplateData: tmplData,
		AsmGroupID:   defaultAsmGroupID,
	}
}

func (s *Service) sendEthPaymentStartedEmail(emailAddress string, tmplData email.TemplateData) {
	req := s.getTemplateRequest2(ethPaymentStartedEmailTemplateID, emailAddress, tmplData)
	s.emailer.SendTemplateEmail(req)
}

func (s *Service) sendEthPaymentFinishedEmail(emailAddress string) {
	req := s.getTemplateRequest(ethPaymentFinishedEmailTemplateID, emailAddress)
	s.emailer.SendTemplateEmail(req)
}

func (s *Service) sendCCPaymentReceiptEmail(emailAddress string) {
	req := s.getTemplateRequest(ccPaymentReceiptEmailTemplateID, emailAddress)
	s.emailer.SendTemplateEmail(req)
}

// CreateEtherPayment confirm that an Ether transaction is valid and store the result as a Payment in the database
func (s *Service) CreateEtherPayment(channelID string, ownerType string, ownerID string, txID string, emailAddress string, tmplData email.TemplateData) (EtherPayment, error) {
	hash := common.HexToHash(txID)
	if (hash == common.Hash{}) {
		return EtherPayment{}, errors.New("invalid tx id")
	}

	payment := PaymentModel{}
	expectedAddress, err := s.channel.GetEthereumPaymentAddress(channelID)
	if err != nil {
		return EtherPayment{}, err
	}
	// generate a new ID
	id, err := uuid.NewV4()
	if err != nil {
		return EtherPayment{}, err
	}
	payment.ID = id.String()
	payment.PaymentType = "ether"
	payment.Reference = txID

	payment.Status = "pending"
	payment.OwnerID = ownerID
	payment.OwnerType = ownerType
	payment.CurrencyCode = "ETH"
	payment.ExchangeRate = 0
	payment.Amount = 0
	payment.EmailAddress = emailAddress

	payment.Data = postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf("{\"PaymentAddress\":\"%v\"}", expectedAddress.String()))}

	if err = s.db.Create(&payment).Error; err != nil {
		log.Errorf("An error occured: %v\n", err)
		return EtherPayment{}, err
	}

	// if no email address given, that's fine
	if emailAddress != "" {
		s.sendEthPaymentStartedEmail(emailAddress, tmplData)
	}

	return EtherPayment{
		PaymentModel: payment,
	}, nil
}

// GetPendingEtherPayments gets all pending ether payments
func (s *Service) GetPendingEtherPayments() ([]PaymentModel, error) {

	var payments []PaymentModel
	if err := s.db.Where("status = 'pending' AND payment_type = 'ether'").Find(&payments).Error; err != nil {
		return nil, err
	}

	return payments, nil
}

// UpdateEtherPayments finds pending payments, checks the status, and updates them accordingly
func (s *Service) UpdateEtherPayments() error {

	payments, err := s.GetPendingEtherPayments()
	if err != nil {
		return err
	}

	for _, payment := range payments {
		err := s.UpdateEtherPayment(&payment)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateEtherPayment handles a single payment object and updates it if needed
func (s *Service) UpdateEtherPayment(payment *PaymentModel) error {
	// create a payment model to hold the updated fields
	update := &PaymentModel{}

	// convert to interface to unmarshal up the Data field
	paymentInterface, err := ModelToInterface(payment)
	if err != nil {
		return err
	}
	// and assert back to an EtherPayment
	etherPayment := paymentInterface.(*EtherPayment)

	// expectedReceiver should be the channel ETH address
	expectedReceiver := common.HexToAddress(etherPayment.PaymentAddress)
	res, err := s.ethereum.ValidateTransaction(payment.Reference, expectedReceiver)
	if err == ErrorTransactionFailed {
		update.Status = "failed"
	} else if err == ErrorReceiptNotFound || err == ErrorTransactionNotFound {
		return nil
	} else if err == ErrorInvalidRecipient {
		update.Status = "invalid"
	} else if err != nil {
		log.Errorf("Error updating payment: %v\n", err)
		// payment.Status = "error"
		return err
	} else {
		data, err := json.Marshal(res)
		if err != nil {
			log.Errorf("Error updating payment: %v\n", err)
			return err
		}

		if res.Amount != 0 {
			update.Status = "complete"
			update.Data = postgres.Jsonb{RawMessage: data}
			update.ExchangeRate = res.ExchangeRate
			update.Amount = res.Amount

			// if no email address given, that's fine
			if etherPayment.EmailAddress != "" {
				s.sendEthPaymentFinishedEmail(etherPayment.EmailAddress)
			}
		}
	}

	// set the `data` column to the result of ValidateTransaction

	if err = s.db.Model(&payment).Update(update).Error; err != nil {
		log.Errorf("Error updating payment: %v\n", err)
		return err
	}

	return nil
}

// CreateStripePayment will create a Stripe charge and then store the result as a Payment in the database
func (s *Service) CreateStripePayment(channelID string, ownerType string, ownerID string, payment StripePayment) (StripePayment, error) {

	stripeAccount, err := s.channel.GetStripePaymentAccount(channelID)
	if err != nil {
		return StripePayment{}, err
	}

	// generate a stripe charge
	res, err := s.stripe.CreateCharge(&CreateChargeRequest{
		Amount:        int64(math.Floor(payment.Amount * 100)),
		SourceToken:   payment.PaymentToken,
		StripeAccount: stripeAccount,
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
	// if no email address given, that's fine
	if payment.EmailAddress != "" {
		s.sendCCPaymentReceiptEmail(payment.EmailAddress)
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

// GetPayment returns the payment with the given ID
func (s *Service) GetPayment(paymentID string) (Payment, error) {
	var paymentModel PaymentModel
	if err := s.db.Where(&PaymentModel{ID: paymentID}).First(&paymentModel).Error; err != nil {
		log.Errorf("An error occured: %v\n", err)
		return nil, err
	}

	payment, err := ModelToInterface(&paymentModel)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// TotalPayments returns the USD equivalent of all payments associated with the post
func (s *Service) TotalPayments(postID string, currencyCode string) (float64, error) {
	if currencyCode != "USD" {
		return 0, errors.New("USD is the only `currencyCode` supported")
	}
	var totals []float64
	s.db.Table("payments").Where(&PaymentModel{OwnerType: "posts", OwnerID: postID}).Select("coalesce(sum(amount * exchange_rate), 0) as total").Pluck("total", &totals)

	return totals[0], nil
}
