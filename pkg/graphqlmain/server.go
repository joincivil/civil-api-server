package graphqlmain

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/airswap"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/rs/cors"
)

const (
	defaultPort                 = "8080"
	invoicingVersion            = "v1"
	checkbookUpdaterRunFreqSecs = 60 * 5 // 5 mins
)

var (
	validCorsOrigins = []string{
		"*",
	}
)

// RunServer starts up the GraphQL server
// Normally called from main.go
func RunServer(config *utils.GraphQLConfig) error {
	log.Infof("proto %v", config.ApproveGrantProtoHost)
	port := strconv.Itoa(config.Port)
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

	// Setup the API services
	err := enableAPIServices(router, config, port)
	if err != nil {
		log.Fatalf("Error setting up services: err: %v", err)
	}

	// Health REST endpoint
	err = healthCheckRouting(router)
	if err != nil {
		log.Fatalf("Error setting up health check: err: %v", err)
	}

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("Error starting api service: err: %v", err)
	}

	return nil
}

func enableAPIServices(router chi.Router, config *utils.GraphQLConfig, port string) error {
	deps, err := initDependencies(config)
	if err != nil {
		log.Fatalf("Error initializing dependencies: err: %v", err)
		return err
	}

	// Enable authentication/authorization handling
	router.Use(auth.Middleware(deps.jwtGenerator))

	// GraphQL Query Endpoint (Crawler/KYC)
	if config.EnableGraphQL {
		rconfig := &resolverConfig{
			config:            config,
			invoicePersister:  deps.invoicePersister,
			authService:       deps.authService,
			userService:       deps.userService,
			jsonbService:      deps.jsonbService,
			nrsignupService:   deps.nrsignupService,
			tokenFoundry:      deps.tokenFoundry,
			onfido:            deps.onfido,
			storefrontService: deps.storefrontService,
			emailListMembers:  deps.mailchimp,
		}
		err = graphQLRouting(router, rconfig)
		if err != nil {
			log.Fatalf("Error setting up graphql routing: err: %v", err)
		}
		log.Infof(
			"Connect to http://localhost:%v/%v/query for Civil GraphQL\n",
			port,
			graphQLVersion,
		)
		// GraphQL Debug Console
		if config.Debug {
			debugGraphQLRouting(router, "query")
			log.Infof("Connect to http://localhost:%v/ for GraphQL playground\n", port)
		}
	}

	// Invoicing REST endpoints
	if config.EnableInvoicing {
		err = invoicingRouting(router, deps.checkbookIO, deps.invoicePersister, deps.emailer, config.CheckbookTest)
		if err != nil {
			log.Fatalf("Error setting up invoicing routing: err: %v", err)
		}
		log.Infof(
			"Connect to http://localhost:%v/%v/invoicing/send for invoicing\n",
			port,
			invoicingVersion,
		)
		log.Infof(
			"Connect to http://localhost:%v/%v/invoicing/cb for checkbook webhook\n",
			port,
			invoicingVersion,
		)

		updater := invoicing.NewCheckoutIOUpdater(
			deps.checkbookIO,
			deps.invoicePersister,
			deps.emailer,
			checkbookUpdaterRunFreqSecs,
		)
		go updater.Run()
	}

	// KYC REST endpoints
	if config.EnableKYC {
		err = kycRouting(router, config, deps.onfido, deps.emailer)
		if err != nil {
			log.Fatalf("Error setting up KYC routing: err: %v", err)
		}
		log.Infof(
			"Connect to http://localhost:%v/%v/kyc/cb for onfido webhook\n",
			port,
			invoicingVersion,
		)
	}

	// Newsroom Signup REST endpoints
	err = nrsignupRouting(router, config, deps.nrsignupService, deps.jwtGenerator)
	if err != nil {
		log.Fatalf("Error setting up newsroom signup routing: err: %v", err)
	}
	log.Infof(
		"Connect to http://localhost:%v/%v/nrsignup/grantapprove for grant approval webhook\n",
		port,
		invoicingVersion,
	)

	// airswap REST endpoints
	airswap.EnableAirswapRouting(router, deps.storefrontService)

	return nil
}
