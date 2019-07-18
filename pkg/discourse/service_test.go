package discourse_test

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/discourse"

	ceth "github.com/joincivil/go-common/pkg/eth"
)

type InMemoryPersister struct {
	store map[string]*discourse.ListingMap
}

func (i *InMemoryPersister) RetrieveListingMaps(listingAddress []string) ([]*discourse.ListingMap, error) {
	// Not implemented, just for interface
	return nil, nil
}

func (i *InMemoryPersister) RetrieveListingMap(listingAddress string) (*discourse.ListingMap, error) {
	listingAddress = ceth.NormalizeEthAddress(listingAddress)
	val, ok := i.store[listingAddress]
	if !ok {
		return nil, errors.New("error retrieving listingmap")
	}
	return val, nil
}

func (i *InMemoryPersister) SaveListingMap(ldm *discourse.ListingMap) error {
	addr := ldm.ListingAddressAsAddr()
	i.store[addr.Hex()] = ldm
	return nil
}

func setupNewService() (*discourse.Service, error) {
	listingMapPersister := &InMemoryPersister{store: map[string]*discourse.ListingMap{}}
	return discourse.NewService(listingMapPersister), nil
}

func TestSaveRetrieveDiscourseTopicID(t *testing.T) {
	service, err := setupNewService()
	if err != nil {
		t.Errorf("Error returning service: err: %v", err)
	}

	addr := common.HexToAddress("0x49fd8f1d3e6f88a4d08cd4a6e445f848e9475caf")
	topicID := int64(100)

	err = service.SaveDiscourseTopicID(addr, topicID)
	if err != nil {
		t.Errorf("Error saving discourse topic ID: err: %v", err)
	}

	retrievedTopicID, err := service.RetrieveDiscourseTopicID(addr)
	if err != nil {
		t.Errorf("Error retrieving discourse topic ID: err: %v", err)
	}

	if retrievedTopicID != topicID {
		t.Errorf("Topic ID should have matched")
	}
}
