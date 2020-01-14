package graphqlmain

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/golang/glog"
	"github.com/pkg/errors"
	"go.uber.org/fx"

	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/events"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/civil-events-processor/pkg/helpers"
	"github.com/joincivil/civil-events-processor/pkg/model"

	cerrors "github.com/joincivil/go-common/pkg/errors"
	"github.com/joincivil/go-common/pkg/pubsub"
)

// PubSubModule initializes and starts EventHandlers
var PubSubModule = fx.Options(
	fx.Provide(
		buildWorkers,
		buildKillChannel,
		buildPubsubConfig,
		buildCvlTokenTransferEventHandler,
		buildMultiSigEventHandler,
		buildGovernanceEventHandler,
		helpers.TokenTransferPersister,
		// helpers.GovernanceEventPersister,
	),
	fx.Invoke(RunEventsWorkers),
)

// QuitChannel is a channel type that is used to quit goroutines and processes
type QuitChannel chan bool

// PubSubConfig defines the fields needed to start PubSub
type PubSubConfig struct {
	PubSubProjectID           string
	PubSubTokenTopicName      string
	PubSubTokenSubName        string
	PubSubMultiSigTopicName   string
	PubSubMultiSigSubName     string
	PubSubGovernanceTopicName string
	PubSubGovernanceSubName   string
	RegistryListID            string
	TokenSaleAddresses        []common.Address
}

func buildPubsubConfig(cfg *utils.GraphQLConfig) *PubSubConfig {
	return &PubSubConfig{
		PubSubProjectID:           cfg.PubSubProjectID,
		PubSubTokenTopicName:      cfg.PubSubTokenTopicName,
		PubSubTokenSubName:        cfg.PubSubTokenSubName,
		PubSubMultiSigTopicName:   cfg.PubSubMultiSigTopicName,
		PubSubMultiSigSubName:     cfg.PubSubMultiSigSubName,
		PubSubGovernanceTopicName: cfg.PubSubGovernanceTopicName,
		PubSubGovernanceSubName:   cfg.PubSubGovernanceSubName,
		RegistryListID:            "6933914",
		TokenSaleAddresses:        cfg.TokenSaleAddresses,
	}
}

// BuildKillChannel builds a channel that is closed when the server is shutting down
func buildKillChannel(lc fx.Lifecycle) chan struct{} {
	quit := make(chan struct{})

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

func buildMultiSigEventHandler(listingPersister model.ListingPersister,
	userService *users.UserService, channelService *channels.Service) *events.MultiSigEventHandler {
	return events.NewMultiSigEventHandler(
		listingPersister,
		userService,
		channelService,
	)
}

func buildGovernanceEventHandler(
	governanceEventPersister model.GovernanceEventPersister,
	listingPersister model.ListingPersister,
	channelService *channels.Service,
) *events.GovernanceEventHandler {
	return events.NewGovernanceEventHandler(
		governanceEventPersister,
		listingPersister,
		channelService,
	)
}

func buildWorkers(config *PubSubConfig, transferHandler *events.CvlTokenTransferEventHandler, multiSigHandler *events.MultiSigEventHandler, governanceHandler *events.GovernanceEventHandler, quit chan struct{}) ([]*pubsub.Workers, error) {

	if config.PubSubProjectID == "" {
		return nil, nil
	}

	tokenHandlers := []pubsub.EventHandler{transferHandler}
	tokenWorkers, err := pubsub.NewWorkers(&pubsub.WorkersConfig{
		PubSubProjectID:        config.PubSubProjectID,
		PubSubTopicName:        config.PubSubTokenTopicName,
		PubSubSubscriptionName: config.PubSubTokenSubName,
		NumWorkers:             1,
		QuitChan:               quit,
		EventHandlers:          tokenHandlers,
	})
	if err != nil {
		return nil, err
	}

	multiSigHandlers := []pubsub.EventHandler{multiSigHandler}
	multiSigWorkers, err := pubsub.NewWorkers(&pubsub.WorkersConfig{
		PubSubProjectID:        config.PubSubProjectID,
		PubSubTopicName:        config.PubSubMultiSigTopicName,
		PubSubSubscriptionName: config.PubSubMultiSigSubName,
		NumWorkers:             1,
		QuitChan:               quit,
		EventHandlers:          multiSigHandlers,
	})
	if err != nil {
		return nil, err
	}

	governanceHandlers := []pubsub.EventHandler{governanceHandler}
	governanceWorkers, err := pubsub.NewWorkers(&pubsub.WorkersConfig{
		PubSubProjectID:        config.PubSubProjectID,
		PubSubTopicName:        config.PubSubGovernanceTopicName,
		PubSubSubscriptionName: config.PubSubGovernanceSubName,
		NumWorkers:             1,
		QuitChan:               quit,
		EventHandlers:          governanceHandlers,
	})
	if err != nil {
		return nil, err
	}

	return []*pubsub.Workers{tokenWorkers, multiSigWorkers, governanceWorkers}, nil
}

// PubSubDependencies defines the dependencies needed for PubSub
type PubSubDependencies struct {
	fx.In
	Config  *PubSubConfig
	Workers []*pubsub.Workers `optional:"true"`
	Quit    chan struct{}
	ErrRep  cerrors.ErrorReporter
}

// RunEventsWorkers starts up the events pubsub worker(s)
// Setting this up to live on it own one day
func RunEventsWorkers(deps PubSubDependencies, lc fx.Lifecycle) error {

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if deps.Workers != nil && len(deps.Workers) > 0 {
				for _, worker := range deps.Workers {
					log.Info("Starting PubSub")
					go worker.Start()
					// Log and report the errors coming out of the workers
					go func(w *pubsub.Workers) {
						for err := range w.Errors {
							log.Errorf("error from worker: err: %v", err)
							deps.ErrRep.Error(errors.WithMessage(err, "error from worker"), nil)
						}
					}(worker)
				}
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Infof("Closing Quit Channel")
			close(deps.Quit)
			return nil
		},
	})

	log.Infof("EventsWorkers started")
	return nil
}
