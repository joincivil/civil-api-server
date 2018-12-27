package auth

import (
	"fmt"

	"github.com/joincivil/go-common/pkg/email"
	cpersist "github.com/joincivil/go-common/pkg/persistence"

	"github.com/joincivil/civil-api-server/pkg/users"
)

const (
	// number of seconds that a JWT token has until it expires
	defaultJWTExpiration = 60 * 60 * 24
	// number of seconds that a JWT token sent for login or signup is valid
	defaultJWTEmailExpiration = 60 * 60 * 24 * 10
	// number of seconds for a signature challenge
	defaultGracePeriod = 15
	// sendgrid template ID for signup that sends a JWT encoded with the email address
	signupEmailConfirmTemplateID = "d-88f731b52a524e6cafc308d0359b84a6"
	// sendgrid template ID for login that sends a JWT encoded with the email address
	loginEmailConfirmTemplateID = "d-a228aa83fed8476b82d4c97288df20d5"
	// OkResponse is sent when an action is completed successfully
	OkResponse = "ok"
)

// Service is used to create and login in Users
type Service struct {
	userService    *users.UserService
	tokenGenerator *JwtTokenGenerator
	emailer        *email.Emailer
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userService *users.UserService, tokenGenerator *JwtTokenGenerator, emailer *email.Emailer) *Service {
	return &Service{
		userService,
		tokenGenerator,
		emailer,
	}
}

// SignupEth validates the Signature input then creates a User for that address
func (s *Service) SignupEth(input *users.SignatureInput) (*LoginResponse, error) {

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
	user, err := s.userService.CreateUser(identifier)
	if err != nil {
		return nil, err
	}

	return s.buildLoginResponse(user)
}

// SignupEmailSend sends an email to allow the user to confirm before creating the User
func (s *Service) SignupEmailSend(emailAddress string) (string, error) {
	err := s.sendEmailToken(emailAddress, signupEmailConfirmTemplateID)
	if err != nil {
		return "", err
	}
	return OkResponse, nil
}

// SignupEmailConfirm validates the JWT token emailed to the user and creates the User account
func (s *Service) SignupEmailConfirm(signupJWT string) (*LoginResponse, error) {
	claims, err := s.tokenGenerator.ValidateToken(signupJWT)
	if err != nil {
		return nil, err
	}
	email := claims["sub"].(string)

	identifier := users.UserCriteria{Email: email}
	user, err := s.userService.CreateUser(identifier)
	if err != nil {
		return nil, err
	}

	return s.buildLoginResponse(user)
}

// LoginEth creates a new user for the address in the Ethereum signature
func (s *Service) LoginEth(input *users.SignatureInput) (*LoginResponse, error) {
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
	user, err := s.userService.GetUser(identifier)
	if err != nil && err == cpersist.ErrPersisterNoResults {
		return nil, fmt.Errorf("User does not exist")
	} else if err != nil {
		return nil, err
	}

	return s.buildLoginResponse(user)
}

// LoginEmailSend sends an email to allow the user to confirm before creating the User
func (s *Service) LoginEmailSend(emailAddress string) (string, error) {
	err := s.sendEmailToken(emailAddress, loginEmailConfirmTemplateID)
	if err != nil {
		return "", err
	}
	return OkResponse, nil
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
	emailToken, err := s.tokenGenerator.GenerateToken(emailAddress, defaultJWTEmailExpiration)
	if err != nil {
		return err
	}

	templateData := email.TemplateData{}
	templateData["email_token"] = emailToken

	emailReq := &email.SendTemplateEmailRequest{
		ToEmail:      emailAddress,
		FromName:     "The Civil Media Company",
		FromEmail:    "support@civil.co",
		TemplateID:   templateID,
		TemplateData: templateData,
		AsmGroupID:   7395,
	}
	return s.emailer.SendTemplateEmail(emailReq)
}
