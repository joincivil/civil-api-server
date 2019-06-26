package graphqlmain

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/golang/glog"
	"go.uber.org/fx"

	"github.com/joincivil/civil-api-server/pkg/events"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/civil-events-processor/pkg/helpers"
	"github.com/joincivil/civil-events-processor/pkg/model"

	"github.com/joincivil/go-common/pkg/pubsub"
)

// PubSubModule initializes and starts EventHandlers
var PubSubModule = fx.Options(
	fx.Provide(
		buildWorkers,
		buildKillChannel,
		buildPubsubConfig,
		buildCvlTokenTransferEventHandler,
		helpers.TokenTransferPersister,
	),
	fx.Invoke(RunTokenEventsWorkers),
)

type QuitChannel chan bool

// PubSubConfig defines the fields needed to start PubSub
type PubSubConfig struct {
	PubSubProjectID      string
	PubSubTokenTopicName string
	PubSubTokenSubName   string
	RegistryListID       string
	TokenSaleAddresses   []common.Address
}

func buildPubsubConfig(cfg *utils.GraphQLConfig) *PubSubConfig {
	return &PubSubConfig{
		PubSubProjectID:      cfg.PubSubProjectID,
		PubSubTokenTopicName: cfg.PubSubTokenTopicName,
		PubSubTokenSubName:   cfg.PubSubTokenSubName,
		RegistryListID:       "6933914",
		TokenSaleAddresses:   cfg.TokenSaleAddresses,
	}
}

// BuildKillChannel builds a channel that is closed when the server is shutting down
func buildKillChannel(lc fx.Lifecycle) QuitChannel {
	quit := make(chan bool)

	return quit
}

func buildCvlTokenTransferEventHandler(tokenPersister model.TokenTransferPersister,
	userService *users.UserService, config *PubSubConfig) *events.CvlTokenTransferEventHandler {
	return events.NewCvlTokenTransferEventHandler(
		tokenPersister,
		userService,
		config.TokenSaleAddresses,
		config.RegistryListID,
	)
}

func buildWorkers(config *PubSubConfig, transferHandler *events.CvlTokenTransferEventHandler, quit QuitChannel) (*pubsub.Workers, error) {

	if config.PubSubProjectID == "" {
		return nil, nil
	}

	handlers := []pubsub.EventHandler{transferHandler}
	return pubsub.NewWorkers(&pubsub.WorkersConfig{
		PubSubProjectID:        config.PubSubProjectID,
		PubSubTopicName:        config.PubSubTokenTopicName,
		PubSubSubscriptionName: config.PubSubTokenSubName,
		NumWorkers:             1,
		QuitChan:               quit,
		EventHandlers:          handlers,
	})
}

// PubSubDependencies defines the dependencies needed for PubSub
type PubSubDependencies struct {
	fx.In
	Config  *PubSubConfig
	Workers *pubsub.Workers `optional:"true"`
	Quit    QuitChannel
}

// RunTokenEventsWorkers starts up the CvlToken events pubsub worker(s)
// Setting this up to live on it own one day
func RunTokenEventsWorkers(deps PubSubDependencies, lc fx.Lifecycle) error {

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if deps.Workers != nil {
				log.Info("Starting PubSub")
				go deps.Workers.Start()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Infof("Closing Quit Channel")
			close(deps.Quit)
			return nil
		},
	})

	log.Infof("TokenEventsWorkers started")
	return nil
}
