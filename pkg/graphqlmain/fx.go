package graphqlmain

import (
	"flag"
	"os"

	log "github.com/golang/glog"
	"github.com/joho/godotenv"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/civil-events-processor/pkg/helpers"
	pconfig "github.com/joincivil/go-common/pkg/config"
	"go.uber.org/fx"
)

// MainModule provides the main module for the graphql server
var MainModule = fx.Options(
	GraphqlModule,
	EventProcessorModule,
	PubSubModule,
)

// EventProcessorModule defines the dependencies for the Event Processor
var EventProcessorModule = fx.Options(
	fx.Provide(
		ProvideVersionNumber,
		ProvidePersisterConfig,
		helpers.ContentRevisionPersister,
		helpers.GovernanceEventPersister,
		helpers.ChallengePersister,
		helpers.ListingPersister,
		helpers.AppealPersister,
		helpers.UserChallengeDataPersister,
		helpers.PollPersister,
	),
)

// ProvideVersionNumber extracts VersionNumber from the config and is needed by the EventProcessorModule
func ProvideVersionNumber(config *utils.GraphQLConfig) string {
	return config.VersionNumber
}

// ProvidePersisterConfig takes the GraphQLConfig and returns pconfig.PersisterConfig
func ProvidePersisterConfig(config *utils.GraphQLConfig) pconfig.PersisterConfig {
	return config
}

// BuildConfig initializes the config
func BuildConfig() *utils.GraphQLConfig {
	config := &utils.GraphQLConfig{}
	flag.Usage = func() {
		config.OutputUsage()
		os.Exit(0)
	}
	flag.Parse()
	env := os.Getenv("GRAPHQL_ENV")
	if "" == env {
		env = "development"
	}

	err := godotenv.Load(".env." + env + ".local")
	if err != nil {
		log.Errorf("Did not load .env.%v.local", env)
	}
	if "test" != env {
		err := godotenv.Load(".env.local")
		if err != nil {
			log.Errorf("Did not load .env.local")
		}
	}
	err = godotenv.Load(".env." + env)
	if err != nil {
		log.Errorf("Did not load .env." + env)
	}
	err = godotenv.Load()
	if err != nil {
		log.Errorf("Did not load .env")
	}

	err = config.PopulateFromEnv()
	if err != nil {
		config.OutputUsage()
		log.Errorf("Invalid graphql config: err: %v\n", err)
		os.Exit(2)
	}

	return config
}
