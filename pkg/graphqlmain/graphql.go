package graphqlmain

import (
	"fmt"

	log "github.com/golang/glog"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"

	"github.com/joincivil/civil-api-server/pkg/auth"
	graphqlgen "github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/graphql"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/civil-events-processor/pkg/helpers"

	cemail "github.com/joincivil/go-common/pkg/email"
)

const (
	graphQLVersion = "v1"
)

type resolverConfig struct {
	config            *utils.GraphQLConfig
	authService       *auth.Service
	userService       *users.UserService
	jsonbService      *jsonstore.Service
	nrsignupService   *nrsignup.Service
	paymentService    *payments.Service
	postService       *posts.Service
	storefrontService *storefront.Service
	emailListMembers  cemail.ListMemberManager
}

func initResolver(rconfig *resolverConfig) (*graphql.Resolver, error) {
	listingPersister, err := helpers.ListingPersister(rconfig.config, rconfig.config.VersionNumber)
	if err != nil {
		log.Errorf("Error w listingPersister: err: %v", err)
		return nil, err
	}
	contentRevisionPersister, err := helpers.ContentRevisionPersister(rconfig.config, rconfig.config.VersionNumber)
	if err != nil {
		log.Errorf("Error w contentRevisionPersister: err: %v", err)
		return nil, err
	}
	governanceEventPersister, err := helpers.GovernanceEventPersister(rconfig.config, rconfig.config.VersionNumber)
	if err != nil {
		log.Errorf("Error w governanceEventPersister: err: %v", err)
		return nil, err
	}
	challengePersister, err := helpers.ChallengePersister(rconfig.config, rconfig.config.VersionNumber)
	if err != nil {
		log.Errorf("Error w challengePersister: err: %v", err)
		return nil, err
	}
	appealPersister, err := helpers.AppealPersister(rconfig.config, rconfig.config.VersionNumber)
	if err != nil {
		log.Errorf("Error w appealPersister: err: %v", err)
		return nil, err
	}
	pollPersister, err := helpers.PollPersister(rconfig.config, rconfig.config.VersionNumber)
	if err != nil {
		log.Errorf("Error w pollPersister: err: %v", err)
		return nil, err
	}
	userChallengeDataPersister, err := helpers.UserChallengeDataPersister(rconfig.config,
		rconfig.config.VersionNumber)
	if err != nil {
		log.Errorf("Error w userChallengeDataPersister: err: %v", err)
		return nil, err
	}

	return graphql.NewResolver(&graphql.ResolverConfig{
		AuthService:                rconfig.authService,
		ListingPersister:           listingPersister,
		RevisionPersister:          contentRevisionPersister,
		GovEventPersister:          governanceEventPersister,
		ChallengePersister:         challengePersister,
		AppealPersister:            appealPersister,
		PollPersister:              pollPersister,
		UserChallengeDataPersister: userChallengeDataPersister,
		UserService:                rconfig.userService,
		JSONbService:               rconfig.jsonbService,
		NrsignupService:            rconfig.nrsignupService,
		PaymentService:             rconfig.paymentService,
		PostService:                rconfig.postService,
		StorefrontService:          rconfig.storefrontService,
		EmailListMembers:           rconfig.emailListMembers,
	}), nil
}

func debugGraphQLRouting(router chi.Router, graphQlEndpoint string) {
	log.Infof("%v", fmt.Sprintf("/%v/%v", graphQLVersion, graphQlEndpoint))
	router.Handle("/", handler.Playground("GraphQL playground",
		fmt.Sprintf("/%v/%v", graphQLVersion, graphQlEndpoint)))
}

func graphQLRouting(router chi.Router, rconfig *resolverConfig) error {
	resolver, rErr := initResolver(rconfig)
	if rErr != nil {
		log.Fatalf("Error retrieving resolver: err: %v", rErr)
		return rErr
	}

	queryHandler := handler.GraphQL(
		graphqlgen.NewExecutableSchema(
			graphqlgen.Config{Resolvers: resolver},
		),
	)

	router.Handle(
		fmt.Sprintf("/%v/query", graphQLVersion),
		graphql.DataloaderMiddleware(resolver, queryHandler))

	return nil
}
