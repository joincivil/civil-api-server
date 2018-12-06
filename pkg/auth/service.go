package auth

import (
	"fmt"

	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/civil-api-server/pkg/users"
	processormodel "github.com/joincivil/civil-events-processor/pkg/model"
)

const (
	// number of seconds that a JWT token has until it expires
	defaultJWTExpiration = 60 * 60 * 24
	// number of seconds for a signature challenge
	defaultGracePeriod = 15
	//
	signupEmailConfirmTemplateID = "d-88f731b52a524e6cafc308d0359b84a6"
	loginEmailConfirmTemplateID  = "d-a228aa83fed8476b82d4c97288df20d5"
)

// Service is used to create and login in Users
type Service struct {
	userService    *users.UserService
	tokenGenerator *JwtTokenGenerator
	emailer        *utils.Emailer
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userService *users.UserService, tokenGenerator *JwtTokenGenerator, emailer *utils.Emailer) *Service {
	return &Service{
		userService,
		tokenGenerator,
		emailer,
	}
}

// SignupEth validates the Signature input then creates a User for that address
func (s *Service) SignupEth(input *users.SetEthAddressInput) (*LoginResponse, error) {

	err := VerifyEthChallengeAndSignature(ChallengeRequest{
		ExpectedPrefix: "Sign up with Civil",
		GracePeriod:    defaultGracePeriod,
		InputAddress:   input.Signer,
		InputChallenge: input.Message,
		Signature:      input.Signature,
	})
	if err != nil {
		return nil, err
	}

	identifier := users.UserCriteria{EthAddress: input.Signer}
	existingUser, err := s.userService.MaybeGetUser(identifier)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, fmt.Errorf("User already exists with this address")
	}

	user, err := s.userService.CreateUser(identifier)
	if err != nil {
		return nil, err
	}

	return s.buildLoginResponse(user)
}

// SignupEmail sends an email to allow the user to confirm before creating the User
func (s *Service) SignupEmail(emailAddress string) (string, error) {
	err := s.sendEmailToken(emailAddress, signupEmailConfirmTemplateID)
	if err != nil {
		return "", err
	}
	return "ok", nil
}

// SignupEmailConfirm validates the JWT token emailed to the user and creates the User account
func (s *Service) SignupEmailConfirm(signupJWT string) (*LoginResponse, error) {
	claims, err := s.tokenGenerator.ValidateToken(signupJWT)
	if err != nil {
		return nil, err
	}
	email := claims["sub"].(string)

	identifier := users.UserCriteria{Email: email}
	existingUser, err := s.userService.MaybeGetUser(identifier)

	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, fmt.Errorf("User already exists with this email")
	}

	user, err := s.userService.CreateUser(identifier)
	if err != nil {
		return nil, err
	}

	return s.buildLoginResponse(user)
}

// LoginEth creates a new user for the address in the Ethereum signature
func (s *Service) LoginEth(input *users.SetEthAddressInput) (*LoginResponse, error) {
	err := VerifyEthChallengeAndSignature(ChallengeRequest{
		ExpectedPrefix: "Log in to Civil",
		GracePeriod:    defaultGracePeriod,
		InputAddress:   input.Signer,
		InputChallenge: input.Message,
		Signature:      input.Signature,
	})
	if err != nil {
		return nil, err
	}

	identifier := users.UserCriteria{EthAddress: input.Signer}
	user, err := s.userService.GetUser(identifier, false)
	if err != nil && err == processormodel.ErrPersisterNoResults {
		return nil, fmt.Errorf("User does not exist")
	} else if err != nil {
		return nil, err
	}

	return s.buildLoginResponse(user)
}

// LoginEmail sends an email to allow the user to confirm before creating the User
func (s *Service) LoginEmail(emailAddress string) (string, error) {
	err := s.sendEmailToken(emailAddress, loginEmailConfirmTemplateID)
	if err != nil {
		return "", err
	}
	return "ok", nil
}

// LoginEmailConfirm validates the JWT token emailed to the user and creates the User account
func (s *Service) LoginEmailConfirm(signupJWT string) (*LoginResponse, error) {
	claims, err := s.tokenGenerator.ValidateToken(signupJWT)
	if err != nil {
		return nil, err
	}
	email := claims["sub"].(string)

	identifier := users.UserCriteria{Email: email}
	user, err := s.userService.MaybeGetUser(identifier)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("unrecognized user")
	}

	return s.buildLoginResponse(user)
}

func (s *Service) buildLoginResponse(user *users.User) (*LoginResponse, error) {
	jwt, err := s.tokenGenerator.GenerateToken(user.UID, defaultJWTExpiration)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.tokenGenerator.GenerateRefreshToken(user.UID)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{UID: user.UID, Token: jwt, RefreshToken: refreshToken}, nil
}

func (s *Service) sendEmailToken(emailAddress string, templateID string) error {
	emailToken, err := s.tokenGenerator.GenerateToken(emailAddress, 60*60*24*10)
	if err != nil {
		return err
	}

	templateData := utils.TemplateData{}
	templateData["email_token"] = emailToken

	emailReq := &utils.SendTemplateEmailRequest{
		ToEmail:      emailAddress,
		FromName:     "The Civil Media Company",
		FromEmail:    "support@civil.co",
		TemplateID:   templateID,
		TemplateData: templateData,
		AsmGroupID:   7395,
	}
	return s.emailer.SendTemplateEmail(emailReq)
}
