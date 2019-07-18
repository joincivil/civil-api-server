package payments

import (
	"encoding/json"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
)

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
