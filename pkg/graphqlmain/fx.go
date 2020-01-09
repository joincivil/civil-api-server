package graphqlmain

import (
	"flag"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
	"github.com/joincivil/civil-api-server/pkg/runtime"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/civil-events-processor/pkg/helpers"
	pconfig "github.com/joincivil/go-common/pkg/config"
	"go.uber.org/fx"
)

// MainModule provides the main module for the graphql server
var MainModule = fx.Options(
	runtime.Module,
	GraphqlModule,
	EventProcessorModule,
	PubSubModule,
	fx.Invoke(RunJsonbPersisterMigrations),
	fx.Invoke(RunPersisterMigrations),
	fx.Invoke(RunPostPersisterMigrations),
	fx.Invoke(RunServer),
	fx.Invoke(payments.PaymentUpdaterCron),
)

// EventProcessorModule defines the dependencies for the Event Processor
var EventProcessorModule = fx.Options(
	fx.Provide(
		ProvideVersionNumber,
		ProvidePersisterConfig,
		ProvideSqlxFromGorm,
		helpers.ContentRevisionPersisterFromSqlx,
		helpers.GovernanceEventPersisterFromSqlx,
		helpers.ChallengePersisterFromSqlx,
		helpers.ListingPersisterFromSqlx,
		helpers.ParameterPersisterFromSqlx,
		helpers.ParameterizerPersisterFromSqlx,
		helpers.AppealPersisterFromSqlx,
		helpers.UserChallengeDataPersisterFromSqlx,
		helpers.PollPersisterFromSqlx,
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

// ProvideSqlxFromGorm takes a Gorm instance and returns it's sqlx.DB object
func ProvideSqlxFromGorm(db *gorm.DB) *sqlx.DB {
	return sqlx.NewDb(db.DB(), "postgres")
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
