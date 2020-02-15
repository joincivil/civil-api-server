package payments

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/applepaydomain"
	"github.com/stripe/stripe-go/card"
	"github.com/stripe/stripe-go/charge"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/paymentintent"
	"github.com/stripe/stripe-go/token"
)

const stripeOAuthURI = "https://connect.stripe.com/oauth/token"

// StripeService provides methods to interact with the Stripe payment provider
type StripeService struct {
	apiKey          string
	applePayDomains []string
}

// CreateChargeRequest contains the data needed to create a charge
type CreateChargeRequest struct {
	Amount        int64
	SourceToken   *string
	CustomerID    *string
	SourceID      *string
	StripeAccount string
	Metadata      map[string]string
}

// CreateChargeResponse contains the result of a stripe charge
type CreateChargeResponse struct {
	ID                 string
	StripeResponseJSON []byte
}

// CreateCustomerRequest contains the data needed to create a customer
type CreateCustomerRequest struct {
	Email         string
	SourceToken   string
	StripeAccount string
}

// CreateCustomerResponse contains the result of creating a customer
type CreateCustomerResponse struct {
	ID string
}

// AddCustomerCardRequest contains the data needed to add a new card to a customer
type AddCustomerCardRequest struct {
	CustomerID  string
	SourceToken string
}

// AddCustomerCardResponse contains the result of adding a new card to a customer
type AddCustomerCardResponse struct {
	ID string
}

// CreatePaymentIntentRequest contains the data needed to create a payment request
type CreatePaymentIntentRequest struct {
	Amount        int64
	CustomerID    *string
	SourceID      *string
	StripeAccount string
	Metadata      map[string]string
}

// NewStripeService constructs an instance of the stripe Service
func NewStripeService(apiKey string, applePayDomains []string) *StripeService {
	return &StripeService{
		apiKey:          apiKey,
		applePayDomains: applePayDomains,
	}
}

// NewStripeServiceFromConfig constructs an instance of the stripe Service
func NewStripeServiceFromConfig(config *utils.GraphQLConfig) *StripeService {
	return &StripeService{
		apiKey:          config.StripeAPIKey,
		applePayDomains: config.StripeApplePayDomains,
	}
}

// GetCustomerInfo returns customer info such as payment sources for display on client
func (s *StripeService) GetCustomerInfo(customerID string) (StripeCustomerInfo, error) {
	stripe.Key = s.apiKey

	cus, err := customer.Get(customerID, nil)
	if err != nil {
		return StripeCustomerInfo{}, err
	}

	sources := make([]StripeSource, cus.Sources.TotalCount)

	for _, s := range cus.Sources.Data {
		source := StripeSource{
			ID:          s.ID,
			Last4Digits: s.Card.Last4,
			ExpMonth:    string(s.Card.ExpMonth),
			ExpYear:     string(s.Card.ExpYear),
		}
		sources = append(sources, source)
	}

	return StripeCustomerInfo{Sources: sources}, nil
}

// AddCustomerCard adds a card to a stripe customer
func (s *StripeService) AddCustomerCard(request *AddCustomerCardRequest) (AddCustomerCardResponse, error) {
	stripe.Key = s.apiKey

	params := &stripe.CardParams{
		Customer: stripe.String(request.CustomerID),
		Token:    stripe.String(request.SourceToken),
	}
	c, err := card.New(params)
	if err != nil {
		return AddCustomerCardResponse{}, err
	}

	return AddCustomerCardResponse{ID: c.ID}, nil
}

// CreateCustomer creates a stripe customer
func (s *StripeService) CreateCustomer(request *CreateCustomerRequest) (CreateCustomerResponse, error) {
	stripe.Key = s.apiKey

	// Create a Customer:
	customerParams := &stripe.CustomerParams{
		Email: stripe.String(request.Email),
	}
	err := customerParams.SetSource(request.SourceToken)
	if err != nil {
		log.Errorf("error creating stripe customer: %v", err)
		return CreateCustomerResponse{}, err
	}
	cus, err := customer.New(customerParams)
	if err != nil {
		log.Errorf("error creating stripe customer: %v", err)
		return CreateCustomerResponse{}, err
	}

	tokenParams := &stripe.TokenParams{
		Customer: stripe.String(cus.ID),
	}
	tokenParams.SetStripeAccount(request.StripeAccount)
	token, err := token.New(tokenParams)
	if err != nil {
		log.Errorf("error creating token: %v", err)
		return CreateCustomerResponse{}, err
	}

	params := &stripe.CustomerParams{}
	params.SetStripeAccount(request.StripeAccount)
	params.SetSource(token.ID)
	_, err = customer.New(params)
	if err != nil {
		log.Errorf("error creating connected customer: %v", err)
		return CreateCustomerResponse{}, err
	}

	return CreateCustomerResponse{ID: cus.ID}, nil
}

// CreateCharge sends a payment to a connected account
func (s *StripeService) CreateCharge(request *CreateChargeRequest) (CreateChargeResponse, error) {
	stripe.Key = s.apiKey

	var params *stripe.ChargeParams
	if request.CustomerID != nil {
		params = &stripe.ChargeParams{
			Amount:   stripe.Int64(request.Amount),
			Currency: stripe.String(string(stripe.CurrencyUSD)),
			Customer: request.CustomerID,
		}
		if request.SourceID != nil {
			err := params.SetSource(*(request.SourceID))
			if err != nil {
				log.Errorf("error creating stripe charge: %v", err)
				return CreateChargeResponse{}, err
			}
		}
	} else {
		params = &stripe.ChargeParams{
			Amount:   stripe.Int64(request.Amount),
			Currency: stripe.String(string(stripe.CurrencyUSD)),
		}

		err := params.SetSource(*(request.SourceToken))
		if err != nil {
			log.Errorf("error creating stripe charge: %v", err)
			return CreateChargeResponse{}, err
		}
	}

	for k, v := range request.Metadata {
		params.AddMetadata(k, v)
	}

	params.SetStripeAccount(request.StripeAccount)

	ch, err := charge.New(params)
	if err != nil {
		log.Errorf("error creating stripe charge: %v", err)
		return CreateChargeResponse{}, err
	}

	if ch.Outcome.NetworkStatus != "approved_by_network" {
		return CreateChargeResponse{}, err
	}

	bytes, err := json.Marshal(ch)
	if err != nil {
		log.Errorf("error creating stripe charge json: %v", err)
		return CreateChargeResponse{}, err
	}

	return CreateChargeResponse{
		StripeResponseJSON: bytes,
		ID:                 ch.ID,
	}, nil

}

// CreateStripePaymentIntent creates a payment intent to be completed on the client
func (s *StripeService) CreateStripePaymentIntent(request CreatePaymentIntentRequest) (StripePaymentIntent, error) {
	stripe.Key = s.apiKey

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(request.Amount),
		Currency: stripe.String(string(stripe.CurrencyUSD)), // @TODO get from input?
		PaymentMethodTypes: []*string{
			stripe.String("card"),
		},
	}
	params.SetStripeAccount(request.StripeAccount)
	for k, v := range request.Metadata {
		params.AddMetadata(k, v)
	}
	// @TODO handle errors
	pi, _ := paymentintent.New(params)

	return StripePaymentIntent{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
	}, nil
}

// https://stripe.com/docs/connect/standard-accounts?origin_team=T9L4Z5JAU#token-request
// "Finalize the account connection" https://stripe.com/docs/connect/quickstart
type responseData struct {
	TokenType            string `json:"token_type"`             // bearer
	StripePublishableKey string `json:"stripe_publishable_key"` //  "{PUBLISHABLE_KEY}"
	Scope                string `json:"scope"`                  //  "read_write"
	Livemode             bool   `json:"livemode"`               // false
	StripeUserID         string `json:"stripe_user_id"`         // "{ACCOUNT_ID}"
	RefreshToken         string `json:"refresh_token"`          // "{REFRESH_TOKEN}"
	AccessToken          string `json:"access_token"`           // "{ACCESS_TOKEN}"
}

// ConnectAccount calls Stripe to finalize a connection to an account given a Stripe auth code
func (s *StripeService) ConnectAccount(code string) (string, error) {
	stripe.Key = s.apiKey
	resp, err := http.PostForm(stripeOAuthURI,
		url.Values{"client_secret": {s.apiKey}, "code": {code}, "grant_type": {"authorization_code"}})

	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		log.Errorf("non-200 response from Stripe: %v, %v", resp.StatusCode, string(body))
		return "", errors.Errorf("status code is not 200")
	}
	data := &responseData{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return "", err
	}

	return data.StripeUserID, nil
}

// GetApplyPayDomains returns the list of domains that have Apple Pay enabled
func (s *StripeService) GetApplyPayDomains(stripeAccountID string) ([]string, error) {
	stripe.Key = s.apiKey
	domains := applepaydomain.List(&stripe.ApplePayDomainListParams{
		ListParams: stripe.ListParams{
			StripeAccount: &stripeAccountID,
		},
	})

	var rtn []string
	for domains.Next() {
		domain := domains.ApplePayDomain()
		rtn = append(rtn, domain.DomainName)
	}

	return rtn, nil
}

// IsApplePayEnabled returns true if the account's Apple Pay enabled domains include the civil's domains
func (s *StripeService) IsApplePayEnabled(stripeAccountID string) (bool, error) {

	domains, err := s.GetApplyPayDomains(stripeAccountID)
	if err != nil {
		return false, errors.Wrap(err, "error getting apple pay domains")
	}

	return IsSubset(s.applePayDomains, domains), nil
}

// EnableApplePay adds configured domains to the connected stripe account's list of Apple Pay domains
func (s *StripeService) EnableApplePay(stripeAccountID string) ([]string, error) {
	var addedDomains []string

	domains, err := s.GetApplyPayDomains(stripeAccountID)
	if err != nil {
		return nil, errors.Wrap(err, "Could not enable Apple Pay")
	}

	domainsToEnable := SetDifference(s.applePayDomains, domains)

	for _, domain := range domainsToEnable {
		newDomain := &stripe.ApplePayDomainParams{
			Params:     stripe.Params{StripeAccount: &stripeAccountID},
			DomainName: &domain,
		}
		_, err := applepaydomain.New(newDomain)

		if err != nil {
			return nil, err
		}

		addedDomains = append(addedDomains, domain)
	}

	return addedDomains, nil
}

// SetDifference returns items in groupA that are not in groupB
func SetDifference(groupA []string, groupB []string) []string {
	var difference []string
	for _, groupAItem := range groupA {
		var included = false
		for _, groupBItem := range groupB {
			if groupBItem == groupAItem {
				included = true
				break
			}
		}

		if !included {
			difference = append(difference, groupAItem)
		}
	}

	return difference
}

// IsSubset returns true if all items in groupA are included in groupB
func IsSubset(groupA []string, groupB []string) bool {
	return len(SetDifference(groupA, groupB)) == 0
}
