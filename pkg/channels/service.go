package channels

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Jeffail/tunny"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/golang/glog"
	"github.com/gosimple/slug"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/email"
	"github.com/nfnt/resize"
	uuid "github.com/satori/go.uuid"
	"github.com/vincent-petithory/dataurl"
	"image"
	"image/jpeg"
	"image/png"
	"regexp"
	"runtime"
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
	imageProcessingPool  *tunny.Pool
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

	s := &Service{
		persister,
		newsroomHelper,
		stripeConnector,
		tokenGenerator,
		emailer,
		signupLoginProtoHost,
		nil,
	}
	multiplier := 1
	numCPUs := runtime.NumCPU() * multiplier
	s.imageProcessingPool = tunny.NewFunc(numCPUs, s.processAvatar)
	return s
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

// CreateNewsroomChannel creates a channel with type "newsroom"
func (s *Service) CreateNewsroomChannel(userID string, userAddresses []common.Address, input CreateNewsroomChannelInput) (*Channel, error) {

	channelType := TypeNewsroom
	reference := strings.ToLower(input.ContractAddress)

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

// CreateChannelMember creates a channel member for the channel
func (s *Service) CreateChannelMember(userID string, channelID string) (*ChannelMember, error) {
	channel, err := s.GetChannel(channelID)
	if err != nil {
		return nil, err
	}
	return s.persister.CreateChannelMember(channel, userID)
}

// DeleteChannelMember deletes a channel member for the channel
func (s *Service) DeleteChannelMember(userID string, channelID string) error {
	channel, err := s.GetChannel(channelID)
	if err != nil {
		return err
	}
	return s.persister.DeleteChannelMember(channel, userID)
}

// CreateGroupChannel creates a channel with type "group"
func (s *Service) CreateGroupChannel(userID string, handle string) (*Channel, error) {
	channelType := TypeGroup

	// groups don't reference anything, so generate a new one
	// TODO(dankins): should this reference a DID on an identity server?
	id := uuid.NewV4()
	reference := id.String()

	return s.persister.CreateChannel(CreateChannelInput{
		CreatorUserID: userID,
		ChannelType:   channelType,
		Reference:     reference,
		Handle:        &handle,
	})
}

func getImageAndDecodedDataURLFromDataURL(dataURL string) (*image.Image, *dataurl.DataURL, error) {
	decodedDataURL, err := dataurl.DecodeString(dataURL)
	if err != nil {
		return nil, nil, err
	}
	if decodedDataURL.Type != "image" {
		return nil, nil, ErrorBadAvatarDataURLType
	}
	if decodedDataURL.Subtype != "png" && decodedDataURL.Subtype != "jpg" {
		return nil, nil, ErrorBadAvatarDataURLSubType
	}
	if decodedDataURL.Encoding != "base64" {
		return nil, nil, ErrorBadAvatarEncoding
	}

	justData := strings.Split(dataURL, ",")[1] // TODO: should be able to get this from the decodedDataURL instead of splitting the string
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(justData))
	m, _, err := image.Decode(reader)
	if err != nil {
		return nil, nil, err
	}
	return &m, decodedDataURL, nil
}

// SetAvatarDataURL sets the avatar data url on a channel of any type
func (s *Service) SetAvatarDataURL(userID string, channelID string, avatarDataURL string) (*Channel, error) {
	image, decodedDataURL, err := getImageAndDecodedDataURLFromDataURL(avatarDataURL)
	if err != nil {
		return nil, err
	}
	if (*image).Bounds().Size().X != 336 || (*image).Bounds().Size().Y != 336 {
		return nil, ErrorBadAvatarSize
	}

	channel, err := s.persister.SetAvatarDataURL(userID, channelID, avatarDataURL)
	if err != nil {
		return nil, err
	}
	go func(p *tunny.Pool) {
		p.Process(processAvatarInputs{userID, channelID, image, decodedDataURL})
	}(s.imageProcessingPool)

	return channel, nil
}

type processAvatarInputs struct {
	userID         string
	channelID      string
	image          *image.Image
	decodedDataURL *dataurl.DataURL
}

func (s *Service) processAvatar(payload interface{}) interface{} {
	inputs := payload.(processAvatarInputs)

	mTiny := resize.Resize(72, 72, *(inputs.image), resize.Lanczos3)

	var buff bytes.Buffer
	if inputs.decodedDataURL.Subtype == "jpeg" {
		err := jpeg.Encode(&buff, mTiny, nil)
		if err != nil {
			return err
		}
	} else if inputs.decodedDataURL.Subtype == "png" {
		err := png.Encode(&buff, mTiny)
		if err != nil {
			return err
		}
	} else {
		return ErrorBadAvatarDataURLSubType
	}

	mTinyBase64Str := base64.StdEncoding.EncodeToString(buff.Bytes())
	mTinyDataURL := "data:" + inputs.decodedDataURL.ContentType() + ";base64," + mTinyBase64Str
	return s.persister.SetTiny72AvatarDataURL(inputs.userID, inputs.channelID, mTinyDataURL)
}

// SetStripeCustomerID sets the stripe customer id on a channel of any type
func (s *Service) SetStripeCustomerID(userID string, channelID string, stripeCustomerID string) (*Channel, error) {
	return s.persister.SetStripeCustomerID(userID, channelID, stripeCustomerID)
}

// ClearStripeCustomerID sets the stripe customer id on a channel of any type
func (s *Service) ClearStripeCustomerID(userID string, channelID string) (*Channel, error) {
	return s.persister.ClearStripeCustomerID(userID, channelID)
}

// SetHandle sets the handle on a channel of any type
func (s *Service) SetHandle(userID string, channelID string, handle string) (*Channel, error) {
	channel, err := s.persister.GetChannel(channelID)
	if err != nil {
		return nil, err
	}
	if channel.Handle != nil && *(channel.Handle) != "" {
		return nil, ErrorHandleAlreadySet
	}
	if !IsValidHandle(handle) {
		return nil, ErrorInvalidHandle
	}
	return s.persister.SetHandle(userID, channelID, handle)
}

// SetNewsroomHandleOnAccepted sets the handle on a newsroom channel
// should only be called by governance event handler
func (s *Service) SetNewsroomHandleOnAccepted(channelID string, newsroomName string) (*Channel, error) {
	channel, err := s.persister.GetChannel(channelID)
	if err != nil {
		return nil, err
	}
	if channel.Handle != nil && *(channel.Handle) != "" {
		return nil, ErrorHandleAlreadySet
	}
	handle := slug.Make(newsroomName)
	handleLength := len(handle)
	if handleLength > 24 {
		handle = string(handle[0:24])
	} else if handleLength < 4 {
		handle = handle + "-news"
	}
	handle = strings.Trim(handle, "-")

	log.Infof("CheckHandle: %s", handle)
	if !IsValidNewsroomHandle(handle) {
		return nil, ErrorInvalidHandle
	}
	return s.persister.SetNewsroomHandleOnAccepted(channelID, handle)
}

// ClearNewsroomHandleOnRemoved clears the handle on a newsroom channel
// should only be called by governance event handler
func (s *Service) ClearNewsroomHandleOnRemoved(channelID string) (*Channel, error) {
	return s.persister.ClearNewsroomHandleOnRemoved(channelID)
}

// don't export since should only be called through email confirm flow
func (s *Service) setEmailAddress(userID string, channelID string, emailAddress string) (*SetEmailResponse, error) {
	_, err := s.persister.GetChannel(channelID)
	if err != nil {
		return &SetEmailResponse{}, err
	}

	isAdmin, err := s.IsChannelAdmin(userID, channelID)
	if err != nil {
		return &SetEmailResponse{}, err
	}
	if !isAdmin {
		return &SetEmailResponse{}, ErrorUnauthorized
	}

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
	if !IsValidEmail(email) {
		return &SetEmailResponse{}, ErrorInvalidEmail
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

// ChannelEmailAddress returns the email address of the channel
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
	matched, err := regexp.Match(`^(\w){4,15}$`, []byte(handle))
	if err != nil {
		return false
	}

	return matched
}

// IsValidNewsroomHandle returns whether the provided newsroom handle is valid
func IsValidNewsroomHandle(handle string) bool {
	matched, err := regexp.Match(`^(\w){4,25}$`, []byte(handle))
	if err != nil {
		return false
	}

	return matched
}

// IsValidEmail returns whether the provided email is valid
func IsValidEmail(handle string) bool {
	matched, err := regexp.Match(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`, []byte(handle))
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
