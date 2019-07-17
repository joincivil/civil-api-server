package discourse

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
)

// NewService is a convenience function to init a new Discourse data service struct
func NewService(listingMapPersister ListingMapPersister) *Service {
	return &Service{
		listingMapPersister: listingMapPersister,
	}
}

// Service provide methods for the Discourse data service
type Service struct {
	listingMapPersister ListingMapPersister
}

// RetrieveDiscourseTopicID returns the Discourse topic ID for a listing address
func (s *Service) RetrieveDiscourseTopicID(listingAddress common.Address) (int64, error) {
	ldm, err := s.listingMapPersister.RetrieveListingMap(listingAddress.Hex())
	if err != nil {
		fmt.Printf("err retrieving listing map: %v", err)
		return 0, errors.Wrap(err, "error retrieving listing map")
	}
	return ldm.TopicID, err
}

// SaveDiscourseTopicID saves a Discourse topic ID for a listing address
func (s *Service) SaveDiscourseTopicID(listingAddress common.Address, topicID int64) error {
	ldm := &ListingMap{
		ListingAddress: listingAddress.Hex(),
		TopicID:        topicID,
	}
	err := s.listingMapPersister.SaveListingMap(ldm)
	if err != nil {
		return errors.Wrap(err, "error saving listing map")
	}
	return nil
}
