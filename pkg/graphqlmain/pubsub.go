package graphqlmain

import (
	log "github.com/golang/glog"

	"github.com/joincivil/civil-api-server/pkg/events"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/civil-events-processor/pkg/helpers"
)

// RunTokenEventsWorkers starts up the CvlToken events pubsub worker(s)
// Setting this up to live on it own one day
func RunTokenEventsWorkers(config *utils.GraphQLConfig, quit chan bool) error {
	tokenPersister, err := helpers.TokenTransferPersister(config, config.PersisterVersion)
	if err != nil {
		return err
	}
	userPersister, err := initUserPersister(config)
	if err != nil {
		return err
	}
	userService, err := initUserService(config, userPersister, nil)
	if err != nil {
		return err
	}

	transferHandler := events.NewCvlTokenTransferEventHandler(
		tokenPersister,
		userService,
		config.TokenSaleAddresses,
		"6933914",
	)

	handlers := []events.EventHandler{transferHandler}
	workers, err := events.NewWorkers(&events.WorkersConfig{
		PubSubProjectID:        config.PubSubProjectID,
		PubSubTopicName:        config.PubSubTokenTopicName,
		PubSubSubscriptionName: config.PubSubTokenSubName,
		NumWorkers:             1,
		QuitChan:               quit,
		EventHandlers:          handlers,
	})
	if err != nil {
		return err
	}

	go workers.Start()
	log.Infof("TokenEventsWorkers started")
	return nil
}
