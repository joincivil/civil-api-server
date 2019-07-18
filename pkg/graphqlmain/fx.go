package graphqlmain

import (
	"flag"
	"os"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/civil-events-processor/pkg/helpers"
	pconfig "github.com/joincivil/go-common/pkg/config"
	"go.uber.org/fx"
)

// RuntimeModule provides concrete implementations
var RuntimeModule = fx.Options(
	channels.RuntimeModule,
	users.RuntimeModule,
	payments.RuntimeModule,
	storefront.RuntimeModule,
)

// MainModule provides the main module for the graphql server
var MainModule = fx.Options(
	RuntimeModule,
	GraphqlModule,
	EventProcessorModule,
	PubSubModule,
	fx.Invoke(RunPersisterMigrations),
	fx.Invoke(RunServer),
	fx.Invoke(payments.PaymentUpdaterCron),
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
	err := config.PopulateFromEnv()
	if err != nil {
		config.OutputUsage()
		log.Errorf("Invalid graphql config: err: %v\n", err)
		os.Exit(2)
	}

	return config
}
