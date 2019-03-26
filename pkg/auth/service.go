package auth

import (
	"fmt"
	"strings"

	"github.com/joincivil/go-common/pkg/email"
	cpersist "github.com/joincivil/go-common/pkg/persistence"

	"github.com/joincivil/civil-api-server/pkg/users"
)

const (
	// number of seconds that a JWT token has until it expires
	// TODO(PN): when client is integrated with refresh, lower this value
	defaultJWTExpiration = 60 * 60 * 24 * 60 // 60 days
	// number of seconds that a JWT token sent for login or signup is valid
	defaultJWTEmailExpiration = 60 * 60 * 6 // 6 hours
	// number of seconds for a signature challenge
	defaultGracePeriod = 5 * 60 // 5 minutes
	// OkResponse is sent when an action is completed successfully
	OkResponse = "ok"
	// EmailNotFoundResponse is sent when an email address is not found for a user
	EmailNotFoundResponse = "emailnotfound"
	// EmailExistsResponse  is sent when an email address already exists for a user
	EmailExistsResponse = "emailexists"

	subDelimiter = "||"

	// sendgrid template ID for signup that sends a JWT encoded with the email address
	// defaultSignupEmailConfirmTemplateID = "d-88f731b52a524e6cafc308d0359b84a6"
	// sendgrid template ID for login that sends a JWT encoded with the email address
	// defaultLoginEmailConfirmTemplateID = "d-a228aa83fed8476b82d4c97288df20d5"

	defaultSignupVerifyURI  = "auth/signup/verify-token"
	defaultLoginVerifyURI   = "auth/login/verify-token"
	newsroomSignupVerifyURI = "apply-to-registry/signup"
	newsroomLoginVerifyURI  = "apply-to-registry/login"

	civilMediaName  = "Civil Media Company"
	civilMediaEmail = "support@civil.co"

	defaultAsmGroupID = 8328 // Civil Registry Alerts

)

// ApplicationEmailTemplateMap represents a mapping of the ApplicationEnum to it's email
// template ID
type ApplicationEmailTemplateMap map[ApplicationEnum]string

// Validate checks to ensure the items/values in the map are valid
func (a ApplicationEmailTemplateMap) Validate() error {
	for appEnum, templateID := range a {
		if !appEnum.IsValid() {
			return fmt.Errorf("app enum is not valid: %v", appEnum.String())
		}
		if templateID == "" {
			return fmt.Errorf("cannot have empty template ID for %v", appEnum.String())
		}
	}
	return nil
}

// FromStringMap converts a map[string]string to an ApplicationEmailTemplateMap
func (a ApplicationEmailTemplateMap) FromStringMap(smap map[string]string) error {
	for appName, templateID := range smap {
		enum := ApplicationEnum(appName)
		if !enum.IsValid() {
			return fmt.Errorf("app enum is not valid: %v", enum.String())
		}
		if templateID == "" {
			return fmt.Errorf("cannot have empty template ID for %v", enum.String())
		}
		a[enum] = templateID
	}
	return nil
}

// Service is used to create and login in Users
type Service struct {
	userService            *users.UserService
	tokenGenerator         *JwtTokenGenerator
	emailer                *email.Emailer
	signupEmailTemplateIDs ApplicationEmailTemplateMap
	loginEmailTemplateIDs  ApplicationEmailTemplateMap
	signupLoginProtoHost   string
	refreshBlacklist       []string
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userService *users.UserService, tokenGenerator *JwtTokenGenerator,
	emailer *email.Emailer, signupTemplateIDs map[string]string,
	loginTemplateIDs map[string]string, signupLoginProtoHost string,
	refreshBlacklist []string) (*Service, error) {
	var signupIDs ApplicationEmailTemplateMap
	if signupTemplateIDs != nil {
		signupIDs = ApplicationEmailTemplateMap{}
		err := signupIDs.FromStringMap(signupTemplateIDs)
		if err != nil {
			return nil, err
		}
	}
	var loginIDs ApplicationEmailTemplateMap
	if loginTemplateIDs != nil {
		loginIDs = ApplicationEmailTemplateMap{}
		err := loginIDs.FromStringMap(loginTemplateIDs)
		if err != nil {
			return nil, err
		}
	}
	return &Service{
		userService:            userService,
		tokenGenerator:         tokenGenerator,
		emailer:                emailer,
		signupEmailTemplateIDs: signupIDs,
		loginEmailTemplateIDs:  loginIDs,
		signupLoginProtoHost:   signupLoginProtoHost,
		refreshBlacklist:       refreshBlacklist,
	}, nil
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
// Returns the repsonse code, the token generated for the email, and a potential error
func (s *Service) SignupEmailSend(emailAddress string) (string, string, error) {
	return s.SignupEmailSendForApplication(emailAddress, ApplicationEnumDefault)
}

// SignupEmailSendForApplication sends an email for the given application to allow the user
// to confirm before creating the User
// Returns the repsonse code, the token generated for the email, and a potential error
func (s *Service) SignupEmailSendForApplication(emailAddress string,
	application ApplicationEnum) (string, string, error) {
	identifier := users.UserCriteria{
		Email: strings.ToLower(emailAddress),
	}
	user, err := s.userService.MaybeGetUser(identifier)
	if err != nil {
		return "", "", err
	}

	// If user does exist, return with code
	if user != nil {
		return EmailExistsResponse, "", nil
	}

	templateID, err := s.SignupTemplateIDForApplication(application)
	if err != nil {
		return "", "", err
	}

	verifyURI := defaultSignupVerifyURI
	if application == ApplicationEnumNewsroom {
		verifyURI = newsroomSignupVerifyURI
	}
	referral := string(application)
	token, err := s.sendEmailToken(emailAddress, templateID, verifyURI, referral)
	if err != nil {
		return "", "", err
	}
	return OkResponse, token, nil
}

// SignupEmailConfirm validates the JWT token emailed to the user and creates the User account
func (s *Service) SignupEmailConfirm(signupJWT string) (*LoginResponse, error) {
	claims, err := s.tokenGenerator.ValidateToken(signupJWT)
	if err != nil {
		return nil, err
	}

	sub := claims["sub"].(string)
	email, referral := s.subData(sub)
	if email == "" {
		return nil, fmt.Errorf("no email found in token")
	}

	// Don't allow refresh token use here
	_, ok := claims["aud"].(string)
	if ok {
		return nil, fmt.Errorf("invalid token")
	}

	identifier := users.UserCriteria{
		Email:       strings.ToLower(email),
		AppReferral: strings.ToLower(referral),
	}
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
		return nil, fmt.Errorf("user does not exist")
	} else if err != nil {
		return nil, err
	}

	return s.buildLoginResponse(user)
}

// LoginEmailSend sends an email to allow the user to confirm before creating the User
// Returns the repsonse code, the token generated for the email, and a potential error
func (s *Service) LoginEmailSend(emailAddress string) (string, string, error) {
	return s.LoginEmailSendForApplication(emailAddress, ApplicationEnumDefault)
}

// LoginEmailSendForApplication sends an email for the given application to allow
// the user to confirm before creating the User
// Returns the repsonse code, the token generated for the email, and a potential error
func (s *Service) LoginEmailSendForApplication(emailAddress string,
	application ApplicationEnum) (string, string, error) {
	identifier := users.UserCriteria{
		Email: strings.ToLower(emailAddress),
	}
	user, err := s.userService.MaybeGetUser(identifier)
	if err != nil {
		return "", "", err
	}

	// If user does not exist, return with code
	if user == nil {
		return EmailNotFoundResponse, "", nil
	}

	templateID, err := s.LoginTemplateIDForApplication(application)
	if err != nil {
		return "", "", err
	}

	verifyURI := defaultLoginVerifyURI
	if application == ApplicationEnumNewsroom {
		verifyURI = newsroomLoginVerifyURI
	}

	referral := string(application)

	token, err := s.sendEmailToken(emailAddress, templateID, verifyURI, referral)
	if err != nil {
		return "", "", err
	}
	return OkResponse, token, nil
}

// LoginEmailConfirm validates the JWT token emailed to the user and creates the User account
func (s *Service) LoginEmailConfirm(signupJWT string) (*LoginResponse, error) {
	claims, err := s.tokenGenerator.ValidateToken(signupJWT)
	if err != nil {
		return nil, err
	}

	sub := claims["sub"].(string)
	email, referral := s.subData(sub)
	if email == "" {
		return nil, fmt.Errorf("no email found in token")
	}

	// Don't allow refresh token use here
	_, ok := claims["aud"].(string)
	if ok {
		return nil, fmt.Errorf("invalid token")
	}

	identifier := users.UserCriteria{
		Email:       strings.ToLower(email),
		AppReferral: strings.ToLower(referral),
	}
	user, err := s.userService.MaybeGetUser(identifier)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("unrecognized user")
	}

	return s.buildLoginResponse(user)
}

// RefreshAccessToken will return a new JWT access token given the refresh token.
func (s *Service) RefreshAccessToken(refreshToken string) (*LoginResponse, error) {
	// Check if refresh token is on blacklist, this is in a env var for emergencies.  If we find the
	// blacklist volume to be high, can move to a store.
	for _, t := range s.refreshBlacklist {
		if strings.ToLower(t) == strings.ToLower(refreshToken) {
			return nil, fmt.Errorf("token blacklisted, rejecting: %v", refreshToken)
		}
	}

	// Validate refresh token
	claims, err := s.tokenGenerator.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Check claims
	uid, ok := claims["sub"].(string)
	if !ok || uid == "" {
		return nil, fmt.Errorf("no uid found in token")
	}
	aud, ok := claims["aud"].(string)
	if !ok || aud == "" {
		return nil, fmt.Errorf("invalid token")
	}
	if aud != "refresh" {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if user exists
	identifier := users.UserCriteria{
		UID: strings.ToLower(uid),
	}
	user, err := s.userService.MaybeGetUser(identifier)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("unrecognized user")
	}

	// Everything is verified, so generate a new token
	jwt, err := s.tokenGenerator.GenerateToken(user.UID, defaultJWTExpiration)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{UID: user.UID, Token: jwt, RefreshToken: refreshToken}, nil

}

// SignupTemplateIDForApplication returns the signup email template ID for the given application enum
func (s *Service) SignupTemplateIDForApplication(application ApplicationEnum) (string, error) {
	templateID, ok := s.signupEmailTemplateIDs[application]
	if !ok || templateID == "" {
		return "", fmt.Errorf("application signup %v template not found", application.String())
	}
	return templateID, nil
}

// LoginTemplateIDForApplication returns the login email template ID for the given application enum
func (s *Service) LoginTemplateIDForApplication(application ApplicationEnum) (string, error) {
	templateID, ok := s.loginEmailTemplateIDs[application]
	if !ok || templateID == "" {
		return "", fmt.Errorf("application login %v template not found", application.String())
	}
	return templateID, nil
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

func (s *Service) buildSignupLoginConfirmLink(emailToken string, verifyURI string) string {
	link := fmt.Sprintf("%v/%v?jwt=%v", s.signupLoginProtoHost, verifyURI, emailToken)
	return link
}

func (s *Service) buildSignupLoginConfirmMarkup(confirmLink string) string {
	return fmt.Sprintf("<a clicktracking=off href=\"%v\">Confirm your email address</a>", confirmLink)
}

func (s *Service) sendEmailToken(emailAddress string, templateID string, verifyURI string,
	referral string) (string, error) {
	if s.emailer == nil {
		return "", fmt.Errorf("emailer is nil, disabling email of magic link")
	}
	if s.signupLoginProtoHost == "" {
		return "", fmt.Errorf("no signup/login host for confirmation email")
	}

	emailToken, err := s.tokenGenerator.GenerateToken(
		s.buildSub(emailAddress, referral),
		defaultJWTEmailExpiration,
	)
	if err != nil {
		return "", err
	}

	verifyLink := s.buildSignupLoginConfirmLink(emailToken, verifyURI)
	verifyMarkup := s.buildSignupLoginConfirmMarkup(verifyLink)

	templateData := email.TemplateData{}
	templateData["host_proto"] = s.signupLoginProtoHost
	templateData["email_token"] = emailToken
	templateData["verify_link"] = verifyLink
	templateData["verify_markup"] = verifyMarkup

	emailReq := &email.SendTemplateEmailRequest{
		ToEmail:      emailAddress,
		FromName:     civilMediaName,
		FromEmail:    civilMediaEmail,
		TemplateID:   templateID,
		TemplateData: templateData,
		AsmGroupID:   defaultAsmGroupID,
	}
	return emailToken, s.emailer.SendTemplateEmail(emailReq)
}

func (s *Service) buildSub(email string, ref string) string {
	if ref == "" {
		return email
	}

	parts := []string{email, ref}
	return strings.Join(parts, subDelimiter)
}

func (s *Service) subData(sub string) (email string, ref string) {
	splitsub := strings.Split(sub, subDelimiter)
	if len(splitsub) == 1 {
		return splitsub[0], ""

	} else if len(splitsub) == 2 {
		return splitsub[0], splitsub[1]
	}

	return "", ""
}
