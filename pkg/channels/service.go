package channels

import (
	"errors"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/golang/glog"
	uuid "github.com/satori/go.uuid"
)

// Service provides methods to interact with Channels
type Service struct {
	persister       Persister
	newsroomHelper  NewsroomHelper
	stripeConnector StripeConnector
}

// NewsroomHelper describes methods needed to get the members of a newsroom multisig
type NewsroomHelper interface {
	GetMultisigMembers(newsroomAddress common.Address) ([]common.Address, error)
	GetOwner(newsroomAddress common.Address) (common.Address, error)
}

// StripeCharger defines the functions needed to connect an account to Stripe
type StripeConnector interface {
	ConnectAccount(code string) (string, error)
}

// NewService builds a new Service instance
func NewService(persister Persister, newsroomHelper NewsroomHelper, stripeConnector StripeConnector) *Service {

	return &Service{
		persister,
		newsroomHelper,
		stripeConnector,
	}
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
func (s *Service) CreateNewsroomChannel(userID string, userAddresses []common.Address, input CreateNewsroomChannelInput) (*Channel, error) {

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
	multisigMembers, err := s.newsroomHelper.GetMultisigMembers(newsroomAddress)
	if err != nil {
		return nil, err
	}

	// check if userID.eth_address is on the multisig for `input.ContractAddress` newsroom contract
	var isMember bool
Loop:
	for _, member := range multisigMembers {
		for _, userAddress := range userAddresses {
			if member == userAddress {
				isMember = true
				break Loop
			}
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

// SetHandle comment
func (s *Service) SetHandle(userID string, channelID string, handle string) (*Channel, error) {
	if !IsValidHandle(handle) {
		return nil, ErrorInvalidHandle
	}
	return s.persister.SetHandle(userID, channelID, handle)
}

// ConnectStripeInput contains the fields needed to set the channel's stripe account
type ConnectStripeInput struct {
	ChannelID string
	OAuthCode string
}

// ConnectStripe comment
func (s *Service) ConnectStripe(userID string, input ConnectStripeInput) (*Channel, error) {
	if input.OAuthCode == "" {
		return nil, ErrorsInvalidInput
	}

	acct, err := s.stripeConnector.ConnectAccount(input.OAuthCode)
	if err != nil {
		log.Errorf("error connecting stripe account: %v", err)
		return nil, ErrorStripeIssue
	}

	return s.persister.SetStripeAccountID(userID, input.ChannelID, acct)

}

// GetStripePaymentAccount returns the stripe account associated with the channel
func (s *Service) GetStripePaymentAccount(channelID string) (string, error) {
	ch, err := s.persister.GetChannel(channelID)
	if err != nil {
		return "", err
	}

	return ch.StripeAccountID, nil
}

// GetEthereumPaymentAddress returns the Ethereum account associated with the channel
func (s *Service) GetEthereumPaymentAddress(channelID string) (common.Address, error) {

	ch, err := s.persister.GetChannel(channelID)
	if err != nil {
		return common.Address{}, err
	}

	if ch.ChannelType != "newsroom" {
		return common.Address{}, errors.New("GetEthereumPaymentAddress only supports channels with type `newsroom`")
	}

	return s.newsroomHelper.GetOwner(common.HexToAddress(ch.Reference))
}

// GetChannel gets a channel by ID
func (s *Service) GetChannel(id string) (*Channel, error) {
	return s.persister.GetChannel(id)
}

func (s *Service) GetChannelMembers(channelID string) ([]*ChannelMember, error) {
	return s.persister.GetChannelMembers(channelID)
}

// GetChannelByReference retrieves a channel by the reference field
func (s *Service) GetChannelByReference(channelType string, reference string) (*Channel, error) {
	return s.persister.GetChannelByReference(channelType, reference)
}

// GetChannelByHandle retrieves a channel by the handle
func (s *Service) GetChannelByHandle(handle string) (*Channel, error) {
	return s.persister.GetChannelByHandle(handle)
}

// IsChannelAdmin returns if the user is an admin of the channel
func (s *Service) IsChannelAdmin(userID string, channelID string) (bool, error) {
	return s.persister.IsChannelAdmin(userID, channelID)
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
