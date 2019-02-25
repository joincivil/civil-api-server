package users

import (
	"errors"

	"github.com/joincivil/civil-api-server/pkg/tokencontroller"

	"github.com/ethereum/go-ethereum/common"
	cpersist "github.com/joincivil/go-common/pkg/persistence"
	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
)

var (
	// ErrUserExists is an error that represents an attempt to create a new user
	// with a duplicate Email or ETH address
	ErrUserExists = errors.New("User already exists with this identifier")
	// ErrInvalidState is returned when a trying to
	ErrInvalidState = errors.New("User already exists with this identifier")
)

// UserService is an implementation of UserService that uses Postgres for persistence
type UserService struct {
	userPersister     UserPersister
	controllerUpdater TokenControllerUpdater
}

// NewUserService instantiates a new DefaultUserService
func NewUserService(userPersister UserPersister, controllerUpdater TokenControllerUpdater) *UserService {
	return &UserService{
		userPersister,
		controllerUpdater,
	}
}

// GetUser retrieves a user from the database
func (s *UserService) GetUser(identifier UserCriteria) (*User, error) {

	var user *User
	user, err := s.userPersister.User(&identifier)

	if err != nil {
		return nil, err
	}
	return user, nil
}

// MaybeGetUser retrieves a user from the database or returns nil if no results
func (s *UserService) MaybeGetUser(identifier UserCriteria) (*User, error) {

	user, err := s.GetUser(identifier)
	if err == cpersist.ErrPersisterNoResults {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateUser creates and persists a new User model
func (s *UserService) CreateUser(identifier UserCriteria) (*User, error) {

	existingUser, err := s.MaybeGetUser(identifier)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrUserExists
	}

	user := &User{
		Email:      identifier.Email,
		EthAddress: identifier.EthAddress,
	}
	newUserErr := s.userPersister.SaveUser(user)
	if newUserErr != nil {
		return nil, newUserErr
	}

	return user, nil
}

// UserUpdateInput describes the input needed for UpdateUser
type UserUpdateInput struct {
	OnfidoApplicantID   string                 `json:"onfidoApplicantID"`
	OnfidoCheckID       string                 `json:"onfidoCheckID"`
	KycStatus           string                 `json:"kycStatus"`
	QuizPayload         cpostgres.JsonbPayload `json:"quizPayload"`
	QuizStatus          string                 `json:"quizStatus"`
	PurchaseTxHashesStr string                 `json:"purchaseTxHashesStr"`
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(uid string, input *UserUpdateInput) (*User, error) {
	user, err := s.GetUser(UserCriteria{UID: uid})
	if err != nil {
		return nil, err
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
	if input.KycStatus != "" {
		user.KycStatus = input.KycStatus
		fields = append(fields, "KycStatus")
	}
	if input.PurchaseTxHashesStr != "" {
		user.PurchaseTxHashesStr = input.PurchaseTxHashesStr
		fields = append(fields, "PurchaseTxHashesStr")
	}

	if input.QuizStatus != "" {
		// if the QuizStatus changes to complete we need to add the user to the Civilian Whitelist on the token controller
		if input.QuizStatus == "complete" && user.QuizStatus != "complete" {
			if user.EthAddress == "" {
				return nil, ErrInvalidState
			}
			addr := common.HexToAddress(user.EthAddress)
			txHash, err := s.controllerUpdater.AddToCivilians(addr)
			if err != nil && err != tokencontroller.ErrAlreadyOnList {
				return nil, err
			}
			user.CivilianWhitelistTxID = txHash.String()
			fields = append(fields, "CivilianWhitelistTxID")
		}
		user.QuizStatus = input.QuizStatus
		fields = append(fields, "QuizStatus")
	}

	err = s.userPersister.UpdateUser(user, fields)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// SignatureInput is used to update a user's ETH address
type SignatureInput struct {
	Message     string `json:"message"`
	MessageHash string `json:"messageHash"`
	Signature   string `json:"signature"`
	Signer      string `json:"signer"`
	R           string `json:"r"`
	S           string `json:"s"`
	V           string `json:"v"`
}

// SetEthAddress verifies that the signature if valid and then updates their ETH address
func (s *UserService) SetEthAddress(identifier UserCriteria, address string) (*User, error) {

	user, err := s.GetUser(identifier)
	if err != nil {
		return nil, err
	}

	user.EthAddress = address

	err = s.userPersister.UpdateUser(user, []string{"EthAddress"})
	if err != nil {
		return nil, err
	}
	return user, nil
}
