package channels

import (
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/go-common/pkg/newsroom"
	uuid "github.com/satori/go.uuid"
)

// Service provides methods to interact with Channels
type Service struct {
	persister              Persister
	newsroomMultisigGetter NewsroomMultisigGetter
	userEthAddressGetter   UserEthAddressGetter
}

// NewsroomMultisigGetter describes methods needed to get the members of a newsroom multisig
type NewsroomMultisigGetter interface {
	GetMultisigMembers(newsroomAddress common.Address) ([]common.Address, error)
}

// UserEthAddressGetter describes methods needed to get the ETH addresses of a User
type UserEthAddressGetter interface {
	GetETHAddresses(userID string) ([]common.Address, error)
}

// NewService builds a new Service instance
func NewService(persister Persister, newsroomMultisigGetter NewsroomMultisigGetter, userEthAddressGetter UserEthAddressGetter) *Service {

	return &Service{
		persister,
		newsroomMultisigGetter,
		userEthAddressGetter,
	}
}

// NewServiceWithImplementations builds a new Service instance concrete implementations
func NewServiceWithImplementations(persister *DBPersister, newsroomService *newsroom.Service, userService *users.UserService) *Service {
	return NewService(
		persister,
		newsroomService,
		userService,
	)
}

// GetUserChannels retrieves the Channels a user is a member of
func (s *Service) GetUserChannels(userID string) ([]*ChannelMember, error) {
	return s.persister.GetUserChannels(userID)
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

	// convert contract address string to common.Address
	newsroomAddress := common.HexToAddress(reference)
	if (newsroomAddress == common.Address{}) {
		return nil, ErrorInvalidHandle
	}

	// get the owners of the multisig
	multisigMembers, err := s.newsroomMultisigGetter.GetMultisigMembers(newsroomAddress)
	if err != nil {
		return nil, err
	}

	// get user's ETH addresses
	userAddresses, err := s.userEthAddressGetter.GetETHAddresses(userID)
	if err != nil {
		glog.Errorf("error getting ETH addresses for user: %v", err)
		return nil, ErrorUnauthorized
	}

	// check if userID.eth_address is on the multisig for `input.ContractAddress` newsroom contract
	var isMember bool
	for _, member := range multisigMembers {
		for _, userAddress := range userAddresses {
			if member == userAddress {
				isMember = true
				break
			}
		}

		if isMember {
			break
		}
	}

	if !isMember {
		return nil, ErrorUnauthorized
	}

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
