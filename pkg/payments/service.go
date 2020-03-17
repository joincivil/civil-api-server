package payments

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/channels"

	log "github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joincivil/go-common/pkg/email"
	uuid "github.com/satori/go.uuid"
	"github.com/stripe/stripe-go"
	"time"
)

const (
	boostEthPaymentStartedEmailTemplateID              = "d-a4595d0eb8c941ab897b9414ac846aff"
	boostEthPaymentFinishedEmailTemplateID             = "d-1e8763f27ef843cd8850ca297a426f3d"
	boostStripePaymentReceiptEmailTemplateID           = "d-b5d79746c540439fac8791b192135aa6"
	externalLinkEthPaymentStartedEmailTemplateID       = "d-a6b6a640436b40cf921811a80f4709b9"
	externalLinkEthPaymentFinishedEmailTemplateID      = "d-7b7a6f1d0d144bcd995dd92a27429976"
	externalLinkStripePaymentReceiptEmailTemplateID    = "d-e6cc3a3a827e418b81f4af65bc802dcc"
	boostPaymentReceivedEmailTemplateID                = "d-897d6f1bdb6d4f44a0507bed3f5b3cf5"
	boostPaymentReceivedNoEmailProvidedEmailTemplateID = "d-e1d786920ebf45b89eda1f74b9d73687"

	civilEmailName      = "Civil"
	supportEmailAddress = "support@civil.co"

	defaultFromEmailName    = civilEmailName
	defaultFromEmailAddress = supportEmailAddress

	// TODO: get correct group ID for payments
	defaultAsmGroupID = 8328 // Civil Registry Alerts

	postTypeBoost        = "boost"
	postTypeExternalLink = "externallink"

	paymentComplete = "complete"
)

// StripeCharger defines the functions needed to create a charge with Stripe
type StripeCharger interface {
	CreateCustomer(request CreateCustomerRequest) (CreateCustomerResponse, error)
	AddCustomerCard(request AddCustomerCardRequest) (AddCustomerCardResponse, error)
	CreateCharge(request CreateChargeRequest) (CreateChargeResponse, error)
	GetCustomerInfo(customerID string) (StripeCustomerInfo, error)
	CreateStripePaymentIntent(request CreatePaymentIntentRequest) (StripePaymentIntent, error)
	ClonePaymentMethod(request ClonePaymentMethodRequest) (ClonePaymentMethodResponse, error)
	RemovePaymentMethod(paymentMethodID string) error
}

// EthereumValidator defines the functions needed to create an Ethereum payment
type EthereumValidator interface {
	ValidateTransaction(transactionID string, expectedAccount common.Address) (*ValidateTransactionResponse, error)
}

// ChannelHelper defines the methods needed to interact with a channel
type ChannelHelper interface {
	GetEthereumPaymentAddress(channelID string) (common.Address, error)
	GetStripePaymentAccount(channelID string) (string, error)
	GetStripeCustomerID(channelID string) (string, error)
	SetStripeCustomerID(channelID string, stripeCustomerID string) (*channels.Channel, error)
	GetChannelAdminUserChannels(channelID string) ([]*channels.Channel, error)
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

func getTemplateRequest(templateID string, emailAddress string, tmplData email.TemplateData) (req *email.SendTemplateEmailRequest) {
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

// GetChannelTotalProceeds gets total proceeds for the channel, broken out by payment type
func (s *Service) GetChannelTotalProceeds(channelID string) *ProceedsQueryResult {
	var result ProceedsQueryResult
	s.db.Raw(fmt.Sprintf(`
	SELECT 
	posts.post_type, 
	sum(amount * exchange_rate) as total_amount, 
	sum(amount * exchange_rate ) FILTER (WHERE LOWER(p.currency_code) = 'usd') as usd, 
	sum(amount * exchange_rate) FILTER (WHERE p.currency_code = 'ETH') as eth_usd_amount, 
	sum(amount) FILTER (WHERE p.currency_code = 'ETH')  as ether 
	from payments p 
	inner join posts 
	on p.owner_id::uuid = posts.id and p.owner_type = 'posts' 
	where posts.channel_id = ? 
	group by post_type 
	order by post_type;`), channelID).Scan(&result)
	return &result
}

// GetChannelTotalProceedsByBoostType gets total proceeds for the channel, broken out by payment type
func (s *Service) GetChannelTotalProceedsByBoostType(channelID string, boostType string) *ProceedsQueryResult {
	var result ProceedsQueryResult
	s.db.Raw(fmt.Sprintf(`
	SELECT 
	posts.post_type, 
	sum(amount * exchange_rate) as total_amount, 
	sum(amount * exchange_rate ) FILTER (WHERE LOWER(p.currency_code) = 'usd') as usd, 
	sum(amount * exchange_rate) FILTER (WHERE p.currency_code = 'ETH') as eth_usd_amount, 
	sum(amount) FILTER (WHERE p.currency_code = 'ETH')  as ether 
	from payments p 
	inner join posts 
	on p.owner_id::uuid = posts.id and p.owner_type = 'posts' and p.owner_post_type = ?
	where posts.channel_id = ?
	group by post_type 
	order by post_type;`), boostType, channelID).Scan(&result)
	return &result
}

func (s *Service) sendBoostEthPaymentStartedEmail(emailAddress string, tmplData email.TemplateData) error {
	req := getTemplateRequest(boostEthPaymentStartedEmailTemplateID, emailAddress, tmplData)
	return s.emailer.SendTemplateEmail(req)
}

func (s *Service) sendBoostEthPaymentFinishedEmail(emailAddress string, tmplData email.TemplateData) error {
	req := getTemplateRequest(boostEthPaymentFinishedEmailTemplateID, emailAddress, tmplData)
	return s.emailer.SendTemplateEmail(req)
}

func (s *Service) sendBoostStripePaymentReceiptEmail(emailAddress string, tmplData email.TemplateData) error {
	req := getTemplateRequest(boostStripePaymentReceiptEmailTemplateID, emailAddress, tmplData)
	return s.emailer.SendTemplateEmail(req)
}

func (s *Service) sendBoostPaymentReceivedEmail(emailAddress string, tmplData email.TemplateData) error {
	req := getTemplateRequest(boostPaymentReceivedEmailTemplateID, emailAddress, tmplData)
	return s.emailer.SendTemplateEmail(req)
}

func (s *Service) sendBoostPaymentReceivedNoEmailProvidedEmail(emailAddress string, tmplData email.TemplateData) error {
	req := getTemplateRequest(boostPaymentReceivedNoEmailProvidedEmailTemplateID, emailAddress, tmplData)
	return s.emailer.SendTemplateEmail(req)
}

func (s *Service) sendExternalLinkEthPaymentStartedEmail(emailAddress string, tmplData email.TemplateData) error {
	req := getTemplateRequest(externalLinkEthPaymentStartedEmailTemplateID, emailAddress, tmplData)
	return s.emailer.SendTemplateEmail(req)
}

func (s *Service) sendExternalLinkEthPaymentFinishedEmail(emailAddress string, tmplData email.TemplateData) error {
	req := getTemplateRequest(externalLinkEthPaymentFinishedEmailTemplateID, emailAddress, tmplData)
	return s.emailer.SendTemplateEmail(req)
}

func (s *Service) sendExternalLinkStripePaymentReceiptEmail(emailAddress string, tmplData email.TemplateData) error {
	req := getTemplateRequest(externalLinkStripePaymentReceiptEmailTemplateID, emailAddress, tmplData)
	return s.emailer.SendTemplateEmail(req)
}

// CreateEtherPayment confirm that an Ether transaction is valid and store the result as a Payment in the database
func (s *Service) CreateEtherPayment(ownerChannelID string, ownerType string, ownerPostType string, ownerID string, ownerTitle string, etherPayment EtherPayment, tmplData email.TemplateData) (EtherPayment, error) {
	hash := common.HexToHash(etherPayment.TransactionID)
	if (hash == common.Hash{}) {
		return EtherPayment{}, errors.New("invalid tx id")
	}

	payment := PaymentModel{}
	expectedAddress, err := s.channel.GetEthereumPaymentAddress(ownerChannelID)
	if err != nil {
		return EtherPayment{}, err
	}
	// generate a new ID
	id := uuid.NewV4()

	payment.ID = id.String()
	payment.PaymentType = "ether"
	payment.Reference = etherPayment.TransactionID

	payment.Status = "pending"
	payment.OwnerID = ownerID
	payment.OwnerType = ownerType
	payment.OwnerPostType = ownerPostType
	payment.OwnerChannelID = ownerChannelID
	payment.OwnerTitle = ownerTitle
	payment.CurrencyCode = "ETH"
	payment.ExchangeRate = 0
	payment.Amount = 0
	payment.EmailAddress = etherPayment.EmailAddress

	payment.Data = postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf("{\"PaymentAddress\":\"%v\"}", expectedAddress.String()))}

	payment.PayerChannelID = etherPayment.PayerChannelID
	payment.ShouldPublicize = etherPayment.ShouldPublicize

	if err = s.db.Create(&payment).Error; err != nil {
		log.Errorf("An error occurred: %v\n", err)
		return EtherPayment{}, err
	}

	// only send payment receipt if email is given
	if etherPayment.EmailAddress != "" {
		if ownerPostType == postTypeBoost {
			err = s.sendBoostEthPaymentStartedEmail(etherPayment.EmailAddress, tmplData)
		} else if ownerPostType == postTypeExternalLink {
			err = s.sendExternalLinkEthPaymentStartedEmail(etherPayment.EmailAddress, tmplData)
		} else {
			log.Errorf("Error when sending ETH payment started email. OwnerPostType unknown.")
		}
		if err != nil {
			return EtherPayment{
				PaymentModel: payment,
			}, err
		}
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
	var err2 error
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
			update.Status = paymentComplete
			update.Data = postgres.Jsonb{RawMessage: data}
			update.ExchangeRate = res.ExchangeRate
			update.Amount = res.Amount
			// only send payment receipt if email is given
			if payment.EmailAddress != "" {
				if payment.OwnerPostType == postTypeBoost {
					tmplData := email.TemplateData{
						"payment_amount_eth": res.Amount,
						"payment_amount_usd": res.Amount * res.ExchangeRate,
						"payment_to_address": etherPayment.PaymentAddress,
						"boost_id":           etherPayment.OwnerID,
					}
					err2 = s.sendBoostEthPaymentFinishedEmail(payment.EmailAddress, tmplData)
				} else if payment.OwnerPostType == postTypeExternalLink {
					tmplData := email.TemplateData{
						"payment_amount_eth": res.Amount,
						"payment_amount_usd": res.Amount * res.ExchangeRate,
						"payment_to_address": etherPayment.PaymentAddress,
					}
					err2 = s.sendExternalLinkEthPaymentFinishedEmail(payment.EmailAddress, tmplData)
				} else {
					log.Errorf("Error when sending ETH payment complete email. OwnerPostType unknown.")
				}
			}

			err := s.sendETHPaymentReceivedEmails(payment, res.Amount*res.ExchangeRate)
			if err != nil {
				log.Errorf("Error sending boost payment received email: %v\n", err)
			}
		}
	}

	// set the `data` column to the result of ValidateTransaction

	if err = s.db.Model(&payment).Update(update).Error; err != nil {
		log.Errorf("Error updating payment: %v\n", err)
		return err
	}

	if err2 != nil {
		return err2
	}
	return nil
}

func (s *Service) sendETHPaymentReceivedEmails(payment *PaymentModel, amount float64) error {
	channelAdminChannels, err := s.channel.GetChannelAdminUserChannels(payment.OwnerChannelID)
	if err != nil {
		return err
	}
	receivedTmplData, err := s.getPaymentReceivedEmailTemplateData(amount, *payment, "ETH")
	if err != nil {
		log.Errorf("Error getting email template data for payment received: %v\n", err)
	}
	for _, c := range channelAdminChannels {
		email := c.EmailAddress
		if email != "" {
			if payment.EmailAddress != "" {
				err := s.sendBoostPaymentReceivedEmail(email, receivedTmplData)
				if err != nil {
					return err
				}
			} else {
				err := s.sendBoostPaymentReceivedNoEmailProvidedEmail(email, receivedTmplData)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// GetStripeCustomerInfo returns stripe customer info for display on client
func (s *Service) GetStripeCustomerInfo(channelID string) (StripeCustomerInfo, error) {
	customerID, err := s.channel.GetStripeCustomerID(channelID)
	if err != nil {
		return StripeCustomerInfo{}, err
	}
	if customerID == "" {
		return StripeCustomerInfo{}, nil
	}

	return s.stripe.GetCustomerInfo(customerID)
}

// SavePaymentMethod saves payment method to customer, creates customer if needed
func (s *Service) SavePaymentMethod(channelID string, paymentMethodID string, emailAddress string) (*StripePaymentMethod, error) {
	stripeCustomerID, err := s.channel.GetStripeCustomerID(channelID)

	if err != nil || stripeCustomerID == "" {
		res, err := s.stripe.CreateCustomer(CreateCustomerRequest{
			PaymentMethodID: paymentMethodID,
			Email:           emailAddress,
		})
		if err != nil {
			log.Errorf("Error Creating Stripe Customer")
			return nil, errors.New("error creating stripe customer")
		}
		_, err = s.channel.SetStripeCustomerID(channelID, res.ID)
		if err != nil {
			log.Errorf("Error setting stripe customer ID")
			return nil, errors.New("error setting stripe customer ID")
		}
		return &StripePaymentMethod{
			PaymentMethodID: paymentMethodID,
			CustomerID:      res.ID,
		}, nil
	}
	_, err = s.stripe.AddCustomerCard(AddCustomerCardRequest{
		CustomerID:      stripeCustomerID,
		PaymentMethodID: paymentMethodID,
	})
	if err != nil {
		log.Errorf("Error Adding Card to Stripe Customer")
		return nil, errors.New("error adding card to stripe customer")
	}
	return &StripePaymentMethod{
		PaymentMethodID: paymentMethodID,
		CustomerID:      stripeCustomerID,
	}, nil
}

// RemovePaymentMethod removes a payment method owned by the specified channel, if it exists
func (s *Service) RemovePaymentMethod(paymentMethodID string, channelID string) error {
	stripeCustomerInfo, err := s.GetStripeCustomerInfo(channelID)
	if err != nil {
		return err
	}

	doesChannelHaveThisPaymentMethod := false
	for _, p := range stripeCustomerInfo.PaymentMethods {
		if p.PaymentMethodID == paymentMethodID {
			doesChannelHaveThisPaymentMethod = true
			break
		}
	}
	if !doesChannelHaveThisPaymentMethod {
		return errors.New("channel does not own specified payment method")
	}

	return s.stripe.RemovePaymentMethod(paymentMethodID)
}

// CreateStripePayment will create a Stripe charge and then store the result as a Payment in the database
func (s *Service) CreateStripePayment(ownerChannelID string, ownerType string, ownerPostType string, ownerID string, ownerTitle string, payment StripePayment, tmplData email.TemplateData) (StripePayment, error) {

	stripeAccount, err := s.channel.GetStripePaymentAccount(ownerChannelID)
	if err != nil {
		return StripePayment{}, err
	}

	// generate a stripe charge
	res, err := s.stripe.CreateCharge(CreateChargeRequest{
		Amount:        int64(math.Floor(payment.Amount * 100)),
		SourceToken:   &(payment.PaymentToken),
		StripeAccount: stripeAccount,
		Metadata:      map[string]string{ownerType: ownerID},
	})
	if err != nil {
		return StripePayment{}, err
	}

	// generate a new ID for the payment model
	id := uuid.NewV4()
	payment.ID = id.String()

	payment.PaymentType = payment.Type()

	// set the `data` column to the stripe response
	payment.Data = postgres.Jsonb{RawMessage: json.RawMessage(res.StripeResponseJSON)}
	payment.OwnerID = ownerID
	payment.OwnerType = ownerType
	payment.OwnerPostType = ownerPostType
	payment.OwnerChannelID = ownerChannelID
	payment.OwnerTitle = ownerTitle
	payment.Reference = res.ID
	payment.Status = paymentComplete

	// TODO(dankins): this should be set when we support currencies other than USD
	payment.ExchangeRate = 1

	if err = s.db.Create(&payment).Error; err != nil {
		log.Errorf("An error occurred: %v\n", err)
		return StripePayment{}, err
	}
	// only send payment receipt if email is given
	if payment.EmailAddress != "" {
		if ownerPostType == postTypeBoost {
			err = s.sendBoostStripePaymentReceiptEmail(payment.EmailAddress, tmplData)
		} else if ownerPostType == postTypeExternalLink {
			err = s.sendExternalLinkStripePaymentReceiptEmail(payment.EmailAddress, tmplData)
		} else {
			log.Errorf("Error when sending Stripe payment complete email. OwnerPostType unknown.")
		}
		if err != nil {
			return payment, err
		}
	}

	return payment, nil
}

// ClonePaymentMethod will clone a payment method to the connected account
func (s *Service) ClonePaymentMethod(payerChannelID string, postChannelID string, payment StripePayment) (StripePayment, error) {

	stripeAccount, err := s.channel.GetStripePaymentAccount(postChannelID)
	if err != nil {
		return StripePayment{}, err
	}

	var customerID = ""
	if payerChannelID != "" {
		customerID, err = s.channel.GetStripeCustomerID(payerChannelID)
		if err != nil {
			return StripePayment{}, err
		}
	}

	res, err := s.stripe.ClonePaymentMethod(ClonePaymentMethodRequest{
		PaymentMethodID: payment.PaymentMethodID,
		CustomerID:      customerID,
		StripeAccountID: stripeAccount,
	})
	if err != nil {
		return StripePayment{}, err
	}

	payment.CustomerID = customerID
	payment.PaymentMethodID = res.PaymentMethodID

	return payment, nil
}

// ConfirmStripePaymentIntent sets the status of a stripe payment after payment_intent.succeeded webhook event received
func (s *Service) ConfirmStripePaymentIntent(paymentIntent stripe.PaymentIntent) error {
	var payment PaymentModel
	if err := s.db.Where("reference = ?", paymentIntent.ID).First(&payment).Error; err != nil {
		log.Errorf("Error getting payment: %v\n", err)
		return err
	}

	data, err := json.Marshal(paymentIntent)
	if err != nil {
		log.Errorf("Error unmarshalling payment intent data: %v\n", err)
		return err
	}

	amount := (float64(paymentIntent.Amount) / 100.0)
	// create a payment model to hold the updated fields
	update := &PaymentModel{}
	update.Status = paymentComplete
	update.Amount = amount
	update.Data = postgres.Jsonb{RawMessage: json.RawMessage(data)}
	if err := s.db.Model(&payment).Update(update).Error; err != nil {
		log.Errorf("Error updating payment: %v\n", err)
		return err
	}
	tmplData, err := s.getStripePaymentEmailTemplateData(paymentIntent.Metadata, amount, payment.OwnerPostType)
	if err != nil {
		log.Errorf("Error getting email template data for payment intent: %v\n", err)
	}
	// only send payment receipt if email is given
	if payment.EmailAddress != "" {
		if payment.OwnerPostType == postTypeBoost {
			err := s.sendBoostStripePaymentReceiptEmail(payment.EmailAddress, tmplData)
			if err != nil {
				return err
			}
		} else if payment.OwnerPostType == postTypeExternalLink {
			err := s.sendExternalLinkStripePaymentReceiptEmail(payment.EmailAddress, tmplData)
			if err != nil {
				return err
			}
		} else {
			log.Errorf("Error when sending Stripe payment complete email. OwnerPostType unknown.")
			return errors.New("error when sending Stripe payment successful email")
		}
	}
	channelAdminChannels, err := s.channel.GetChannelAdminUserChannels(payment.OwnerChannelID)
	if err != nil {
		return err
	}
	receivedTmplData, err := s.getPaymentReceivedEmailTemplateData(amount, payment, "Stripe")
	if err != nil {
		log.Errorf("Error getting email template data for payment received: %v\n", err)
	}
	for _, c := range channelAdminChannels {
		email := c.EmailAddress
		if email != "" {
			if payment.EmailAddress != "" {
				err = s.sendBoostPaymentReceivedEmail(email, receivedTmplData)
				if err != nil {
					log.Errorf("Error sending boost payment received email: %v\n", err)
				}
			} else {
				err = s.sendBoostPaymentReceivedNoEmailProvidedEmail(email, receivedTmplData)
				if err != nil {
					log.Errorf("Error sending boost payment received with no email provided email: %v\n", err)
				}
			}
		}
	}

	return nil
}

func (s *Service) getPaymentReceivedEmailTemplateData(amount float64, payment PaymentModel, paymentType string) (email.TemplateData, error) {
	if payment.OwnerPostType == postTypeBoost || payment.OwnerPostType == postTypeExternalLink {
		return (email.TemplateData{
			"boost_title":         payment.OwnerTitle,
			"boost_id":            payment.OwnerID,
			"payer_email_address": payment.EmailAddress,
			"payment_amount_usd":  amount,
			"payment_type":        paymentType,
		}), nil
	}
	return nil, errors.New("NOT IMPLEMENTED")
}

func (s *Service) getStripePaymentEmailTemplateData(metadata map[string]string, amount float64, postType string) (email.TemplateData, error) {
	if postType == postTypeBoost {
		return (email.TemplateData{
			"newsroom_name":      metadata["newsroomName"],
			"boost_short_desc":   metadata["title"],
			"payment_amount_usd": amount,
			"boost_id":           metadata["posts"],
		}), nil
	} else if postType == postTypeExternalLink {
		return (email.TemplateData{
			"newsroom_name":      metadata["newsroomName"],
			"payment_amount_usd": amount,
		}), nil
	}
	return nil, errors.New("NOT IMPLEMENTED")
}

// FailStripePaymentIntent sets the status of a stripe payment after payment_intent.payment_failed webhook event received
func (s *Service) FailStripePaymentIntent(paymentIntentID string) (bool, error) {
	var payment PaymentModel
	if err := s.db.Where("reference = ?", paymentIntentID).First(&payment).Error; err != nil {
		log.Errorf("Error getting payment: %v\n", err)
		return false, err
	}

	// create a payment model to hold the updated fields
	update := &PaymentModel{}
	update.Status = "failed"

	if err := s.db.Model(&payment).Update(update).Error; err != nil {
		log.Errorf("Error updating payment: %v\n", err)
		return false, err
	}

	return true, nil
}

// CreateStripePaymentIntent creates a stripe payment intent and "unconfirmed" payment in DB and returns payment intent
func (s *Service) CreateStripePaymentIntent(ownerChannelID string, ownerType string, postType string, ownerID string, newsroomName string, boostTitle string, payment StripePayment) (StripePaymentIntent, error) {
	stripeAccount, err := s.channel.GetStripePaymentAccount(ownerChannelID)
	if err != nil {
		return StripePaymentIntent{}, err
	}
	paymentIntent, err := s.stripe.CreateStripePaymentIntent(
		CreatePaymentIntentRequest{
			Amount:          int64(math.Floor(payment.Amount * 100)),
			StripeAccount:   stripeAccount,
			Metadata:        map[string]string{ownerType: ownerID, "newsroomName": newsroomName, "title": boostTitle, "postChannelID": ownerChannelID},
			PaymentMethodID: &(payment.PaymentMethodID),
			CustomerID:      &(payment.CustomerID),
		})
	if err != nil {
		return StripePaymentIntent{}, nil
	}

	data, err := json.Marshal(paymentIntent)
	if err != nil {
		log.Errorf("Error unmarshalling payment intent data: %v\n", err)
		return StripePaymentIntent{}, nil
	}

	// generate a new ID for the payment model
	id := uuid.NewV4()
	payment.ID = id.String()

	payment.PaymentType = payment.Type()

	payment.Status = "pending"
	payment.OwnerID = ownerID
	payment.OwnerType = ownerType
	payment.OwnerPostType = postType
	payment.Reference = paymentIntent.ID
	payment.OwnerChannelID = ownerChannelID
	payment.OwnerTitle = boostTitle
	payment.Amount = 0
	payment.Data = postgres.Jsonb{RawMessage: json.RawMessage(data)}

	// TODO(dankins): this should be set when we support currencies other than USD
	payment.ExchangeRate = 1

	if err = s.db.Create(&payment).Error; err != nil {
		log.Errorf("An error occurred: %v\n", err)
		return StripePaymentIntent{}, err
	}

	return paymentIntent, nil
}

// GetPaymentsByPayerChannel returns payments made by a channel, exposes potentially sensitive info
// so should only be called after checking user is authorized to view this data
func (s *Service) GetPaymentsByPayerChannel(channelID string) ([]Payment, error) {
	var pays []PaymentModel
	if err := s.db.Where(&PaymentModel{OwnerType: "posts", PayerChannelID: channelID}).Find(&pays).Error; err != nil {
		log.Errorf("An error occurred: %v\n", err)
		return nil, err
	}

	var paymentsSlice []Payment
	for _, result := range pays {
		payment, err := ModelToInterface(&result)
		if err != nil {
			log.Errorf("An error occurred: %v\n", err)
			return nil, err
		}
		paymentsSlice = append(paymentsSlice, payment)
	}

	return paymentsSlice, nil
}

// GetPayments returns the payments associated with a Post
func (s *Service) GetPayments(postID string) ([]Payment, error) {
	var pays []PaymentModel
	if err := s.db.Where(&PaymentModel{OwnerType: "posts", OwnerID: postID}).Find(&pays).Error; err != nil {
		log.Errorf("An error occurred: %v\n", err)
		return nil, err
	}

	var paymentsSlice []Payment
	for _, result := range pays {
		if !result.ShouldPublicize {
			result.PayerChannelID = ""
		}
		payment, err := ModelToInterface(&result)
		if err != nil {
			log.Errorf("An error occurred: %v\n", err)
			return nil, err
		}
		paymentsSlice = append(paymentsSlice, payment)
	}

	return paymentsSlice, nil
}

// GetGroupedSanitizedPayments returns the payments associated with a Post, grouped by channelID if payment should be publicized
func (s *Service) GetGroupedSanitizedPayments(postID string) ([]*SanitizedPayment, error) {
	var pays []SanitizedPayment

	// nolint: gosec
	stmt := s.db.Raw(fmt.Sprintf(`
		SELECT * FROM(

			SELECT * FROM(
				SELECT SUM(amount * exchange_rate) as usd_equivalent,
					max(created_at) as most_recent_update,  
					payer_channel_id
				FROM payments WHERE owner_id = '%s' AND status = 'complete' AND should_publicize = true GROUP BY payer_channel_id
			) publicized_group

			UNION

			SELECT * FROM( 
				SELECT (amount * exchange_rate) as usd_equivalent,
					created_at as most_recent_update, 
					'' as payer_channel_id
				FROM payments WHERE owner_id = '%s' AND status = 'complete' AND should_publicize = false
			) unpublicized_ungroup

		) data 
		ORDER BY most_recent_update DESC`, postID, postID))

	results := stmt.Scan(&pays)

	if results.Error != nil {
		return nil, results.Error
	}

	var paymentsSlice []*SanitizedPayment
	for _, result := range pays {
		sanitizedPayment := SanitizedPayment{
			UsdEquivalent:    result.UsdEquivalent,
			MostRecentUpdate: result.MostRecentUpdate,
			PayerChannelID:   result.PayerChannelID,
		}
		paymentsSlice = append(paymentsSlice, &sanitizedPayment)
	}

	return paymentsSlice, nil
}

// GetPayment returns the payment with the given ID
func (s *Service) GetPayment(paymentID string) (Payment, error) {
	var paymentModel PaymentModel
	if err := s.db.Where(&PaymentModel{ID: paymentID}).First(&paymentModel).Error; err != nil {
		log.Errorf("An error occurred: %v\n", err)
		return nil, err
	}

	payment, err := ModelToInterface(&paymentModel)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// GetPaymentByReference returns the payment with the given reference
func (s *Service) GetPaymentByReference(reference string) (Payment, error) {
	paymentModel := &PaymentModel{}
	s.db.Where("reference = ?", reference).First(paymentModel)

	if (paymentModel.CreatedAt == time.Time{}) {
		return nil, errors.New("Payment Not Found")
	}

	return ModelToInterface(paymentModel)
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
