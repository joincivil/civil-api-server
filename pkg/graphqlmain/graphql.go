package graphqlmain

import (
	"context"
	"fmt"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/gqlerror"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/graphql"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/email"
	cemail "github.com/joincivil/go-common/pkg/email"
	"github.com/joincivil/go-common/pkg/newsroom"
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
	payments.PaymentModule,
	channels.ChannelModule,
	posts.PostModule,
	users.UserModule,
	storefront.StorefrontModule,
	fx.Provide(
		NewRouter,
		graphql.NewResolver,
		BuildConfig,
		initJsonbPersister,
		NewGorm,
		NewETHHelper,
		initDiscourseListingMapPersister,
		initJsonbService,
		initDiscourseService,
		initNrsignupService,
		auth.NewAuthServiceFromConfig,
		initStorefrontService,
		initErrorReporter,
		newsroom.NewService,
		NewDeployerContractAddresses,
		tokencontroller.NewService,
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
				errorReporter.Error(e, nil)
				return gqlgen.DefaultErrorPresenter(ctx, e)
			},
		),
		handler.RecoverFunc(func(ctx context.Context, err interface{}) error {
			// Send the error to the error reporter
			switch val := err.(type) {
			case error:
				errorReporter.Error(val, nil)
			}
			return fmt.Errorf("Internal server error: %v", err)
		}),
	)

	router.Handle(
		fmt.Sprintf("/%v/query", graphQLVersion),
		graphql.DataloaderMiddleware(resolver, queryHandler))

	return nil
}
