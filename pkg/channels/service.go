package channels

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/email"
	uuid "github.com/satori/go.uuid"
	"regexp"
	"strings"
)

const (
	// number of seconds that a JWT token sent for set email confirmation is valid
	defaultJWTEmailExpiration = 60 * 60 * 6 // 6 hours

	// OkResponse is sent when an action is completed successfully
	OkResponse = "ok"

	subDelimiter             = "||"
	defaultSetEmailVerifyURI = "auth/confirm-email"
	confirmEmailTemplate     = "d-6fb32255ddc5461c86db126042e3ed78"
	civilMediaName           = "Civil Media Company"
	civilMediaEmail          = "support@civil.co"

	defaultAsmGroupID = 8328 // Civil Registry Alerts
)

// Service provides methods to interact with Channels
type Service struct {
	persister            Persister
	newsroomHelper       NewsroomHelper
	stripeConnector      StripeConnector
	tokenGenerator       *utils.JwtTokenGenerator
	emailer              *email.Emailer
	signupLoginProtoHost string
}

// NewsroomHelper describes methods needed to get the members of a newsroom multisig
type NewsroomHelper interface {
	GetMultisigMembers(newsroomAddress common.Address) ([]common.Address, error)
	GetOwner(newsroomAddress common.Address) (common.Address, error)
}

// StripeConnector defines the functions needed to connect an account to Stripe
type StripeConnector interface {
	ConnectAccount(code string) (string, error)
}

// NewServiceFromConfig creates a new channels.Service using the main graphql config
func NewServiceFromConfig(persister Persister, newsroomHelper NewsroomHelper, stripeConnector StripeConnector, tokenGenerator *utils.JwtTokenGenerator,
	emailer *email.Emailer, config *utils.GraphQLConfig) *Service {
	signupLoginProtoHost := config.SignupLoginProtoHost
	return NewService(persister, newsroomHelper, stripeConnector, tokenGenerator, emailer, signupLoginProtoHost)
}

// NewService builds a new Service instance
func NewService(persister Persister, newsroomHelper NewsroomHelper, stripeConnector StripeConnector, tokenGenerator *utils.JwtTokenGenerator,
	emailer *email.Emailer, signupLoginProtoHost string) *Service {

	return &Service{
		persister,
		newsroomHelper,
		stripeConnector,
		tokenGenerator,
		emailer,
		signupLoginProtoHost,
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

// SetHandle sets the handle on a channel of any type
func (s *Service) SetHandle(userID string, channelID string, handle string) (*Channel, error) {
	channel, err := s.persister.GetChannel(channelID)
	if err != nil {
		return nil, err
	}
	if channel.Handle != nil {
		return nil, ErrorHandleAlreadySet
	}
	if !IsValidHandle(handle) {
		return nil, ErrorInvalidHandle
	}
	return s.persister.SetHandle(userID, channelID, handle)
}

// don't export since should only be called through email confirm flow
func (s *Service) setEmailAddress(userID string, channelID string, emailAddress string) (*SetEmailResponse, error) {
	_, err := s.persister.GetChannel(channelID)
	if err != nil {
		return &SetEmailResponse{}, err
	}

	// check again that email is valid? ehh

	_, err = s.persister.SetEmailAddress(userID, channelID, emailAddress)
	if err != nil {
		return &SetEmailResponse{}, err
	}
	return &SetEmailResponse{UserID: userID, ChannelID: channelID}, nil
}

// SendEmailConfirmation sends an email to the user with link containing jwt that can be used to confirm ownership of email address
func (s *Service) SendEmailConfirmation(userID string, channelID string, emailAddress string, channelType SetEmailEnum) (*Channel, error) {
	channel, err := s.persister.GetChannel(channelID)
	if err != nil {
		return nil, err
	}
	if !IsValidEmail(emailAddress) {
		return nil, ErrorInvalidEmail
	}

	referral := string(channelType)
	_, err = s.sendEmailToken(emailAddress, userID, channelID, confirmEmailTemplate, defaultSetEmailVerifyURI, referral)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (s *Service) sendEmailToken(emailAddress string, userID string, channelID string, templateID string, verifyURI string,
	referral string) (string, error) {
	if s.emailer == nil {
		return "", fmt.Errorf("emailer is nil, disabling email of magic link")
	}
	if s.signupLoginProtoHost == "" {
		return "", fmt.Errorf("no signup/login host for confirmation email")
	}

	sub, err := s.buildSub(emailAddress, referral, userID, channelID)
	if err != nil {
		return "", err
	}
	emailToken, err := s.tokenGenerator.GenerateToken(
		sub,
		defaultJWTEmailExpiration,
	)
	if err != nil {
		return "", err
	}

	verifyLink := s.buildSetEmailConfirmLink(emailToken, verifyURI)
	verifyMarkup := s.buildSetEmailConfirmMarkup(verifyLink)

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

func (s *Service) buildSetEmailConfirmLink(emailToken string, verifyURI string) string {
	link := fmt.Sprintf("%v/%v?jwt=%v", s.signupLoginProtoHost, verifyURI, emailToken)
	return link
}

func (s *Service) buildSetEmailConfirmMarkup(confirmLink string) string {
	return fmt.Sprintf("<a clicktracking=off href=\"%v\">Confirm your 1 email address</a>", confirmLink)
}

func (s *Service) buildSub(emailAddress string, ref string, userID string, channelID string) (string, error) {
	if ref == "" || emailAddress == "" || userID == "" || channelID == "" {
		return "", errors.New("unable to build Sub")
	}

	parts := []string{emailAddress, ref, userID, channelID}
	return strings.Join(parts, subDelimiter), nil
}

// SetEmailConfirm validates the JWT token emailed to the user and creates the User account
func (s *Service) SetEmailConfirm(signupJWT string) (*SetEmailResponse, error) {
	claims, err := s.tokenGenerator.ValidateToken(signupJWT)
	if err != nil {
		return &SetEmailResponse{}, err
	}

	sub := claims["sub"].(string)
	email, _, userID, channelID := s.subData2(sub)
	if email == "" {
		return &SetEmailResponse{}, fmt.Errorf("no email found in token")
	}

	// Don't allow refresh token use here
	_, ok := claims["aud"].(string)
	if ok {
		return &SetEmailResponse{}, fmt.Errorf("invalid token")
	}

	return s.setEmailAddress(userID, channelID, email)
}

// ConnectStripeInput contains the fields needed to set the channel's stripe account
type ConnectStripeInput struct {
	ChannelID string
	OAuthCode string
}

// ConnectStripe connects the Stripe Account and sets the Stripe Account ID on a channel
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

// GetChannelMembers returns a list of channel members given a channel id
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

// ChannelHandle returns the email address of the channel
func (s *Service) ChannelEmailAddress(channelID string) (string, error) {
	channel, err := s.persister.GetChannel(channelID)
	if err != nil {
		return "", err
	}
	return channel.EmailAddress, nil
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

// IsValidEmail returns whether the provided email is valid
func IsValidEmail(handle string) bool {
	matched, err := regexp.Match(`\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`, []byte(handle))
	if err != nil {
		return false
	}

	return matched
}

func (s *Service) subData2(sub string) (email string, ref string, userID string, channelID string) {
	splitsub := strings.Split(sub, subDelimiter)
	if len(splitsub) == 1 {
		return splitsub[0], "", "", ""

	} else if len(splitsub) == 2 {
		return splitsub[0], splitsub[1], "", ""
	} else if len(splitsub) == 3 {
		return splitsub[0], splitsub[1], splitsub[2], ""
	} else if len(splitsub) == 4 {
		return splitsub[0], splitsub[1], splitsub[2], splitsub[3]
	}

	return "", "", "", ""
}
