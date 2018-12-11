//go:generate gorunpkg github.com/99designs/gqlgen

// NOTE(PN): gqlgen does not update this file if major updates to the schema are made.
// To completely update, need to move this file and run gqlgen again and replace
// the code.  Fixed when gqlgen matures a bit more?

package graphql

import (
	model "github.com/joincivil/civil-events-processor/pkg/model"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/eth"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/kyc"
	"github.com/joincivil/civil-api-server/pkg/tokenfoundry"
	"github.com/joincivil/civil-api-server/pkg/users"
)

// ResolverConfig is the config params for the Resolver
type ResolverConfig struct {
	AuthService         *auth.Service
	EthService          *eth.Service
	InvoicePersister    *invoicing.PostgresPersister
	ListingPersister    model.ListingPersister
	GovEventPersister   model.GovernanceEventPersister
	RevisionPersister   model.ContentRevisionPersister
	ChallengePersister  model.ChallengePersister
	AppealPersister     model.AppealPersister
	PollPersister       model.PollPersister
	UserPersister       users.UserPersister
	OnfidoAPI           *kyc.OnfidoAPI
	OnfidoTokenReferrer string
	TokenFoundry        *tokenfoundry.API
	UserService         *users.UserService
}

// NewResolver is a convenience function to init a Resolver struct
func NewResolver(config *ResolverConfig) *Resolver {
	return &Resolver{
		authService:         config.AuthService,
		ethService:          config.EthService,
		invoicePersister:    config.InvoicePersister,
		listingPersister:    config.ListingPersister,
		revisionPersister:   config.RevisionPersister,
		govEventPersister:   config.GovEventPersister,
		challengePersister:  config.ChallengePersister,
		appealPersister:     config.AppealPersister,
		pollPersister:       config.PollPersister,
		userPersister:       config.UserPersister,
		onfidoAPI:           config.OnfidoAPI,
		onfidoTokenReferrer: config.OnfidoTokenReferrer,
		tokenFoundry:        config.TokenFoundry,
		userService:         config.UserService,
	}
}

// Resolver is the main resolver for the GraphQL endpoint
type Resolver struct {
	authService         *auth.Service
	ethService          *eth.Service
	invoicePersister    *invoicing.PostgresPersister
	listingPersister    model.ListingPersister
	revisionPersister   model.ContentRevisionPersister
	govEventPersister   model.GovernanceEventPersister
	challengePersister  model.ChallengePersister
	appealPersister     model.AppealPersister
	pollPersister       model.PollPersister
	userPersister       users.UserPersister
	onfidoAPI           *kyc.OnfidoAPI
	onfidoTokenReferrer string
	tokenFoundry        *tokenfoundry.API
	userService         *users.UserService
}

// Query is the resolver for the Query type
func (r *Resolver) Query() graphql.QueryResolver {
	return &queryResolver{r}
}

// Mutation is the resolver for the Mutation type
func (r *Resolver) Mutation() graphql.MutationResolver {
	return &mutationResolver{r}
}

// Subscription is the resolver for the Subscription type
func (r *Resolver) Subscription() graphql.SubscriptionResolver {
	return &subscriptionResolver{r}
}

type queryResolver struct{ *Resolver }

type mutationResolver struct{ *Resolver }

type subscriptionResolver struct{ *Resolver }
