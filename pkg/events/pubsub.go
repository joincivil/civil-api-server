package events

import (
	"errors"

	cpubsub "github.com/joincivil/go-common/pkg/pubsub"
)

func initPubSub(projectID string) (*cpubsub.GooglePubSub, error) {
	// If no project ID, quit
	if projectID == "" {
		return nil, errors.New("Need PubSubProjectID")
	}

	ps, err := cpubsub.NewGooglePubSub(projectID)
	if err != nil {
		return nil, err
	}
	return ps, err
}

func initPubSubSubscribers(ps *cpubsub.GooglePubSub, topicName string, subName string) error {
	// If no crawl topic name, quit
	if topicName == "" {
		return errors.New("Pubsub topic name should be specified")
	}
	// If no subscription name, quit
	if subName == "" {
		return errors.New("Pubsub subscription name should be specified")
	}

	return ps.StartSubscribersWithConfig(
		cpubsub.SubscribeConfig{
			Name:    subName,
			AutoAck: false,
		},
	)
}
