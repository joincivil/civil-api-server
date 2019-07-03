package channels

import (
	"regexp"
	"strings"

	uuid "github.com/satori/go.uuid"
)

// Service provides methods to interact with Channels
type Service struct {
	persister Persister
}

// NewService builds a new Service instance
func NewService(persister Persister) *Service {
	return &Service{
		persister,
	}
}

// CreateUserChannel creates a channel with type "user"
func (s *Service) CreateUserChannel(userID string) (*Channel, error) {

	// make sure there is not a channel for this user already
	ch, err := s.persister.GetChannelByReference(TypeUser, userID)
	if err != nil && err != ErrorNotFound {
		return nil, err
	}
	if ch != nil {
		return nil, ErrorNotUnique
	}
	return s.persister.CreateChannel(CreateChannelInput{
		CreatorUserID: userID,
		ChannelType:   TypeUser,
		Reference:     userID,
	})
}

// CreateNewsroomChannelInput contains the fields needed to create a newsroom channel
type CreateNewsroomChannelInput struct {
	ContractAddress string
}

// CreateNewsroomChannel creates a channel with type "user"
func (s *Service) CreateNewsroomChannel(userID string, input CreateNewsroomChannelInput) (*Channel, error) {
	channelType := TypeNewsroom
	reference := input.ContractAddress

	// make sure there is not a channel for this newsroom smart contract already
	ch, err := s.persister.GetChannelByReference(channelType, reference)
	if err != nil && err != ErrorNotFound {
		return nil, err
	}
	if ch != nil {
		return nil, ErrorNotUnique
	}

	// TODO(dankins): check if userID.eth_address is on the multisig for `input.ContractAddress` newsroom contract

	return s.persister.CreateChannel(CreateChannelInput{
		CreatorUserID: userID,
		ChannelType:   channelType,
		Reference:     reference,
	})
}

// CreateGroupChannel creates a channel with type "group"
func (s *Service) CreateGroupChannel(userID string, handle string) (*Channel, error) {
	channelType := TypeGroup
	normalizedHandle, err := NormalizeHandle(handle)
	if err != nil {
		return nil, err
	}

	// make sure there is not a channel with this handle already
	ch, err := s.persister.GetChannelByHandle(normalizedHandle)
	if err != nil && err != ErrorNotFound {
		return nil, err
	}
	if ch != nil {
		return nil, ErrorNotUnique
	}

	// groups don't reference anything, so generate a new one
	// TODO(dankins): should this reference a DID on an identity server?
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	reference := id.String()

	return s.persister.CreateChannel(CreateChannelInput{
		CreatorUserID: userID,
		ChannelType:   channelType,
		Reference:     reference,
		Handle:        &handle,
	})
}

// GetChannel saves a new channel
func (s *Service) GetChannel(id string) (*Channel, error) {
	return s.persister.GetChannel(id)
}

// NormalizeHandle takes a string handle and removes
func NormalizeHandle(handle string) (string, error) {
	if !IsValidHandle(handle) {
		return "", ErrorInvalidHandle
	}
	return strings.ToLower(handle), nil
}

// IsValidHandle returns whether the provided handle is valid
func IsValidHandle(handle string) bool {
	matched, err := regexp.Match(`^(\w){1,15}$`, []byte(handle))
	if err != nil {
		return false
	}

	return matched
}
