// +build integration

package events_test

import (
	"sync"
	"testing"
	"time"
	"encoding/json"

	"github.com/joincivil/civil-events-processor/pkg/processor"
	"github.com/joincivil/go-common/pkg/pubsub"
	cpubsub "github.com/joincivil/go-common/pkg/pubsub"

	"github.com/joincivil/civil-api-server/pkg/events"
)

type TestEventHandler struct {
	HandleBool bool
	HandleErr  error
	Msg        *processor.PubSubMessage
	m          sync.Mutex
}

func (t *TestEventHandler) TheMsg() *processor.PubSubMessage {
	t.m.Lock()
	defer t.m.Unlock()
	return t.Msg
}

func (t *TestEventHandler) Handle(msg []byte) (bool, error) {
	t.m.Lock()
	defer t.m.Unlock()
	// Unmarshal into the processor pubsub message
	p := &processor.PubSubMessage{}
	err := json.Unmarshal(msg, p)
	if err != nil {
		return false, err
	}
	t.Msg = p
	return t.HandleBool, t.HandleErr
}

func (t *TestEventHandler) Name() string {
	return "TestHandler"
}

func TestWorkers(t *testing.T) {
	testHandler1 := &TestEventHandler{
		HandleBool: true,
		HandleErr:  nil,
	}
	testHandler2 := &TestEventHandler{
		HandleBool: false,
		HandleErr:  nil,
	}

	eventHandlers := []events.EventHandler{
		testHandler1,
		testHandler2,
	}
	quit := make(chan bool)
	config := &events.WorkersConfig{
		PubSubProjectID:        "civil-media",
		PubSubTopicName:        "governance-events",
		PubSubSubscriptionName: "sub-governance-events",
		NumWorkers:             1,
		QuitChan:               quit,
		EventHandlers:          eventHandlers,
	}

	ps, err := cpubsub.NewGooglePubSub(config.PubSubProjectID)
	if err != nil {
		t.Fatalf("Unable to setup pubsub: err: %v", err)
	}
	err = ps.DeleteSubscription(config.PubSubSubscriptionName)
	if err != nil {
		t.Logf("Error deleting subscription %v", err)
	}
	err = ps.DeleteTopic(config.PubSubTopicName)
	if err != nil {
		t.Logf("Error deleting topic %v", err)
	}
	err = ps.CreateTopic(config.PubSubTopicName)
	if err != nil {
		t.Fatalf("Error creating topic %v", err)
	}
	err = ps.CreateSubscription(config.PubSubTopicName, config.PubSubSubscriptionName)
	if err != nil {
		t.Fatalf("Error creating subscription %v", err)
	}
	err = ps.StartPublishers()
	if err != nil {
		t.Fatalf("Error starting publishers: %v", err)
	}
	defer func() {
		err = ps.DeleteSubscription(config.PubSubSubscriptionName)
		if err != nil {
			t.Fatalf("Error deleting subscription %v", err)
		}
		err = ps.DeleteTopic(config.PubSubTopicName)
		if err != nil {
			t.Fatalf("Error deleting topic %v", err)
		}
	}()

	workers, err := events.NewWorkers(config)
	if err != nil {
		t.Fatalf("Should not have given an error on workers creation: err: %v", err)
	}

	itquit := make(chan bool)

	go func() {
		workers.Start()
		close(itquit)
	}()

	go func() {
		time.Sleep(2 * time.Second)
		if workers.NumActiveWorkers() != 1 {
			t.Errorf("Num of active workers should be 1: %v", workers.NumActiveWorkers())
		}
		msg := &pubsub.GooglePubSubMsg{
			Topic:   config.PubSubTopicName,
			Payload: "{\"txHash\": \"0x4fa779b4dbf20f8df5b4e523c49920858234172492dc4fb477aee4f7abd67899\"}",
		}
		err = ps.Publish(msg)
		if err != nil {
			t.Errorf("Should not have error when publishing: err: %v", err)
		}
		time.Sleep(2 * time.Second)
		close(quit)
	}()

	select {
	case <-itquit:
		if testHandler1.TheMsg() == nil {
			t.Errorf("Pubsub message not sent to handler")
			if testHandler1.TheMsg().TxHash != "0x4fa779b4dbf20f8df5b4e523c49920858234172492dc4fb477aee4f7abd67899" {
				t.Errorf("Pubsub message has the wrong txhash")
			}
		}
	case <-time.After(time.Second * 10):
		t.Error("Should have quit properly")
	}
}

func TestWorkersRecovery(t *testing.T) {
	testHandler1 := &TestEventHandler{
		HandleBool: true,
		HandleErr:  nil,
	}
	testHandler2 := &TestEventHandler{
		HandleBool: false,
		HandleErr:  nil,
	}

	eventHandlers := []events.EventHandler{
		testHandler1,
		testHandler2,
	}
	quit := make(chan bool)
	config := &events.WorkersConfig{
		PubSubProjectID:        "civil-media",
		PubSubTopicName:        "governance-events",
		PubSubSubscriptionName: "sub-governance-events",
		NumWorkers:             1,
		QuitChan:               quit,
		EventHandlers:          eventHandlers,
	}

	workers, err := events.NewWorkers(config)
	if err != nil {
		t.Fatalf("Should not have given an error on workers creation: err: %v", err)
	}

	itquit := make(chan bool)

	go func() {
		workers.Start()
		close(itquit)
	}()

	go func() {
		time.Sleep(2 * time.Second)
		close(quit)
	}()

	<-time.After(time.Second * 3)
}
