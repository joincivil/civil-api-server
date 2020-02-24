package graphqlmain

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/airswap"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/graphql"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/utils"
	cerrors "github.com/joincivil/go-common/pkg/errors"
	"github.com/rs/cors"
	"go.uber.org/fx"
)

const (
	defaultPort      = "8080"
	invoicingVersion = "v1"
)

var (
	validCorsOrigins = []string{
		"*",
	}
)

// ServerDeps define the fields needed to run the Server
type ServerDeps struct {
	fx.In
	Config                *utils.GraphQLConfig
	Resolver              *graphql.Resolver
	ErrorReporter         cerrors.ErrorReporter
	JwtGenerator          *utils.JwtTokenGenerator
	NewsroomSignupService *nrsignup.Service
	StorefrontService     *storefront.Service
	PaymentService        *payments.Service
	PostService           *posts.Service
	ChannelService        *channels.Service
	NewsroomService       newsrooms.Service
	Router                chi.Router
}

// NewRouter builds a new chi router
func NewRouter(lc fx.Lifecycle, config *utils.GraphQLConfig) chi.Router {
	log.Infof("proto %v", config.ApproveGrantProtoHost)
	port := strconv.Itoa(config.GqlPort)
	if port == "" {
		port = defaultPort
	}
	router := chi.NewRouter()

	// Some middleware bits for tracking
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	cors := cors.New(cors.Options{
		AllowedOrigins:   validCorsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
		Debug:            true,
	})
	router.Use(cors.Handler)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("Starting HTTP server.")
			go func() {
				err := http.ListenAndServe(":"+port, router)
				if err != nil {
					log.Errorf("Error starting HTTP server, %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping HTTP server.")
			return nil
		},
	})

	return router
}

// RunServer starts up the GraphQL server
func RunServer(deps ServerDeps) error {

	router := deps.Router

	// Setup the API services
	err := enableAPIServices(router, "8080", deps)
	if err != nil {
		log.Fatalf("Error setting up services: err: %v", err)
	}

	// Health REST endpoint
	err = healthCheckRouting(router)
	if err != nil {
		log.Fatalf("Error setting up health check: err: %v", err)
	}

	return nil
}

func enableAPIServices(router chi.Router, port string, deps ServerDeps) error {

	// Enable authentication/authorization handling
	router.Use(auth.Middleware(deps.JwtGenerator))

	err := graphQLRouting(router, deps.ErrorReporter, deps.Resolver)
	if err != nil {
		log.Fatalf("Error setting up graphql routing: err: %v", err)
	}
	log.Infof(
		"Connect to http://localhost:%v/%v/query for Civil GraphQL\n",
		port,
		graphQLVersion,
	)
	// GraphQL Debug Console
	if deps.Config.Debug {
		debugGraphQLRouting(router, "query")
		log.Infof("Connect to http://localhost:%v/ for GraphQL playground\n", port)
	}

	// Newsroom Signup REST endpoints
	err = nrsignupRouting(deps)
	if err != nil {
		log.Fatalf("Error setting up newsroom signup routing: err: %v", err)
	}
	log.Infof(
		"Connect to http://localhost:%v/%v/nrsignup/grantapprove for grant approval webhook\n",
		port,
		invoicingVersion,
	)

	err = deps.PaymentService.WebhookRouting(router, deps.Config.StripeWebhookSigningSecret)
	if err != nil {
		log.Fatalf("Error setting up webhook routing: err: %v", err)
	}

	// airswap REST endpoints
	airswap.EnableAirswapRouting(router, deps.StorefrontService)

	return nil
}
