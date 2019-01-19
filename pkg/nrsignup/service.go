package nrsignup

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/joincivil/civil-api-server/pkg/jsonstore"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/go-common/pkg/email"
)

const (
	// ApprovedSubValue is the value embedded in the token sub value to indicate
	// approval
	ApprovedSubValue = "app"

	// RejectedSubValue is the value embedded in the token sub value to indicate
	// rejection
	RejectedSubValue = "rej"
)

const (
	approvalLinkExpirySecs = 60 * 60 * 24 * 14 // 14 days
	rejectLinkExpirySecs   = 60 * 60 * 24 * 14 // 14 days

	signupWelcomeEmailTemplateID       = "d-125a9b151d99483f9dc8a8cc61fcb786"
	requestGrantUserEmailTemplateID    = "d-f17d8ba462ce4ac9ab24e447d9ee099d"
	requestGrantCouncilEmailTemplateID = "d-2ffc71848ea743b0a5b56a7c0d6b9ac3"

	grantApprovalUserEmailTemplateID = ""
	// grantApprovalCouncilEmailTemplateID = ""

	foundationEmailName = "The Civil Council"
	// foundationEmailAddress = "foundation@civil.co"
	foundationEmailAddress = "peter@civil.co"
	councilEmailName       = "The Civil Council"

	registryEmailName   = "The Civil Registry"
	noreplyEmailAddress = "noreply@civil.co"

	defaultFromEmailName    = foundationEmailName
	defaultFromEmailAddress = foundationEmailAddress

	defaultAsmGroupID = 8328 // Civil Registry Alerts

	grantApprovalURI = "v1/nrsignup/grantapprove"
)

// NewNewsroomSignupService is a convenience function to initialize a new newsroom
// signup service struct
func NewNewsroomSignupService(emailer *email.Emailer, userService *users.UserService,
	jsonbService *jsonstore.Service, tokenGenerator *auth.JwtTokenGenerator,
	grantLandingProtoHost string) (*Service, error) {
	return &Service{
		emailer:               emailer,
		userService:           userService,
		jsonbService:          jsonbService,
		tokenGenerator:        tokenGenerator,
		grantLandingProtoHost: grantLandingProtoHost,
	}, nil
}

// Service is a struct and methods used for handling newsroom signup functionality
type Service struct {
	emailer               *email.Emailer
	userService           *users.UserService
	jsonbService          *jsonstore.Service
	tokenGenerator        *auth.JwtTokenGenerator
	grantLandingProtoHost string
	alterMutex            sync.Mutex
}

// SendWelcomeEmail sends a newsroom signup welcome email to the given newsroom owner.
func (s *Service) SendWelcomeEmail(newsroomOwnerUID string) error {
	user, err := s.userService.MaybeGetUser(users.UserCriteria{
		UID: newsroomOwnerUID,
	})
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("No user found: uid: %v", newsroomOwnerUID)
	}

	tmplData := email.TemplateData{}
	tmplReq := &email.SendTemplateEmailRequest{
		ToName:       user.Email,
		ToEmail:      user.Email,
		FromName:     defaultFromEmailName,
		FromEmail:    defaultFromEmailAddress,
		TemplateID:   signupWelcomeEmailTemplateID,
		TemplateData: tmplData,
		AsmGroupID:   defaultAsmGroupID,
	}

	return s.emailer.SendTemplateEmail(tmplReq)
}

// RequestGrant sends a request for a grant to the Foundation on behalf of a newsroom.
// Sends along the data in the newsroom charter for review.
// Will send emails to the Foundation and newsroom owner.
// The Foundation email will have magic links to approve/reject the grant.
func (s *Service) RequestGrant(newsroomOwnerUID string, newsroomData *Newsroom) error {
	user, err := s.userService.MaybeGetUser(users.UserCriteria{
		UID: newsroomOwnerUID,
	})
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("No user found: uid: %v", newsroomOwnerUID)
	}

	// Set the grant requested flag to true for this user UID
	err = s.setGrantRequestedFlag(newsroomOwnerUID, true)
	if err != nil {
		return err
	}

	// Email to council with charter info
	tmplData := email.TemplateData{}
	s.buildCharterDataIntoTemplate(tmplData, newsroomData)
	_, err = s.buildGrantDecisionLinksIntoTemplate(tmplData, newsroomOwnerUID)
	if err != nil {
		return err
	}

	tmplReq := &email.SendTemplateEmailRequest{
		ToName:       councilEmailName,
		ToEmail:      foundationEmailAddress,
		FromName:     registryEmailName,
		FromEmail:    noreplyEmailAddress,
		TemplateID:   requestGrantCouncilEmailTemplateID,
		TemplateData: tmplData,
		AsmGroupID:   defaultAsmGroupID,
	}
	err1 := s.emailer.SendTemplateEmail(tmplReq)

	// Email to newsroom owner to tell them to wait for a response
	// TODO(PN): Do we need any other data for this template?
	tmplData = email.TemplateData{}
	tmplReq = &email.SendTemplateEmailRequest{
		ToName:       user.Email,
		ToEmail:      user.Email,
		FromName:     defaultFromEmailName,
		FromEmail:    defaultFromEmailAddress,
		TemplateID:   requestGrantUserEmailTemplateID,
		TemplateData: tmplData,
		AsmGroupID:   defaultAsmGroupID,
	}
	err2 := s.emailer.SendTemplateEmail(tmplReq)

	if err1 != nil {
		return fmt.Errorf("Failed to send grant request foundation email: err: %v", err1)
	}
	if err2 != nil {
		return fmt.Errorf("Failed to send grant request user email: err: %v", err2)
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
		return fmt.Errorf("No user found: uid: %v", newsroomOwnerUID)
	}

	// Set the grant requested flag to true for this user UID
	err = s.setGrantApprovedFlag(newsroomOwnerUID, approved)
	if err != nil {
		return err
	}

	// Email to newsroom owner to tell them to wait for a response
	// TODO(PN): Do we need any other data for this template?
	// TODO(PN): Will the foundation send this email manually?
	tmplData := email.TemplateData{}
	tmplData["nr_grant_approved"] = approved

	tmplReq := &email.SendTemplateEmailRequest{
		ToName:       user.Email,
		ToEmail:      user.Email,
		FromName:     defaultFromEmailName,
		FromEmail:    defaultFromEmailAddress,
		TemplateID:   grantApprovalUserEmailTemplateID,
		TemplateData: tmplData,
		AsmGroupID:   defaultAsmGroupID,
	}
	err = s.emailer.SendTemplateEmail(tmplReq)
	if err != nil {
		return fmt.Errorf("Failed to send grant approval user email: err: %v", err)
	}
	return nil
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

	link := fmt.Sprintf("%v/%v?t=%v", s.grantLandingProtoHost, grantApprovalURI, token)
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
		approveLink,
	)
	return tmplData, nil
}

func (s *Service) buildCharterDataIntoTemplate(tmplData email.TemplateData,
	newsroomData *Newsroom) email.TemplateData {
	newsroomCharter := newsroomData.Charter

	tmplData["nr_name"] = newsroomData.Name
	tmplData["nr_logo_url"] = newsroomCharter.LogoURL
	tmplData["nr_logo_markup"] = s.buildLogoURLMarkup(newsroomCharter.LogoURL)
	tmplData["nr_url"] = newsroomCharter.NewsroomURL
	tmplData["nr_tagline"] = newsroomCharter.Tagline
	tmplData["nr_mission"] = newsroomCharter.Mission.AsMap()

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

	socials := []map[string]interface{}{}
	for _, social := range newsroomCharter.SocialURLs {
		socials = append(socials, social.AsMap())
	}
	tmplData["nr_social_urls"] = socials
	return tmplData
}

func (s *Service) buildLogoURLMarkup(logoURL string) string {
	return fmt.Sprintf("<img src=\"%v\" />", logoURL)
}

func (s *Service) setGrantRequestedFlag(newsroomOwnerUID string, requested bool) error {
	grantRequestedUpdateFn := func(d *SignupUserJSONData) *SignupUserJSONData {
		d.GrantRequested = requested
		return d
	}
	return s.alterUserDataInJSONStore(newsroomOwnerUID, grantRequestedUpdateFn)
}

func (s *Service) setGrantApprovedFlag(newsroomOwnerUID string, approved bool) error {
	grantApproveUpdateFn := func(d *SignupUserJSONData) *SignupUserJSONData {
		d.GrantApproved = approved
		return d
	}
	return s.alterUserDataInJSONStore(newsroomOwnerUID, grantApproveUpdateFn)
}

type userDataUpdateFn func(*SignupUserJSONData) *SignupUserJSONData

func (s *Service) alterUserDataInJSONStore(newsroomOwnerUID string, updateFn userDataUpdateFn) error {
	s.alterMutex.Lock()
	defer s.alterMutex.Unlock()

	// Set both the namespace and ID as the newsroom owner ID
	// TODO(PN): Confirm this as the key with the client developers.
	jsonbs, err := s.jsonbService.RetrieveJSONb(
		newsroomOwnerUID,
		newsroomOwnerUID,
	)
	if err != nil {
		return err
	}
	if len(jsonbs) != 1 {
		return fmt.Errorf("Retrieved more than 1 result from the JSONb store")
	}

	jsonb := jsonbs[0]

	// Unmarshall json, flip the switch, then re-save
	signupData := &SignupUserJSONData{}
	err = json.Unmarshal([]byte(jsonb.RawJSON), signupData)
	if err != nil {
		return err
	}

	signupData = updateFn(signupData)

	bys, err := json.Marshal(signupData)
	if err != nil {
		return err
	}

	_, err = s.jsonbService.SaveRawJSONb(
		newsroomOwnerUID,
		newsroomOwnerUID,
		string(bys),
	)
	return err
}
