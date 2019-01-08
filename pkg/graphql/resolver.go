//go:generate gorunpkg github.com/99designs/gqlgen

// NOTE(PN): gqlgen does not update this file if major updates to the schema are made.
// To completely update, need to move this file and run gqlgen again and replace
// the code.  Fixed when gqlgen matures a bit more?

package graphql

import (
	pmodel "github.com/joincivil/civil-events-processor/pkg/model"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/kyc"
	"github.com/joincivil/civil-api-server/pkg/tokenfoundry"
	"github.com/joincivil/civil-api-server/pkg/users"
)

// ResolverConfig is the config params for the Resolver
type ResolverConfig struct {
	AuthService         *auth.Service
	InvoicePersister    *invoicing.PostgresPersister
	ListingPersister    pmodel.ListingPersister
	GovEventPersister   pmodel.GovernanceEventPersister
	RevisionPersister   pmodel.ContentRevisionPersister
	ChallengePersister  pmodel.ChallengePersister
	AppealPersister     pmodel.AppealPersister
	PollPersister       pmodel.PollPersister
	UserPersister       users.UserPersister
	OnfidoAPI           *kyc.OnfidoAPI
	OnfidoTokenReferrer string
	TokenFoundry        *tokenfoundry.API
	UserService         *users.UserService
	JsonbPersister      jsonstore.JsonbPersister
}

// NewResolver is a convenience function to init a Resolver struct
func NewResolver(config *ResolverConfig) *Resolver {
	return &Resolver{
		authService:         config.AuthService,
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
		jsonbPersister:      config.JsonbPersister,
	}
}

// Resolver is the main resolver for the GraphQL endpoint
type Resolver struct {
	authService         *auth.Service
	invoicePersister    *invoicing.PostgresPersister
	listingPersister    pmodel.ListingPersister
	revisionPersister   pmodel.ContentRevisionPersister
	govEventPersister   pmodel.GovernanceEventPersister
	challengePersister  pmodel.ChallengePersister
	appealPersister     pmodel.AppealPersister
	pollPersister       pmodel.PollPersister
	userPersister       users.UserPersister
	onfidoAPI           *kyc.OnfidoAPI
	onfidoTokenReferrer string
	tokenFoundry        *tokenfoundry.API
	userService         *users.UserService
	jsonbPersister      jsonstore.JsonbPersister
}

// Query is the resolver for the Query type
func (r *Resolver) Query() graphql.QueryResolver {
	return &queryResolver{r}
}

// Mutation is the resolver for the Mutation type
func (r *Resolver) Mutation() graphql.MutationResolver {
	return &mutationResolver{r}
}

type queryResolver struct{ *Resolver }

type mutationResolver struct{ *Resolver }
