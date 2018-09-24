package main

import (
	"errors"
	"flag"
	"fmt"
	log "github.com/golang/glog"
	"net/http"
	"os"
	"strconv"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"github.com/joincivil/civil-events-processor/pkg/helpers"

	graphqlgen "github.com/joincivil/civil-api-server/pkg/generated/graphql"
	graphql "github.com/joincivil/civil-api-server/pkg/graphql"
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/utils"
)

const (
	defaultPort = "8080"

	graphQLVersion   = "v1"
	invoicingVersion = "v1"

	// checkbookUpdaterRunFreqSecs = 60 * 5
	checkbookUpdaterRunFreqSecs = 30
)

var (
	validCorsOrigins = []string{
		"*",
	}
)

func initResolver(config *utils.GraphQLConfig) (*graphql.Resolver, error) {
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
	return graphql.NewResolver(
		listingPersister,
		contentRevisionPersister,
		governanceEventPersister,
	), nil
}

func debugGraphQLRouting(router chi.Router) {
	router.Handle("/", handler.Playground("GraphQL playground",
		fmt.Sprintf("/%v/query", graphQLVersion)))
}

func graphQLRouting(router chi.Router, config *utils.GraphQLConfig) error {
	resolver, rErr := initResolver(config)
	if rErr != nil {
		log.Fatalf("Error retrieving resolver: err: %v", rErr)
		return rErr
	}

	router.Handle(
		fmt.Sprintf("/%v/query", graphQLVersion),
		handler.GraphQL(
			graphqlgen.NewExecutableSchema(
				graphqlgen.Config{Resolvers: resolver},
			),
		),
	)
	return nil
}

func invoicePersister(config *utils.GraphQLConfig) (*invoicing.PostgresPersister, error) {
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
	)
	return checkbookIOClient, nil
}

func invoicingRouting(router chi.Router, client *invoicing.CheckbookIO,
	persister *invoicing.PostgresPersister) error {
	invoicingConfig := &invoicing.SendInvoiceHandlerConfig{
		CheckbookIOClient: client,
		InvoicePersister:  persister,
	}
	whConfig := &invoicing.CheckbookIOWebhookConfig{
		InvoicePersister: persister,
	}
	router.Route(fmt.Sprintf("/%v/invoicing", invoicingVersion), func(r chi.Router) {
		r.Post("/send", invoicing.SendInvoiceHandler(invoicingConfig))
		r.Post("/cb", invoicing.CheckbookIOWebhookHandler(whConfig))
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

	// TODO(PN): Here is where we can add our own auth middleware
	//router.Use(//Authentication)

	// GraphQL Debug Console
	if config.Debug {
		debugGraphQLRouting(router)
		log.Infof("Connect to http://localhost:%v/ for GraphQL playground\n", port)
	}

	// GraphQL Query Endpoint
	if config.EnableGraphQL {
		err = graphQLRouting(router, config)
		if err != nil {
			log.Fatalf("Error setting up graphql routing: err: %v", err)
		}
		log.Infof(
			"Connect to http://localhost:%v/%v/query for Civil GraphQL\n",
			port,
			graphQLVersion,
		)
	}

	// REST invoicing endpoint
	if config.EnableInvoicing {
		persister, perr := invoicePersister(config)
		if perr != nil {
			log.Fatalf("Error setting up invoicing persister: err: %v", perr)
		}

		checkbookIOClient, cerr := invoiceCheckbookIO(config)
		if cerr != nil {
			log.Fatalf("Error setting up invoicing client: err: %v", cerr)
		}

		err = invoicingRouting(router, checkbookIOClient, persister)
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

		updater := invoicing.NewCheckoutIOUpdater(checkbookIOClient, persister, checkbookUpdaterRunFreqSecs)
		go updater.Run()
	}

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("Error starting api service: err: %v", err)
	}

}
