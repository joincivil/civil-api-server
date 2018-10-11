package users

import (
	"github.com/joincivil/civil-api-server/pkg/auth"
	processormodel "github.com/joincivil/civil-events-processor/pkg/model"
)

// UserService is an implementation of UserService that uses Postgres for persistence
type UserService struct {
	userPersister UserPersister
}

// GetUser no comment
func (s *UserService) GetUser(identifier UserCriteria, createIfNotExists bool) (*User, error) {

	var user *User

	user, err := s.userPersister.User(&identifier)

	if err != nil && err == processormodel.ErrPersisterNoResults {
		user = &User{
			Email: identifier.Email,
		}
		newUserErr := s.userPersister.SaveUser(user)
		if newUserErr != nil {
			return nil, newUserErr
		}
	} else if err != nil {
		return nil, err
	}

	return user, nil
}

// SetEthAddressInput is used to update a user's ETH address
type SetEthAddressInput struct {
	Message     string `json:"message"`
	MessageHash string `json:"messageHash"`
	Signature   string `json:"signature"`
	Signer      string `json:"signer"`
	R           string `json:"r"`
	S           string `json:"s"`
	V           string `json:"v"`
}

// SetEthAddress verifies that the signature if valid and then updates their ETH address
func (s *UserService) SetEthAddress(identifier UserCriteria, request *SetEthAddressInput) (*User, error) {

	user, err := s.GetUser(identifier, false)
	if err != nil {
		return nil, err
	}

	err = auth.VerifyEthChallengeAndSignature(auth.ChallengeRequest{
		ExpectedPrefix: "I control this address",
		GracePeriod:    120,
		InputAddress:   request.Signer,
		InputChallenge: request.Message,
		Signature:      request.Signature,
	})
	if err != nil {
		return nil, err
	}

	user.EthAddress = request.Signer

	err = s.userPersister.UpdateUser(user, []string{"EthAddress"})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// NewUserService instantiates a new DefaultUserService
func NewUserService(userPersister UserPersister) *UserService {

	return &UserService{
		userPersister,
	}
}
