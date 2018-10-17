package users

import (
	"errors"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-events-crawler/pkg/persistence/postgres"
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

// UserUpdateInput describes the input needed for UpdateUser
type UserUpdateInput struct {
	OnfidoApplicantID string                `json:"onfidoApplicantID"`
	OnfidoCheckID     string                `json:"onfidoCheckID"`
	KycStatus         string                `json:"kycStatus"`
	QuizPayload       postgres.JsonbPayload `json:"quizPayload"`
	QuizStatus        string                `json:"quizStatus"`
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(requestor *auth.Token, uid string, input *UserUpdateInput) (*User, error) {

	user, err := s.GetUser(UserCriteria{UID: uid}, false)
	if err != nil {
		return nil, err
	}

	if user.Email != requestor.Sub {
		return nil, errors.New("user is not authorized")
	}

	// TODO(dankins): inspecting each attribute feels dirty, can we do this via reflection or something?
	fields := []string{}
	if input.QuizPayload != nil {
		user.QuizPayload = input.QuizPayload
		fields = append(fields, "QuizPayload")
	}
	if input.OnfidoApplicantID != "" {
		user.OnfidoApplicantID = input.OnfidoApplicantID
		fields = append(fields, "OnfidoApplicantID")
	}
	if input.OnfidoCheckID != "" {
		user.OnfidoCheckID = input.OnfidoCheckID
		fields = append(fields, "OnfidoCheckID")
	}
	if input.QuizStatus != "" {
		user.QuizStatus = input.QuizStatus
		fields = append(fields, "QuizStatus")
	}
	if input.KycStatus != "" {
		user.KycStatus = input.KycStatus
		fields = append(fields, "KycStatus")
	}

	err = s.userPersister.UpdateUser(user, fields)
	if err != nil {
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
