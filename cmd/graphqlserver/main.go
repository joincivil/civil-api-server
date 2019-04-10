package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/joincivil/go-common/pkg/eth"

	"github.com/joincivil/civil-api-server/pkg/airswap"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"

	log "github.com/golang/glog"

	"github.com/99designs/gqlgen/handler"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"github.com/joincivil/civil-events-processor/pkg/helpers"

	"github.com/joincivil/civil-api-server/pkg/auth"
	graphqlgen "github.com/joincivil/civil-api-server/pkg/generated/graphql"
	graphql "github.com/joincivil/civil-api-server/pkg/graphql"
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/kyc"
	"github.com/joincivil/civil-api-server/pkg/tokenfoundry"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	cemail "github.com/joincivil/go-common/pkg/email"
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

type resolverConfig struct {
	config            *utils.GraphQLConfig
	invoicePersister  *invoicing.PostgresPersister
	authService       *auth.Service
	userService       *users.UserService
	jsonbService      *jsonstore.Service
	nrsignupService   *nrsignup.Service
	tokenFoundry      *tokenfoundry.API
	onfido            *kyc.OnfidoAPI
	storefrontService *storefront.Service
	emailListMembers  cemail.ListMemberManager
}

func initResolver(rconfig *resolverConfig) (*graphql.Resolver, error) {
	listingPersister, err := helpers.ListingPersister(rconfig.config)
	if err != nil {
		log.Errorf("Error w listingPersister: err: %v", err)
		return nil, err
	}
	contentRevisionPersister, err := helpers.ContentRevisionPersister(rconfig.config)
	if err != nil {
		log.Errorf("Error w contentRevisionPersister: err: %v", err)
		return nil, err
	}
	governanceEventPersister, err := helpers.GovernanceEventPersister(rconfig.config)
	if err != nil {
		log.Errorf("Error w governanceEventPersister: err: %v", err)
		return nil, err
	}
	challengePersister, err := helpers.ChallengePersister(rconfig.config)
	if err != nil {
		log.Errorf("Error w challengePersister: err: %v", err)
		return nil, err
	}
	appealPersister, err := helpers.AppealPersister(rconfig.config)
	if err != nil {
		log.Errorf("Error w appealPersister: err: %v", err)
		return nil, err
	}
	pollPersister, err := helpers.PollPersister(rconfig.config)
	if err != nil {
		log.Errorf("Error w pollPersister: err: %v", err)
		return nil, err
	}

	return graphql.NewResolver(&graphql.ResolverConfig{
		AuthService:         rconfig.authService,
		InvoicePersister:    rconfig.invoicePersister,
		ListingPersister:    listingPersister,
		RevisionPersister:   contentRevisionPersister,
		GovEventPersister:   governanceEventPersister,
		ChallengePersister:  challengePersister,
		AppealPersister:     appealPersister,
		PollPersister:       pollPersister,
		OnfidoAPI:           rconfig.onfido,
		OnfidoTokenReferrer: rconfig.config.OnfidoReferrer,
		TokenFoundry:        rconfig.tokenFoundry,
		UserService:         rconfig.userService,
		JSONbService:        rconfig.jsonbService,
		NrsignupService:     rconfig.nrsignupService,
		StorefrontService:   rconfig.storefrontService,
		EmailListMembers:    rconfig.emailListMembers,
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
	err = persister.RunMigrations()
	if err != nil {
		return nil, fmt.Errorf("Error running migrations: err: %v", err)
	}
	return persister, nil
}

func initUserService(config *utils.GraphQLConfig, userPersister *users.PostgresPersister,
	tokenControllerService *tokencontroller.Service) (
	*users.UserService, error) {
	if userPersister == nil {
		var perr error
		userPersister, perr = initUserPersister(config)
		if perr != nil {
			return nil, perr
		}
	}
	userService := users.NewUserService(userPersister, tokenControllerService)
	if userService == nil {
		return nil, fmt.Errorf("User service was not initialized")
	}
	return userService, nil

}

func initJsonbPersister(config *utils.GraphQLConfig) (jsonstore.JsonbPersister, error) {
	jsonbPersister, err := jsonstore.NewPostgresPersister(
		config.PostgresAddress(),
		config.PostgresPort(),
		config.PostgresUser(),
		config.PostgresPw(),
		config.PostgresDbname(),
	)
	if err != nil {
		return nil, err
	}
	err = jsonbPersister.CreateTables()
	if err != nil {
		return nil, fmt.Errorf("Error creating tables: err: %v", err)
	}
	err = jsonbPersister.CreateIndices()
	if err != nil {
		return nil, fmt.Errorf("Error creating indices: err: %v", err)
	}
	err = jsonbPersister.RunMigrations()
	if err != nil {
		return nil, fmt.Errorf("Error running migrations: err: %v", err)
	}
	return jsonbPersister, nil
}

func initJsonbService(config *utils.GraphQLConfig, jsonbPersister jsonstore.JsonbPersister) (
	*jsonstore.Service, error) {
	if jsonbPersister == nil {
		var perr error
		jsonbPersister, perr = initJsonbPersister(config)
		if perr != nil {
			return nil, perr
		}
	}
	jsonbService := jsonstore.NewJsonbService(jsonbPersister)
	return jsonbService, nil
}

func initNrsignupService(config *utils.GraphQLConfig, client bind.ContractBackend,
	emailer *cemail.Emailer, userService *users.UserService, jsonbService *jsonstore.Service,
	jwtGenerator *auth.JwtTokenGenerator) (*nrsignup.Service, error) {
	nrsignupService, err := nrsignup.NewNewsroomSignupService(
		client,
		emailer,
		userService,
		jsonbService,
		jwtGenerator,
		config.ApproveGrantProtoHost,
		config.ContractAddresses["civilparameterizer"],
		config.RegistryAlertsID,
	)
	if err != nil {
		return nil, err
	}
	return nrsignupService, nil
}

func initStorefrontService(config *utils.GraphQLConfig, ethHelper *eth.Helper,
	userService *users.UserService, mailchimp *cemail.MailchimpAPI) (*storefront.Service, error) {
	emailLists := storefront.NewMailchimpServiceEmailLists(mailchimp)

	return storefront.NewService(
		config.ContractAddresses["CVLToken"],
		config.TokenSaleAddresses,
		ethHelper,
		userService,
		emailLists,
	)
}

func initAuthService(config *utils.GraphQLConfig, emailer *cemail.Emailer,
	userService *users.UserService, jwtGenerator *auth.JwtTokenGenerator) (*auth.Service, error) {
	return auth.NewAuthService(
		userService,
		jwtGenerator,
		emailer,
		config.AuthEmailSignupTemplates,
		config.AuthEmailLoginTemplates,
		config.SignupLoginProtoHost,
		config.RefreshTokenBlacklist,
	)
}

func initTokenFoundryAPI(config *utils.GraphQLConfig) *tokenfoundry.API {
	return tokenfoundry.NewAPI(
		"https://tokenfoundry.com",
		config.TokenFoundryUser,
		config.TokenFoundryPassword,
	)
}

func initOnfidoAPI(config *utils.GraphQLConfig) *kyc.OnfidoAPI {
	return kyc.NewOnfidoAPI(
		kyc.ProdAPIURL,
		config.OnfidoKey,
	)
}

func initETHHelper(config *utils.GraphQLConfig) (*eth.Helper, error) {
	if config.EthAPIURL != "" {
		// todo(dankins): we don't actually need any private keys yet, but we will for CIVIL-5
		accounts := map[string]string{}
		if config.EthereumDefaultPrivateKey != "" {
			log.Infof("Initialized default Ethereum account\n")
			accounts["default"] = config.EthereumDefaultPrivateKey
		}
		ethHelper, err := eth.NewETHClientHelper(config.EthAPIURL, accounts)
		if err != nil {
			return nil, err
		}
		log.Infof("Connected to Ethereum using %v\n", config.EthAPIURL)
		return ethHelper, nil
	}

	ethHelper, err := eth.NewSimulatedBackendHelper()
	if err != nil {
		return nil, err
	}
	log.Infof("Connected to Ethereum using Simulated Backend\n")
	return ethHelper, nil
}

func initTokenControllerService(config *utils.GraphQLConfig, ethHelper *eth.Helper) (
	*tokencontroller.Service, error) {
	return tokencontroller.NewService(config.ContractAddresses["CivilTokenController"], ethHelper)
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
	persister *invoicing.PostgresPersister, emailer *cemail.Emailer, testMode bool) error {
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
	emailer *cemail.Emailer) error {

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

func nrsignupRouting(router chi.Router, config *utils.GraphQLConfig,
	nrsignupService *nrsignup.Service, tokenGenerator *auth.JwtTokenGenerator) error {

	grantApproveConfig := &nrsignup.NewsroomSignupApproveGrantConfig{
		NrsignupService: nrsignupService,
		TokenGenerator:  tokenGenerator,
	}

	// Set some rate limiters for the invoice handlers
	limiter := tollbooth.NewLimiter(2, nil) // 2 req/sec max
	limiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	limiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	router.Route(fmt.Sprintf("/%v/nrsignup", invoicingVersion), func(r chi.Router) {
		r.Route(fmt.Sprintf("/grantapprove/{%v}", nrsignup.GrantApproveTokenURLParam), func(r chi.Router) {
			r.Use(tollbooth_chi.LimitHandler(limiter))
			r.Get("/", nrsignup.NewsroomSignupApproveGrantHandler(grantApproveConfig))
		})
	})
	return nil
}

type dependencies struct {
	emailer                *cemail.Emailer
	mailchimp              *cemail.MailchimpAPI
	jwtGenerator           *auth.JwtTokenGenerator
	invoicePersister       *invoicing.PostgresPersister
	checkbookIO            *invoicing.CheckbookIO
	userService            *users.UserService
	authService            *auth.Service
	jsonbService           *jsonstore.Service
	nrsignupService        *nrsignup.Service
	tokenFoundry           *tokenfoundry.API
	onfido                 *kyc.OnfidoAPI
	ethHelper              *eth.Helper
	storefrontService      *storefront.Service
	tokenControllerService *tokencontroller.Service
}

func initDependencies(config *utils.GraphQLConfig) (*dependencies, error) {
	var err error

	ethHelper, err := initETHHelper(config)
	if err != nil {
		log.Fatalf("Error w init ETH Helper: err: %v", err)
		return nil, err
	}

	tokenControllerService, err := initTokenControllerService(config, ethHelper)
	if err != nil {
		log.Fatalf("Error w init tokenControllerService Service: err: %v", err)
		return nil, err
	}

	var checkbookIO *invoicing.CheckbookIO
	if config.EnableInvoicing {
		checkbookIO, err = invoiceCheckbookIO(config)
		if err != nil {
			log.Fatalf("Error setting up invoicing client: err: %v", err)
		}
	}

	invoicePersister, err := initInvoicePersister(config)
	if err != nil {
		log.Fatalf("Error setting up invoicing persister: err: %v", err)
		return nil, err
	}

	jwtGenerator := auth.NewJwtTokenGenerator([]byte(config.JwtSecret))
	tokenFoundry := initTokenFoundryAPI(config)
	onfido := initOnfidoAPI(config)

	var emailer *cemail.Emailer
	if config.SendgridKey != "" {
		emailer = cemail.NewEmailer(config.SendgridKey)
	}

	var mailchimpAPI *cemail.MailchimpAPI
	if config.MailchimpKey != "" {
		mailchimpAPI = cemail.NewMailchimpAPI(config.MailchimpKey)
	}

	userService, err := initUserService(config, nil, tokenControllerService)
	if err != nil {
		log.Fatalf("Error w init userService: err: %v", err)
		return nil, err
	}

	jsonbService, err := initJsonbService(config, nil)
	if err != nil {
		log.Fatalf("Error w init jsonbService: err: %v", err)
		return nil, err
	}
	nrsignupService, err := initNrsignupService(
		config,
		ethHelper.Blockchain,
		emailer,
		userService,
		jsonbService,
		jwtGenerator,
	)
	if err != nil {
		log.Fatalf("Error w init newsroom signup service: err: %v", err)
		return nil, err
	}

	authService, err := initAuthService(
		config,
		emailer,
		userService,
		jwtGenerator,
	)
	if err != nil {
		log.Fatalf("Error w init auth service: err: %v", err)
		return nil, err
	}

	storefrontService, err := initStorefrontService(
		config,
		ethHelper,
		userService,
		mailchimpAPI,
	)
	if err != nil {
		log.Fatalf("Error w init Storefront Service: err: %v", err)
		return nil, err
	}

	return &dependencies{
		emailer:                emailer,
		mailchimp:              mailchimpAPI,
		jwtGenerator:           jwtGenerator,
		invoicePersister:       invoicePersister,
		checkbookIO:            checkbookIO,
		userService:            userService,
		authService:            authService,
		jsonbService:           jsonbService,
		nrsignupService:        nrsignupService,
		tokenFoundry:           tokenFoundry,
		onfido:                 onfido,
		ethHelper:              ethHelper,
		storefrontService:      storefrontService,
		tokenControllerService: tokenControllerService,
	}, nil

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

func healthCheckRouting(router chi.Router) error {
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK")) // nolint: errcheck
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
	err = enableAPIServices(router, config, port)
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

}
