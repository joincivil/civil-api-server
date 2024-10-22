package nrsignup

import (
	"encoding/json"
	"fmt"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"go.uber.org/fx"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/joincivil/civil-api-server/pkg/jsonstore"

	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/go-common/pkg/generated/contract"
	"github.com/joincivil/go-common/pkg/newsroom"

	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/go-common/pkg/email"
)

const (
	// ApprovedSubValue is the value embedded in the token sub value to indicate
	// approval
	ApprovedSubValue = "app"

	// RejectedSubValue is the value embedded in the token sub value to indicate
	// rejection
	RejectedSubValue = "rej"

	// DefaultJsonbID is the default ID value for user data in the JSONb store.
	DefaultJsonbID = "nrsignup"
)

const (
	approvalLinkExpirySecs = 60 * 60 * 24 * 14 // 14 days
	rejectLinkExpirySecs   = 60 * 60 * 24 * 14 // 14 days

	signupWelcomeEmailTemplateID       = "d-125a9b151d99483f9dc8a8cc61fcb786"
	requestGrantUserEmailTemplateID    = "d-f17d8ba462ce4ac9ab24e447d9ee099d"
	requestGrantCouncilEmailTemplateID = "d-2ffc71848ea743b0a5b56a7c0d6b9ac3"
	appliedToRegistryEmailTemplateID   = "d-5a449104d603419195084adb1a536b9d"

	// grantApprovalUserEmailTemplateID = "d-f363c4aa8d404bd39e7c14f527318d4f"
	// grantApprovalCouncilEmailTemplateID = ""

	foundationEmailName    = "Civil Foundation"
	foundationEmailAddress = "foundation@civilfound.org"
	supportEmailAddress    = "support@civil.co"
	civilPipedriveEmail    = "grants@joincivil.com"

	defaultFromEmailName    = foundationEmailName
	defaultFromEmailAddress = foundationEmailAddress

	defaultAsmGroupID = 8328 // Civil Registry Alerts

	grantApprovalURI = "v1/nrsignup/grantapprove"

	applyStageLenName        = "applyStageLen"
	defaultApplyStageLenSecs = 1209600
)

const (
	// newsroom signup steps from Civil/newsroom-signup/Newsroom.tsx
	stepApplyComplete = 13
)

// NewNewsroomSignupService is a convenience function to initialize a new newsroom
// signup service struct
func NewNewsroomSignupService(helper *eth.Helper, emailer *email.Emailer, userService *users.UserService,
	jsonbService *jsonstore.Service, tokenGenerator *utils.JwtTokenGenerator,
	grantLandingProtoHost string, paramAddr string, registryListID string) (*Service, error) {
	return NewService(ServiceDeps{
		EthHelper:      helper,
		Emailer:        emailer,
		UserService:    userService,
		JsonbService:   jsonbService,
		TokenGenerator: tokenGenerator,
		Config: ServiceConfig{
			grantLandingProtoHost, paramAddr, registryListID,
		},
	}), nil
}

// ServiceConfig contains the configuration fields for ServiceDeps
type ServiceConfig struct {
	GrantLandingProtoHost string
	ParamAddr             string
	RegistryListID        string
}

// ServiceDeps contains the fields needed to build a new Service
type ServiceDeps struct {
	fx.In
	EthHelper      *eth.Helper
	Emailer        *email.Emailer
	UserService    *users.UserService
	JsonbService   *jsonstore.Service
	ChannelService *channels.Service
	TokenGenerator *utils.JwtTokenGenerator
	Config         ServiceConfig
}

// NewService constructs a new Service instance
func NewService(deps ServiceDeps) *Service {
	return &Service{
		client:                deps.EthHelper.Blockchain,
		emailer:               deps.Emailer,
		userService:           deps.UserService,
		jsonbService:          deps.JsonbService,
		tokenGenerator:        deps.TokenGenerator,
		grantLandingProtoHost: deps.Config.GrantLandingProtoHost,
		parameterizerAddr:     deps.Config.ParamAddr,
		registryListID:        deps.Config.RegistryListID,
		channelService:        deps.ChannelService,
	}
}

// Service is a struct and methods used for handling newsroom signup functionality
type Service struct {
	client                bind.ContractBackend
	emailer               *email.Emailer
	userService           *users.UserService
	jsonbService          *jsonstore.Service
	tokenGenerator        *utils.JwtTokenGenerator
	channelService        *channels.Service
	grantLandingProtoHost string
	alterMutex            sync.Mutex
	parameterizerAddr     string
	registryListID        string
}

// SendWelcomeEmail sends a newsroom signup welcome email to the given newsroom owner.
func (s *Service) SendWelcomeEmail(newsroomOwnerUID string) error {
	ch, err := s.channelService.GetChannelByReference(channels.TypeUser, newsroomOwnerUID)
	if err != nil {
		return err
	}
	if ch == nil {
		return fmt.Errorf("no user channel found: uid: %v", newsroomOwnerUID)
	}

	if ch.EmailAddress == "" {
		return fmt.Errorf("no email set for user: uid: %v", newsroomOwnerUID)
	}

	tmplData := email.TemplateData{
		"name": ch,
	}
	tmplReq := &email.SendTemplateEmailRequest{
		ToName:       ch.EmailAddress,
		ToEmail:      ch.EmailAddress,
		FromName:     defaultFromEmailName,
		FromEmail:    defaultFromEmailAddress,
		TemplateID:   signupWelcomeEmailTemplateID,
		TemplateData: tmplData,
		AsmGroupID:   defaultAsmGroupID,
	}

	return s.emailer.SendTemplateEmail(tmplReq)
}

// UpdateCharter takes a user id and a charter object and updates the SignupUserJSONData
// for that user with that charter. it creates it if it doesnt exist
func (s *Service) UpdateCharter(newsroomOwnerUID string, charter newsroom.Charter) error {
	charterUpdateFn := func(d *SignupUserJSONData) (*SignupUserJSONData, error) {
		d.Charter = &charter
		return d, nil
	}

	err := s.alterUserDataInJSONStore(newsroomOwnerUID, charterUpdateFn)

	if err != nil {
		newSignupData := SignupUserJSONData{Charter: &charter}
		return s.saveUserJSONData(newsroomOwnerUID, &newSignupData)
	}

	return err
}

// RequestGrant sends a request for a grant to the Foundation on behalf of a newsroom via a
// newsroom owner.
// Sends along the data in the newsroom charter for review.
// Will send emails to the Foundation and newsroom owner.
// The Foundation email will have magic links to approve/reject the grant.
func (s *Service) RequestGrant(newsroomOwnerUID string, requested bool) error {
	user, err := s.userService.MaybeGetUser(users.UserCriteria{
		UID: newsroomOwnerUID,
	})
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("no user found: uid: %v", newsroomOwnerUID)
	}

	ch, err := s.channelService.GetChannelByReference(channels.TypeUser, user.UID)
	if err != nil {
		return err
	}
	if ch == nil {
		return fmt.Errorf("no user found: uid: %v", newsroomOwnerUID)
	}

	userEmail := ch.EmailAddress

	// Set the grant requested flag to true for this user UID
	err = s.setGrantRequestedFlag(newsroomOwnerUID, requested)
	if err != nil || !requested {
		return err
	}

	signupData, err := s.RetrieveUserJSONData(newsroomOwnerUID)
	if err != nil {
		return err
	}

	// Email to council with charter info via pipedrive
	tmplData := email.TemplateData{}
	s.buildCharterDataIntoTemplate(tmplData, signupData)
	_, err = s.buildGrantDecisionLinksIntoTemplate(tmplData, newsroomOwnerUID)
	if err != nil {
		return err
	}
	tmplData["nr_applicant_email"] = userEmail
	tmplData["nr_applicant_address"] = user.EthAddress

	tmplReq := &email.SendTemplateEmailRequest{
		ToName:       civilPipedriveEmail,
		ToEmail:      civilPipedriveEmail,
		FromName:     supportEmailAddress,
		FromEmail:    supportEmailAddress,
		TemplateID:   requestGrantCouncilEmailTemplateID,
		TemplateData: tmplData,
		AsmGroupID:   defaultAsmGroupID,
	}
	err1 := s.emailer.SendTemplateEmail(tmplReq)

	// Email to newsroom owner to tell them to wait for a response
	tmplData = email.TemplateData{
		"name": userEmail,
	}
	tmplReq = &email.SendTemplateEmailRequest{
		ToName:       userEmail,
		ToEmail:      userEmail,
		FromName:     defaultFromEmailName,
		FromEmail:    defaultFromEmailAddress,
		TemplateID:   requestGrantUserEmailTemplateID,
		TemplateData: tmplData,
		AsmGroupID:   defaultAsmGroupID,
	}
	if userEmail != "" {
		err2 := s.emailer.SendTemplateEmail(tmplReq)
		if err2 != nil {
			return fmt.Errorf("failed to send grant request user email: err: %v", err2)
		}
	}

	if err1 != nil {
		return fmt.Errorf("failed to send grant request foundation email: err: %v", err1)
	}

	return nil
}

// ApproveGrant approves or rejects a grant on behalf of the Foundation.
// Will send emails to the Foundation and newsroom owner.
func (s *Service) ApproveGrant(newsroomOwnerUID string, approved bool) error {
	user, err := s.userService.MaybeGetUser(users.UserCriteria{
		UID: newsroomOwnerUID,
	})
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("no user found: uid: %v", newsroomOwnerUID)
	}

	// NOTE(PN): Email to newsroom owner will be sent by the foundation via pipedrive.

	// Set the grant requested flag to true for this user UID
	return s.setGrantApprovedFlag(newsroomOwnerUID, approved)
}

// UpdateUserSteps updates the user steps through the newsroom flow.  This differs
// from userUpdate is that emails or actions will be triggered based on the user furthest
// step.
func (s *Service) UpdateUserSteps(newsroomOwnerUID string, step *int,
	furthestStep *int, lastSeen *int) error {
	user, err := s.userService.MaybeGetUser(users.UserCriteria{
		UID: newsroomOwnerUID,
	})
	if err != nil {
		return err
	}

	input := &users.UserUpdateInput{
		NrStep:         step,
		NrFurthestStep: furthestStep,
		NrLastSeen:     lastSeen,
	}
	_, err = s.userService.UpdateUser(newsroomOwnerUID, input)
	if err != nil {
		return err
	}

	// If furthest step hasn't changed or is nil, don't do anything
	if furthestStep == nil || (user != nil && user.NewsroomFurthestStep >= *furthestStep) {
		return err
	}

	if *furthestStep == stepApplyComplete {

		ch, err := s.channelService.GetChannelByReference(channels.TypeUser, user.UID)
		if err != nil {
			return err
		}
		if ch == nil {
			return fmt.Errorf("no user found: uid: %v", newsroomOwnerUID)
		}

		if ch.EmailAddress != "" {
			// Send user the application was complete email
			err = s.sendApplicationCompleteEmail(newsroomOwnerUID, ch.EmailAddress)
			if err != nil {
				return err
			}

			// Add the user to the registry alerts list if ID is set
			err = s.addToRegistryAlertsList(ch.EmailAddress)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

// SaveNewsroomDeployTxHash saves the txhash for a newsroom deploy
func (s *Service) SaveNewsroomDeployTxHash(newsroomOwnerUID string, txHash string) error {
	newsroomDeployTxHashUpdateFn := func(d *SignupUserJSONData) (*SignupUserJSONData, error) {
		d.NewsroomDeployTx = txHash
		return d, nil
	}

	return s.alterUserDataInJSONStore(newsroomOwnerUID, newsroomDeployTxHashUpdateFn)
}

// SaveNewsroomAddress saves the newsrooms address
func (s *Service) SaveNewsroomAddress(newsroomOwnerUID string, address string) error {
	newsroomAdressUpdateFn := func(d *SignupUserJSONData) (*SignupUserJSONData, error) {
		d.NewsroomAddress = address
		return d, nil
	}

	return s.alterUserDataInJSONStore(newsroomOwnerUID, newsroomAdressUpdateFn)
}

// SaveNewsroomApplyTxHash saves the txHash for the newsrooms application transaction
func (s *Service) SaveNewsroomApplyTxHash(newsroomOwnerUID string, txhash string) error {
	newsroomApplyTxHashUpdateFn := func(d *SignupUserJSONData) (*SignupUserJSONData, error) {
		d.TcrApplyTx = txhash
		return d, nil
	}

	return s.alterUserDataInJSONStore(newsroomOwnerUID, newsroomApplyTxHashUpdateFn)
}

// StartPollNewsroomDeployTx starts a polling process that detects when a newsroom
// contract deployment transaction has completed.  Will send emails
// alerting the Foundation and the newsroom owner that it has completed.
func (s *Service) StartPollNewsroomDeployTx(newsroomOwnerUID string) error {
	return nil
}

// StartPollApplicationTx starts a polling process that detects when an application
// transaction has completed.  Will send emails alerting the Foundation and the
// newsroom owner that it has completed.
func (s *Service) StartPollApplicationTx(newsroomOwnerUID string) error {
	return nil
}

// DeleteNewsroomData deletes the nrsignup data for the user
func (s *Service) DeleteNewsroomData(newsroomOwnerUID string) error {
	s.alterMutex.Lock()
	defer s.alterMutex.Unlock()

	// Set both the namespace and ID as the newsroom owner ID
	jsonbs, err := s.jsonbService.RetrieveJSONb(
		DefaultJsonbID,
		jsonstore.DefaultJsonbGraphqlNs,
		newsroomOwnerUID,
	)
	if err != nil {
		return err
	}
	if len(jsonbs) != 1 {
		return fmt.Errorf("retrieved more than 1 result from the JSONb store")
	}

	return s.jsonbService.DeleteJSONb(jsonbs[0])
}

func (s *Service) buildGrantDecisionLink(newsroomOwnerUID string, approved bool) (string, error) {
	var sub string
	var expiry int
	if approved {
		sub = fmt.Sprintf("%v:%v", newsroomOwnerUID, ApprovedSubValue)
		expiry = approvalLinkExpirySecs
	} else {
		sub = fmt.Sprintf("%v:%v", newsroomOwnerUID, RejectedSubValue)
		expiry = rejectLinkExpirySecs
	}

	token, err := s.tokenGenerator.GenerateToken(sub, expiry)
	if err != nil {
		return "", err
	}

	link := fmt.Sprintf("%v/%v/%v", s.grantLandingProtoHost, grantApprovalURI, token)
	return link, nil
}

func (s *Service) buildGrantDecisionLinksIntoTemplate(tmplData email.TemplateData,
	newsroomOwnerUID string) (email.TemplateData, error) {
	approveLink, err := s.buildGrantDecisionLink(newsroomOwnerUID, true)
	if err != nil {
		return tmplData, err
	}
	rejectLink, err := s.buildGrantDecisionLink(newsroomOwnerUID, false)
	if err != nil {
		return tmplData, err
	}
	tmplData["nr_grant_approve_link"] = approveLink
	tmplData["nr_grant_approve_markup"] = fmt.Sprintf(
		"<a href=\"%v\">Approve</a>",
		approveLink,
	)
	tmplData["nr_grant_reject_link"] = rejectLink
	tmplData["nr_grant_reject_markup"] = fmt.Sprintf(
		"<a href=\"%v\">Reject</a>",
		rejectLink,
	)
	return tmplData, nil
}

func (s *Service) buildCharterDataIntoTemplate(tmplData email.TemplateData,
	signupData *SignupUserJSONData) email.TemplateData {
	newsroomCharter := signupData.Charter

	tmplData["nr_name"] = newsroomCharter.Name
	tmplData["nr_logo_url"] = newsroomCharter.LogoURL
	tmplData["nr_logo_markup"] = s.buildLogoURLMarkup(newsroomCharter.LogoURL)
	tmplData["nr_url"] = newsroomCharter.NewsroomURL
	tmplData["nr_tagline"] = newsroomCharter.Tagline
	tmplData["nr_mission"] = newsroomCharter.Mission.AsMap()
	if newsroomCharter.SocialURLs != nil {
		tmplData["nr_social_urls"] = newsroomCharter.SocialURLs.AsMap()
	}

	roster := []map[string]interface{}{}
	for _, member := range newsroomCharter.Roster {
		roster = append(roster, member.AsMap())
	}
	tmplData["nr_roster"] = roster

	signatures := []map[string]interface{}{}
	for _, signature := range newsroomCharter.Signatures {
		signatures = append(signatures, signature.AsMap())
	}
	tmplData["nr_signatures"] = signatures

	return tmplData
}

func (s *Service) buildLogoURLMarkup(logoURL string) string {
	return fmt.Sprintf("<img src=\"%v\" />", logoURL)
}

func (s *Service) setGrantRequestedFlag(newsroomOwnerUID string, requested bool) error {
	grantRequestedUpdateFn := func(d *SignupUserJSONData) (*SignupUserJSONData, error) {
		d.GrantRequested = &requested
		return d, nil
	}
	return s.alterUserDataInJSONStore(newsroomOwnerUID, grantRequestedUpdateFn)
}

func (s *Service) setGrantApprovedFlag(newsroomOwnerUID string, approved bool) error {
	grantApproveUpdateFn := func(d *SignupUserJSONData) (*SignupUserJSONData, error) {
		if !*d.GrantRequested {
			return nil, fmt.Errorf("Grant was not requested, failing approval")
		}
		d.GrantApproved = &approved
		return d, nil
	}
	return s.alterUserDataInJSONStore(newsroomOwnerUID, grantApproveUpdateFn)
}

// RetrieveUserJSONData gets SignupUserJSONData for a given user
func (s *Service) RetrieveUserJSONData(newsroomOwnerUID string) (*SignupUserJSONData, error) {
	s.alterMutex.Lock()
	defer s.alterMutex.Unlock()

	// Set both the namespace and ID as the newsroom owner ID
	jsonbs, err := s.jsonbService.RetrieveJSONb(
		DefaultJsonbID,
		jsonstore.DefaultJsonbGraphqlNs,
		newsroomOwnerUID,
	)
	if err != nil {
		return nil, err
	}
	if len(jsonbs) != 1 {
		return nil, fmt.Errorf("retrieved more than 1 result from the JSONb store")
	}

	jsonb := jsonbs[0]

	// Unmarshall json, flip the switch, then re-save
	signupData := &SignupUserJSONData{}
	err = json.Unmarshal([]byte(jsonb.RawJSON), signupData)
	if err != nil {
		return nil, err
	}

	return signupData, nil
}

func (s *Service) saveUserJSONData(newsroomOwnerUID string, signupData *SignupUserJSONData) error {
	s.alterMutex.Lock()
	defer s.alterMutex.Unlock()

	bys, err := json.Marshal(signupData)
	if err != nil {
		return err
	}

	_, err = s.jsonbService.SaveRawJSONb(
		DefaultJsonbID,
		jsonstore.DefaultJsonbGraphqlNs,
		newsroomOwnerUID,
		string(bys),
		&newsroomOwnerUID,
	)

	return err
}

type userDataUpdateFn func(*SignupUserJSONData) (*SignupUserJSONData, error)

func (s *Service) alterUserDataInJSONStore(newsroomOwnerUID string, updateFn userDataUpdateFn) error {
	signupData, err := s.RetrieveUserJSONData(newsroomOwnerUID)
	if err != nil {
		return err
	}

	signupData, err = updateFn(signupData)
	if err != nil {
		return err
	}

	return s.saveUserJSONData(newsroomOwnerUID, signupData)
}

func (s *Service) addToRegistryAlertsList(emailAddress string) error {
	if s.registryListID != "" {
		err := utils.AddToRegistryAlertsList(emailAddress, s.registryListID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) sendApplicationCompleteEmail(newsroomOwnerUID string, emailAddress string) error {
	if s.emailer == nil {
		return nil
	}

	applyLen, err := s.getApplyStageLength()
	if err != nil {
		return err
	}

	signupData, err := s.RetrieveUserJSONData(newsroomOwnerUID)
	if err != nil {
		return err
	}

	tmplData := email.TemplateData{
		"greeting":         emailAddress,
		"apply_stage_days": s.convertToDaysStr(applyLen),
		"nr_name":          signupData.Charter.Name,
	}

	tmplReq := &email.SendTemplateEmailRequest{
		ToName:       emailAddress,
		ToEmail:      emailAddress,
		FromName:     defaultFromEmailName,
		FromEmail:    defaultFromEmailAddress,
		TemplateID:   appliedToRegistryEmailTemplateID,
		TemplateData: tmplData,
		AsmGroupID:   defaultAsmGroupID,
	}
	return s.emailer.SendTemplateEmail(tmplReq)
}

func (s *Service) convertToDaysStr(applyStageLen *big.Int) string {
	periodInSecs := applyStageLen.Int64()
	inDays := periodInSecs / 60 / 60 / 24

	applyStageDays := fmt.Sprintf("%v days", inDays)
	if inDays <= 1 {
		applyStageDays = fmt.Sprintf("%v day", inDays)
	}
	return applyStageDays
}

func (s *Service) getApplyStageLength() (*big.Int, error) {
	if s.client == nil || s.parameterizerAddr == "" {
		return big.NewInt(defaultApplyStageLenSecs), nil
	}

	addr := common.HexToAddress(s.parameterizerAddr)
	pcontract, err := contract.NewParameterizerContract(addr, s.client)
	if err != nil {
		return nil, err
	}

	return pcontract.Get(&bind.CallOpts{}, applyStageLenName)
}
