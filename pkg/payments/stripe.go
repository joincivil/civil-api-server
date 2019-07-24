package payments

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
)

const stripeOAuthURI = "https://connect.stripe.com/oauth/token"

// StripeService provides methods to interact with the Stripe payment provider
type StripeService struct {
	apiKey string
}

// CreateChargeRequest contains the data needed to create a charge
type CreateChargeRequest struct {
	Amount        int64
	SourceToken   string
	StripeAccount string
	Metadata      map[string]string
}

// CreateChargeResponse contains the result of a stripe charge
type CreateChargeResponse struct {
	StripeResponseJSON []byte
}

// NewStripeService constructs an instance of the stripe Service
func NewStripeService(apiKey string) *StripeService {
	return &StripeService{
		apiKey: apiKey,
	}
}

// NewStripeServiceFromConfig constructs an instance of the stripe Service
func NewStripeServiceFromConfig(config *utils.GraphQLConfig) *StripeService {
	return &StripeService{
		apiKey: config.StripeAPIKey,
	}
}

// CreateCharge sends a payment to a connected account
func (s *StripeService) CreateCharge(request *CreateChargeRequest) (CreateChargeResponse, error) {

	stripe.Key = s.apiKey

	params := &stripe.ChargeParams{
		Amount:   stripe.Int64(request.Amount),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
	}
	err := params.SetSource(request.SourceToken)
	if err != nil {
		log.Errorf("error creating stripe charge: %v", err)
		return CreateChargeResponse{}, err
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

// CreateCharge sends a payment to a connected account
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
