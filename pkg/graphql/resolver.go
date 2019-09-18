// NOTE(PN): gqlgen does not update this file if major updates to the schema are made.
// To completely update, need to move this file and run gqlgen again and replace
// the code.  Fixed when gqlgen matures a bit more?

package graphql

import (
	"errors"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	"strings"

	pmodel "github.com/joincivil/civil-events-processor/pkg/model"
	"go.uber.org/fx"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/discourse"
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
	AuthService                  *auth.Service
	ListingPersister             pmodel.ListingPersister
	ParameterPersister           pmodel.ParameterPersister
	ParamProposalPersister       pmodel.ParamProposalPersister
	GovEventPersister            pmodel.GovernanceEventPersister
	RevisionPersister            pmodel.ContentRevisionPersister
	ChallengePersister           pmodel.ChallengePersister
	AppealPersister              pmodel.AppealPersister
	PollPersister                pmodel.PollPersister
	UserChallengeDataPersister   pmodel.UserChallengeDataPersister
	DiscourseListingMapPersister discourse.ListingMapPersister
	ChannelService               *channels.Service
	UserService                  *users.UserService
	JSONbService                 *jsonstore.Service
	NewsroomService              newsrooms.Service
	NrsignupService              *nrsignup.Service
	PaymentService               *payments.Service
	PostService                  *posts.Service
	StorefrontService            *storefront.Service
	DiscourseService             *discourse.Service
	EmailListMembers             cemail.ListMemberManager
	LowercaseAddr                *bool `optional:"true"`
	ErrorReporter                cerrors.ErrorReporter
}

// NewResolver is a convenience function to init a Resolver struct
func NewResolver(config ResolverConfig) *Resolver {
	return &Resolver{
		authService:                  config.AuthService,
		listingPersister:             config.ListingPersister,
		parameterPersister:           config.ParameterPersister,
		paramProposalPersister:       config.ParamProposalPersister,
		revisionPersister:            config.RevisionPersister,
		govEventPersister:            config.GovEventPersister,
		challengePersister:           config.ChallengePersister,
		appealPersister:              config.AppealPersister,
		pollPersister:                config.PollPersister,
		channelService:               config.ChannelService,
		userChallengeDataPersister:   config.UserChallengeDataPersister,
		discourseListingMapPersister: config.DiscourseListingMapPersister,
		userService:                  config.UserService,
		jsonbService:                 config.JSONbService,
		newsroomService:              config.NewsroomService,
		nrsignupService:              config.NrsignupService,
		paymentService:               config.PaymentService,
		postService:                  config.PostService,
		storefrontService:            config.StorefrontService,
		discourseService:             config.DiscourseService,
		emailListMembers:             config.EmailListMembers,
		lowercaseAddr:                config.LowercaseAddr,
		errorReporter:                config.ErrorReporter,
	}
}

// Resolver is the main resolver for the GraphQL endpoint
type Resolver struct {
	authService                  *auth.Service
	listingPersister             pmodel.ListingPersister
	parameterPersister           pmodel.ParameterPersister
	paramProposalPersister       pmodel.ParamProposalPersister
	revisionPersister            pmodel.ContentRevisionPersister
	govEventPersister            pmodel.GovernanceEventPersister
	challengePersister           pmodel.ChallengePersister
	appealPersister              pmodel.AppealPersister
	pollPersister                pmodel.PollPersister
	userChallengeDataPersister   pmodel.UserChallengeDataPersister
	discourseListingMapPersister discourse.ListingMapPersister
	userService                  *users.UserService
	jsonbService                 *jsonstore.Service
	nrsignupService              *nrsignup.Service
	channelService               *channels.Service
	newsroomService              newsrooms.Service
	paymentService               *payments.Service
	postService                  *posts.Service
	storefrontService            *storefront.Service
	discourseService             *discourse.Service
	emailListMembers             cemail.ListMemberManager
	lowercaseAddr                *bool
	errorReporter                cerrors.ErrorReporter
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
