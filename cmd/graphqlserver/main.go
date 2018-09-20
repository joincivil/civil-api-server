package main

import (
	"flag"
	"fmt"
	log "github.com/golang/glog"
	"net/http"
	"os"
	"strconv"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/joincivil/civil-events-processor/pkg/helpers"
	"github.com/joincivil/civil-events-processor/pkg/utils"

	graphqlgen "github.com/joincivil/civil-api-server/pkg/generated/graphql"
	graphql "github.com/joincivil/civil-api-server/pkg/graphql"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

const (
	defaultPort = "8080"
)

func jsonbPersister(config *utils.GraphQLConfig) (jsonstore.JsonbPersister, error) {
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
	return jsonbPersister, nil
}

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
	jsonbPersister, err := jsonbPersister(config)
	if err != nil {
		log.Errorf("Error w jsonbPersister: err: %v", err)
		return nil, err
	}
	return graphql.NewResolver(
		listingPersister,
		contentRevisionPersister,
		governanceEventPersister,
		jsonbPersister,
	), nil
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

	// TODO(PN): Here is where we can add our own auth middleware
	//router.Use(//Authentication)

	if config.Debug {
		router.Handle("/", handler.Playground("GraphQL playground", "/query"))
		log.Infof("Connect to http://localhost:%v/ for GraphQL playground\n", port)
	}

	resolver, err := initResolver(config)
	if err != nil {
		log.Fatalf("Error retrieving resolver: err: %v", err)
	}

	router.Handle(
		"/query",
		handler.GraphQL(
			graphqlgen.NewExecutableSchema(
				graphqlgen.Config{Resolvers: resolver},
			),
		),
	)
	log.Infof("Connect to http://localhost:%v/query for Civil GraphQL\n", port)

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("Error starting graphql service: err: %v", err)
	}

}
