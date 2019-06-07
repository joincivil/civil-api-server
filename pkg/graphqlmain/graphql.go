package graphqlmain

import (
	"context"
	"fmt"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/graphql"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/email"
	cemail "github.com/joincivil/go-common/pkg/email"
	"go.uber.org/fx"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"

	graphqlgen "github.com/joincivil/civil-api-server/pkg/generated/graphql"
)

const (
	graphQLVersion = "v1"
)

// GraphqlModule provides the graphql server
var GraphqlModule = fx.Options(
	fx.Provide(
		NewRouter,
		graphql.NewResolver,
		BuildConfig,
		initJsonbPersister,
		initGorm,
		initETHHelper,
		initTokenControllerService,
		initJsonbService,
		initNrsignupService,
		auth.NewAuthServiceFromConfig,
		initStorefrontService,
		initPaymentService,
		posts.NewDBPostPersister,
		posts.NewService,
		users.NewPersisterFromGorm,
		initUserService,
		func(config *utils.GraphQLConfig) *auth.JwtTokenGenerator {
			return auth.NewJwtTokenGenerator([]byte(config.JwtSecret))
		},
		func(config *utils.GraphQLConfig) *email.Emailer {
			return cemail.NewEmailer(config.SendgridKey)
		},
		// convert Mailchimp API to email.ListMemberManager interface
		func(config *utils.GraphQLConfig) email.ListMemberManager {
			return cemail.NewMailchimpAPI(config.MailchimpKey)
		},
	),
	fx.Invoke(RunPersisterMigrations),
	fx.Invoke(RunServer),
)

func debugGraphQLRouting(router chi.Router, graphQlEndpoint string) {
	log.Infof("%v", fmt.Sprintf("/%v/%v", graphQLVersion, graphQlEndpoint))
	router.Handle("/", handler.Playground("GraphQL playground",
		fmt.Sprintf("/%v/%v", graphQLVersion, graphQlEndpoint)))
}

func graphQLRouting(router chi.Router, resolver *graphql.Resolver) error {

	queryHandler := handler.GraphQL(
		graphqlgen.NewExecutableSchema(
			graphqlgen.Config{Resolvers: resolver},
		),
		handler.ErrorPresenter(
			func(ctx context.Context, e error) *gqlerror.Error {
				// Send the error to the error reporter
				rconfig.errorReporter.Error(e, nil)
				return gqlgen.DefaultErrorPresenter(ctx, e)
			},
		),
		handler.RecoverFunc(func(ctx context.Context, err interface{}) error {
			// Send the error to the error reporter
			switch val := err.(type) {
			case error:
				rconfig.errorReporter.Error(val, nil)
			}
			return fmt.Errorf("Internal server error: %v", err)
		}),
	)

	router.Handle(
		fmt.Sprintf("/%v/query", graphQLVersion),
		graphql.DataloaderMiddleware(resolver, queryHandler))

	return nil
}
