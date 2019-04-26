package processor

import (
	"encoding/json"

	crawlermodel "github.com/joincivil/civil-events-crawler/pkg/model"

	"github.com/joincivil/go-common/pkg/pubsub"
)

func (e *EventProcessor) pubSub(event *crawlermodel.Event, topicName string) error {
	if !e.pubsubEnabled(topicName) {
		return nil
	}

	// NOTE(IS): We only want to send notifications on watched events
	if event.RetrievalMethod() == crawlermodel.Filterer {
		return nil
	}

	payload, err := e.pubSubBuildPayload(event, topicName)
	if err != nil {
		return err
	}

	return e.googlePubSub.Publish(payload)
}

// PubSubMessage is a struct that represents a message to be published to the pubsub.
type PubSubMessage struct {
	TxHash string `json:"txHash"`
}

func (e *EventProcessor) pubSubBuildPayload(event *crawlermodel.Event,
	topicName string) (*pubsub.GooglePubSubMsg, error) {
	msg := &PubSubMessage{TxHash: event.TxHash().Hex()}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	googlePubSubMsg := &pubsub.GooglePubSubMsg{
		Topic:   topicName,
		Payload: string(msgBytes),
	}

	return googlePubSubMsg, nil
}

func (e *EventProcessor) pubsubEnabled(topicName string) bool {
	if e.googlePubSub == nil {
		return false
	}
	if topicName == "" {
		return false
	}
	return true
}
