package graphqlmain

import (
	"context"
	"fmt"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/gqlerror"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/graphql"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/email"
	cemail "github.com/joincivil/go-common/pkg/email"
	"go.uber.org/fx"

	cerrors "github.com/joincivil/go-common/pkg/errors"

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
		NewGorm,
		initDiscourseListingMapPersister,
		initDiscourseService,
		auth.NewAuthServiceFromConfig,
		initStorefrontService,
		initErrorReporter,
		initIPFS,
		tokencontroller.NewService,
		func(config *utils.GraphQLConfig) *utils.JwtTokenGenerator {
			return utils.NewJwtTokenGenerator([]byte(config.JwtSecret))
		},
		func(config *utils.GraphQLConfig) *email.Emailer {
			return cemail.NewEmailer(config.SendgridKey)
		},
		// convert Mailchimp API to email.ListMemberManager interface
		func(config *utils.GraphQLConfig) email.ListMemberManager {
			return cemail.NewMailchimpAPI(config.MailchimpKey)
		},
	),
)

func debugGraphQLRouting(router chi.Router, graphQlEndpoint string) {
	log.Infof("%v", fmt.Sprintf("/%v/%v", graphQLVersion, graphQlEndpoint))
	router.Handle("/", handler.Playground("GraphQL playground",
		fmt.Sprintf("/%v/%v", graphQLVersion, graphQlEndpoint)))
}

func graphQLRouting(router chi.Router, errorReporter cerrors.ErrorReporter, resolver *graphql.Resolver) error {

	queryHandler := handler.GraphQL(
		graphqlgen.NewExecutableSchema(
			graphqlgen.Config{Resolvers: resolver},
		),
		handler.ErrorPresenter(
			func(ctx context.Context, e error) *gqlerror.Error {
				// Send the error to the error reporter
				err := errors.Cause(e)
				log.Errorf("gql error: %+v", err)
				errorReporter.Error(err, nil)
				return gqlgen.DefaultErrorPresenter(ctx, err)
			},
		),
		handler.RecoverFunc(func(ctx context.Context, err interface{}) error {
			// Send the error to the error reporter
			switch val := err.(type) {
			case error:
				errorReporter.Error(errors.Cause(val), nil)
				log.Errorf("gql panic error: %+v", errors.Cause(val))
			}
			return fmt.Errorf("Internal server error: %v", err)
		}),
	)

	router.Handle(
		fmt.Sprintf("/%v/query", graphQLVersion),
		graphql.DataloaderMiddleware(resolver, queryHandler))

	return nil
}
