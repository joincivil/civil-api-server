package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/ethclient"

	log "github.com/golang/glog"
	"github.com/gorilla/websocket"

	"github.com/99designs/gqlgen/handler"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"github.com/joincivil/civil-events-processor/pkg/helpers"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/eth"
	graphqlgen "github.com/joincivil/civil-api-server/pkg/generated/graphql"
	graphql "github.com/joincivil/civil-api-server/pkg/graphql"
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/jobs"
	"github.com/joincivil/civil-api-server/pkg/kyc"
	"github.com/joincivil/civil-api-server/pkg/tokenfoundry"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"
)

const (
	defaultPort = "8080"

	graphQLVersion   = "v1"
	invoicingVersion = "v1"

	checkbookUpdaterRunFreqSecs = 60 * 5 // 5 mins
)

var (
	validCorsOrigins = []string{
		"*",
	}
)

func initResolver(config *utils.GraphQLConfig, invoicePersister *invoicing.PostgresPersister,
	userPersister *users.PostgresPersister) (*graphql.Resolver, error) {
	listingPersister, err := helpers.ListingPersister(config)
	if err != nil {
		log.Errorf("Error w listingPersister: err: %v", err)
		return nil, err
	}
	contentRevisionPersister, err := helpers.ContentRevisionPersister(config)
	if err != nil {
		log.Errorf("Error w contentRevisionPersister: err: %v", err)
		return nil, err
	}
	governanceEventPersister, err := helpers.GovernanceEventPersister(config)
	if err != nil {
		log.Errorf("Error w governanceEventPersister: err: %v", err)
		return nil, err
	}
	challengePersister, err := helpers.ChallengePersister(config)
	if err != nil {
		log.Errorf("Error w challengePersister: err: %v", err)
		return nil, err
	}
	appealPersister, err := helpers.AppealPersister(config)
	if err != nil {
		log.Errorf("Error w appealPersister: err: %v", err)
		return nil, err
	}
	pollPersister, err := helpers.PollPersister(config)
	if err != nil {
		log.Errorf("Error w pollPersister: err: %v", err)
		return nil, err
	}
	onfido := kyc.NewOnfidoAPI(
		kyc.ProdAPIURL,
		config.OnfidoKey,
	)

	tokenFoundry := tokenfoundry.NewAPI(
		"https://tokenfoundry.com",
		config.TokenFoundryUser,
		config.TokenFoundryPassword,
	)

	userService := users.NewUserService(userPersister)
	if err != nil {
		log.Errorf("Error w userService: err: %v", err)
		return nil, err
	}

	// TODO(dankins): building a new instance of tokenGenerator here, which is a little hacky.
	jwtGenerator := auth.NewJwtTokenGenerator([]byte(config.JwtSecret))
	emailer := utils.NewEmailer(config.SendgridKey)
	authService := auth.NewAuthService(userService, jwtGenerator, emailer)

	log.Info("Ethereum RPC Address: ", config.EthereumRPCAddress)
	blockchain, err := ethclient.Dial(config.EthereumRPCAddress)
	if err != nil {
		log.Errorf("Error starting ETH client: %v", err)
		return nil, err
	}
	jobService := jobs.NewInMemoryJobService()
	txListener := eth.NewTxListener(blockchain, jobService)
	ethService := eth.NewService(txListener)

	return graphql.NewResolver(&graphql.ResolverConfig{
		AuthService:         authService,
		EthService:          ethService,
		InvoicePersister:    invoicePersister,
		ListingPersister:    listingPersister,
		RevisionPersister:   contentRevisionPersister,
		GovEventPersister:   governanceEventPersister,
		ChallengePersister:  challengePersister,
		AppealPersister:     appealPersister,
		PollPersister:       pollPersister,
		UserPersister:       userPersister,
		OnfidoAPI:           onfido,
		OnfidoTokenReferrer: config.OnfidoReferrer,
		TokenFoundry:        tokenFoundry,
		UserService:         userService,
	}), nil
}

func debugGraphQLRouting(router chi.Router, graphQlEndpoint string) {
	log.Infof("%v", fmt.Sprintf("/%v/%v", graphQLVersion, graphQlEndpoint))
	router.Handle("/", handler.Playground("GraphQL playground",
		fmt.Sprintf("/%v/%v", graphQLVersion, graphQlEndpoint)))
}

func graphQLRouting(router chi.Router, config *utils.GraphQLConfig, invoicePersister *invoicing.PostgresPersister,
	userPersister *users.PostgresPersister) error {
	resolver, rErr := initResolver(config, invoicePersister, userPersister)
	if rErr != nil {
		log.Fatalf("Error retrieving resolver: err: %v", rErr)
		return rErr
	}
	queryHandler := handler.GraphQL(
		graphqlgen.NewExecutableSchema(
			graphqlgen.Config{Resolvers: resolver},
		),
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}),
	)
	router.Handle(
		fmt.Sprintf("/%v/query", graphQLVersion),
		graphql.DataloaderMiddleware(resolver, queryHandler))
	return nil
}

func initInvoicePersister(config *utils.GraphQLConfig) (*invoicing.PostgresPersister, error) {
	persister, err := invoicing.NewPostgresPersister(
		config.PostgresAddress(),
		config.PostgresPort(),
		config.PostgresUser(),
		config.PostgresPw(),
		config.PostgresDbname(),
	)
	if err != nil {
		return nil, err
	}
	err = persister.CreateTables()
	if err != nil {
		return nil, fmt.Errorf("Error creating tables: err: %v", err)
	}
	err = persister.CreateIndices()
	if err != nil {
		return nil, fmt.Errorf("Error creating indices: err: %v", err)
	}
	return persister, nil
}

func initUserPersister(config *utils.GraphQLConfig) (*users.PostgresPersister, error) {
	persister, err := users.NewPostgresPersister(
		config.PostgresAddress(),
		config.PostgresPort(),
		config.PostgresUser(),
		config.PostgresPw(),
		config.PostgresDbname(),
	)
	if err != nil {
		return nil, err
	}
	err = persister.CreateTables()
	if err != nil {
		return nil, fmt.Errorf("Error creating tables: err: %v", err)
	}
	err = persister.CreateIndices()
	if err != nil {
		return nil, fmt.Errorf("Error creating indices: err: %v", err)
	}
	return persister, nil
}

func invoiceCheckbookIO(config *utils.GraphQLConfig) (*invoicing.CheckbookIO, error) {
	key := config.CheckbookKey
	secret := config.CheckbookSecret
	test := config.CheckbookTest

	if key == "" || secret == "" {
		return nil, errors.New("Checkbook key and secret required")
	}

	checkbookBaseURL := invoicing.ProdCheckbookIOBaseURL
	if test {
		checkbookBaseURL = invoicing.SandboxCheckbookIOBaseURL
	}

	checkbookIOClient := invoicing.NewCheckbookIO(
		checkbookBaseURL,
		key,
		secret,
		test,
	)
	return checkbookIOClient, nil
}

func invoicingRouting(router chi.Router, client *invoicing.CheckbookIO,
	persister *invoicing.PostgresPersister, emailer *utils.Emailer, testMode bool) error {
	invoicingConfig := &invoicing.SendInvoiceHandlerConfig{
		CheckbookIOClient: client,
		InvoicePersister:  persister,
		Emailer:           emailer,
		TestMode:          testMode,
	}
	whConfig := &invoicing.CheckbookIOWebhookConfig{
		InvoicePersister: persister,
		Emailer:          emailer,
	}

	// Set some rate limiters for the invoice handlers
	limiter := tollbooth.NewLimiter(2, nil) // 2 req/sec max
	limiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	limiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	cblimiter := tollbooth.NewLimiter(10, nil) // 10 req/sec max
	cblimiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	cblimiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	router.Route(fmt.Sprintf("/%v/invoicing", invoicingVersion), func(r chi.Router) {
		r.Route("/send", func(r chi.Router) {
			r.Use(tollbooth_chi.LimitHandler(limiter))
			r.Post("/", invoicing.SendInvoiceHandler(invoicingConfig))
		})

		r.Route("/cb", func(r chi.Router) {
			r.Use(tollbooth_chi.LimitHandler(cblimiter))
			r.Post("/", invoicing.CheckbookIOWebhookHandler(whConfig))
		})
	})
	return nil
}

func kycRouting(router chi.Router, config *utils.GraphQLConfig, onfido *kyc.OnfidoAPI,
	emailer *utils.Emailer) error {

	ofConfig := &kyc.OnfidoWebhookHandlerConfig{
		OnfidoWebhookToken: config.OnfidoWebhookToken,
	}

	cblimiter := tollbooth.NewLimiter(10, nil) // 10 req/sec max
	cblimiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	cblimiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	router.Route(fmt.Sprintf("/%v/kyc", invoicingVersion), func(r chi.Router) {
		r.Route("/cb", func(r chi.Router) {
			r.Use(tollbooth_chi.LimitHandler(cblimiter))
			r.Post("/", kyc.OnfidoWebhookHandler(ofConfig))
		})
	})
	return nil
}

func main() {
	config := &utils.GraphQLConfig{}
	flag.Usage = func() {
		config.OutputUsage()
		os.Exit(0)
	}
	flag.Parse()

	err := config.PopulateFromEnv()
	if err != nil {
		config.OutputUsage()
		log.Errorf("Invalid graphql config: err: %v\n", err)
		os.Exit(2)
	}

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

	tokenGenerator := auth.NewJwtTokenGenerator([]byte(config.JwtSecret))
	router.Use(auth.Middleware(tokenGenerator))

	// set up persisters
	var invoicePersister *invoicing.PostgresPersister
	var userPersister *users.PostgresPersister
	var perr error

	if config.EnableInvoicing || config.EnableGraphQL {
		invoicePersister, perr = initInvoicePersister(config)
		if perr != nil {
			log.Fatalf("Error setting up invoicing persister: err: %v", perr)
		}
	}
	if config.EnableKYC || config.EnableGraphQL {
		userPersister, perr = initUserPersister(config)
		if perr != nil {
			log.Fatalf("Error setting up user persister: err: %v", perr)
		}
	}

	// GraphQL Query Endpoint (Crawler/KYC)
	if config.EnableGraphQL {
		err = graphQLRouting(router, config, invoicePersister, userPersister)
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

		checkbookIOClient, cerr := invoiceCheckbookIO(config)
		if cerr != nil {
			log.Fatalf("Error setting up invoicing client: err: %v", cerr)
		}

		emailer := utils.NewEmailer(config.SendgridKey)
		err = invoicingRouting(router, checkbookIOClient, invoicePersister, emailer, config.CheckbookTest)
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
			checkbookIOClient,
			invoicePersister,
			emailer,
			checkbookUpdaterRunFreqSecs,
		)
		go updater.Run()

	}

	// KYC REST endpoints
	if config.EnableKYC {
		onfido := kyc.NewOnfidoAPI(
			kyc.ProdAPIURL,
			config.OnfidoKey,
		)
		emailer := utils.NewEmailer(config.SendgridKey)
		err = kycRouting(router, config, onfido, emailer)
		if err != nil {
			log.Fatalf("Error setting up KYC routing: err: %v", err)
		}
		log.Infof(
			"Connect to http://localhost:%v/%v/kyc/cb for onfido webhook\n",
			port,
			invoicingVersion,
		)
	}

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("Error starting api service: err: %v", err)
	}

}
