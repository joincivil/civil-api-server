// NOTE(PN): gqlgen does not update this file if major updates to the schema are made.
// To completely update, need to move this file and run gqlgen again and replace
// the code.  Fixed when gqlgen matures a bit more?

package graphql

import (
	"errors"
	"strings"

	pmodel "github.com/joincivil/civil-events-processor/pkg/model"
	"go.uber.org/fx"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/users"

	cemail "github.com/joincivil/go-common/pkg/email"
	cerrors "github.com/joincivil/go-common/pkg/errors"
)

var (
	// ErrUserNotAuthorized is a generic error for unauthorized users
	ErrUserNotAuthorized = errors.New("User is not authorized")

	// ErrAccessDenied is a generic error for unauthorized access
	ErrAccessDenied = errors.New("Access denied")

	// ResponseOK is a generic OK response string
	ResponseOK = "ok"

	// ResponseError is a generic error response string
	ResponseError = "error"

	// ResponseNotImplemented is a generic response string for non-implemented endpoints
	ResponseNotImplemented = "not implemented"
)

// ResolverConfig is the config params for the Resolver
type ResolverConfig struct {
	fx.In
	AuthService                *auth.Service
	ListingPersister           pmodel.ListingPersister
	GovEventPersister          pmodel.GovernanceEventPersister
	RevisionPersister          pmodel.ContentRevisionPersister
	ChallengePersister         pmodel.ChallengePersister
	AppealPersister            pmodel.AppealPersister
	PollPersister              pmodel.PollPersister
	UserChallengeDataPersister pmodel.UserChallengeDataPersister
	UserService                *users.UserService
	JSONbService               *jsonstore.Service
	NrsignupService            *nrsignup.Service
	PaymentService             *payments.Service
	PostService                *posts.Service
	StorefrontService          *storefront.Service
	EmailListMembers           cemail.ListMemberManager
	LowercaseAddr              *bool
	ErrorReporter              cerrors.ErrorReporter
}

// NewResolver is a convenience function to init a Resolver struct
func NewResolver(config ResolverConfig) *Resolver {
	return &Resolver{
		authService:                config.AuthService,
		listingPersister:           config.ListingPersister,
		revisionPersister:          config.RevisionPersister,
		govEventPersister:          config.GovEventPersister,
		challengePersister:         config.ChallengePersister,
		appealPersister:            config.AppealPersister,
		pollPersister:              config.PollPersister,
		userChallengeDataPersister: config.UserChallengeDataPersister,
		userService:                config.UserService,
		jsonbService:               config.JSONbService,
		nrsignupService:            config.NrsignupService,
		paymentService:             config.PaymentService,
		postService:                config.PostService,
		storefrontService:          config.StorefrontService,
		emailListMembers:           config.EmailListMembers,
		lowercaseAddr:              config.LowercaseAddr,
		errorReporter:              config.ErrorReporter,
	}
}

// Resolver is the main resolver for the GraphQL endpoint
type Resolver struct {
	authService                *auth.Service
	listingPersister           pmodel.ListingPersister
	revisionPersister          pmodel.ContentRevisionPersister
	govEventPersister          pmodel.GovernanceEventPersister
	challengePersister         pmodel.ChallengePersister
	appealPersister            pmodel.AppealPersister
	pollPersister              pmodel.PollPersister
	userChallengeDataPersister pmodel.UserChallengeDataPersister
	userService                *users.UserService
	jsonbService               *jsonstore.Service
	nrsignupService            *nrsignup.Service
	paymentService             *payments.Service
	postService                *posts.Service
	storefrontService          *storefront.Service
	emailListMembers           cemail.ListMemberManager
	lowercaseAddr              *bool
	errorReporter              cerrors.ErrorReporter
}

// Query is the resolver for the Query type
func (r *Resolver) Query() graphql.QueryResolver {
	return &queryResolver{r}
}

// Mutation is the resolver for the Mutation type
func (r *Resolver) Mutation() graphql.MutationResolver {
	return &mutationResolver{r}
}

// DetermineAddrCase determines the case of an address
func (r *Resolver) DetermineAddrCase(addr string) string {
	if *r.lowercaseAddr {
		return strings.ToLower(addr)
	}
	return addr
}

type queryResolver struct{ *Resolver }

type mutationResolver struct{ *Resolver }
