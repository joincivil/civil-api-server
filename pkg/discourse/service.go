package discourse

import (
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
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
// Returns 0 for topic id if none found for address
func (s *Service) RetrieveDiscourseTopicID(listingAddress common.Address) (int64, error) {
	if listingAddress == (common.Address{}) {
		return 0, errors.New("empty listing address")
	}

	ldm, err := s.listingMapPersister.RetrieveListingMap(listingAddress.Hex())
	if err != nil {
		if err != cpersist.ErrPersisterNoResults {
			return 0, errors.Wrap(err, "error retrieving listing map")
		}
		return 0, nil
	}

	return ldm.TopicID, nil
}

// SaveDiscourseTopicID saves a Discourse topic ID for a listing address
func (s *Service) SaveDiscourseTopicID(listingAddress common.Address, topicID int64) error {
	if listingAddress == (common.Address{}) {
		return errors.New("empty listing address")
	}
	if topicID <= 0 {
		return errors.Errorf("invalid topic id: %v", topicID)
	}

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
