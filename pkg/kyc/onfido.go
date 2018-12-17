package kyc

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/golang/glog"

	chttp "github.com/joincivil/go-common/pkg/http"
)

const (
	// ProdAPIURL is the production URL for the Onfido API
	// Might also be the same for the sandbox URL?
	ProdAPIURL = "https://api.onfido.com"

	// DefaultTokenReferrer is the default token referrer to use
	DefaultTokenReferrer = "*://*.civil.co/*" //nolint: gosec
)

const (
	// ResourceTypeCheck represents a resource type "check"
	ResourceTypeCheck = "check"

	// ReportResultClear is the clear result for report
	ReportResultClear = "clear"
	// ReportResultConsider is the consider result for report
	ReportResultConsider = "consider"
	// ReportResultUnidentified is the unidentified result for report
	ReportResultUnidentified = "unidentified"

	// CheckResultClear is the clear result for report
	CheckResultClear = "clear"
	// CheckResultConsider is the consider result for report
	CheckResultConsider = "consider"

	// CheckStatusInProgress in_progress
	CheckStatusInProgress = "in_progress"
	// CheckStatusAwaitingApplicant awaiting_applicant
	CheckStatusAwaitingApplicant = "awaiting_applicant"
	// CheckStatusComplete complete
	CheckStatusComplete = "complete"
	// CheckStatusWithdrawn withdrawn
	CheckStatusWithdrawn = "withdrawn"
	// CheckStatusPaused paused
	CheckStatusPaused = "paused"
	// CheckStatusReopened reopened
	CheckStatusReopened = "reopened"

	// CheckTypeExpress is a check type express
	CheckTypeExpress = "express"
	// CheckTypeStandard is a check type standard
	CheckTypeStandard = "standard"

	// ApplicantGenderMale represents a male applicant value
	ApplicantGenderMale = "Male"
	// ApplicantGenderFemale represents a female applicant value
	ApplicantGenderFemale = "Female"

	// ReportNameIdentity represents an identity report type
	ReportNameIdentity = "identity"
	// ReportNameDocument represents a document report type
	ReportNameDocument = "document"
	// ReportNameFacialSimilarity represents a facial similarity report type
	ReportNameFacialSimilarity = "facial_similarity"
	// ReportNameWatchlist represents a watchlist report type
	ReportNameWatchlist = "watchlist"

	// ReportVariantFacialSimilarityStandard represents the standard facial sim variant
	ReportVariantFacialSimilarityStandard = "standard"
	// ReportVariantFacialSimilarityVideo represents the video facial sim variant
	ReportVariantFacialSimilarityVideo = "video"

	// WebhookActionReportCompleted indicates that a report has been completed
	WebhookActionReportCompleted = "report.completed"
	// WebhookActionReportWithdrawn indicates that a report has been withdrawn
	WebhookActionReportWithdrawn = "report.withdrawn"
	// WebhookActionReportCancelled indicates that a report has been cancelled
	WebhookActionReportCancelled = "report.cancelled"
	// WebhookActionReportAwaitingApproval indicates that a report is awaiting approval
	WebhookActionReportAwaitingApproval = "report.awaiting_approval"

	// WebhookActionCheckCompleted indicates that a check has completed
	WebhookActionCheckCompleted = "check.completed"
	// WebhookActionCheckWithdrawn indicates that a check has been withdrawn
	WebhookActionCheckWithdrawn = "check.withdrawn"
	// WebhookActionCheckStarted indicates that a check has been started
	WebhookActionCheckStarted = "check.started"
	// WebhookActionCheckReopened indicates that a check has been reopened
	WebhookActionCheckReopened = "check.reopened"
)

// NewOnfidoAPI is a convenience function to create a new OnfidoAPI struct
func NewOnfidoAPI(baseAPIURL string, apiKey string) *OnfidoAPI {
	authHeader := fmt.Sprintf("Token token=%v", apiKey)
	return &OnfidoAPI{
		apiKey:     apiKey,
		baseAPIURL: baseAPIURL,
		rest:       chttp.NewRestHelper(baseAPIURL, authHeader),
	}
}

// OnfidoAPI is a wrapper around the Onfido API
type OnfidoAPI struct {
	baseAPIURL string
	apiKey     string
	rest       *chttp.RestHelper
}

// Hash represents a map of values
type Hash map[string]interface{}

// IdentityStandardReport presents a identity report type of standard
// variation.
var IdentityStandardReport = &Report{
	Name:    ReportNameIdentity,
	Variant: "standard",
}

// IdentityKycReport presents a identity report type of KYC
// variation
var IdentityKycReport = &Report{
	Name:    ReportNameIdentity,
	Variant: "kyc",
}

// DocumentReport presents a document report type
var DocumentReport = &Report{
	Name: ReportNameDocument,
}

// FacialSimilarityStandardReport presents a facial_similiarity report type
// with the standard variation
var FacialSimilarityStandardReport = &Report{
	Name:    ReportNameFacialSimilarity,
	Variant: ReportVariantFacialSimilarityStandard,
}

// FacialSimilarityVideoReport presents a facial_similiarity report type
// with the video variation
var FacialSimilarityVideoReport = &Report{
	Name:    ReportNameFacialSimilarity,
	Variant: ReportVariantFacialSimilarityVideo,
}

// WatchlistKycReport presents a watchlist report type of KYC
// variation
var WatchlistKycReport = &Report{
	Name:    ReportNameWatchlist,
	Variant: "kyc",
}

// WatchlistFullReport presents a watchlist report type of full
// variation
var WatchlistFullReport = &Report{
	Name:    ReportNameWatchlist,
	Variant: "full",
}

// Report represents a report for an applicant
type Report struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	Href       string `json:"href,omitempty"`
	Status     string `json:"status,omitempty"`
	Result     string `json:"result,omitempty"`
	SubResult  string `json:"sub_result,omitempty"`
	Variant    string `json:"variant,omitempty"`
	Options    string `json:"options,omitempty"`
	Breakdown  Hash   `json:"breakdown,omitempty"`
	Properties Hash   `json:"properties,omitempty"`
}

// Address represents a physical address for an applicant
type Address struct {
	FlatNumber     string `json:"flat_number,omitempty"`
	BuildingNumber string `json:"building_number,omitempty"`
	BuildingName   string `json:"building_name,omitempty"`
	Street         string `json:"street,omitempty"`
	SubStreet      string `json:"sub_street,omitempty"`
	Town           string `json:"town,omitempty"`
	State          string `json:"state,omitempty"`
	Postcode       string `json:"postcode,omitempty"`
	Country        string `json:"country,omitempty"` // 3 letter code
	StartDate      string `json:"start_date,omitempty"`
	EndDate        string `json:"end_date,omitempty"`
}

// IDNumber represents some ID number for an applicant
type IDNumber struct {
	Type      string `json:"type,omitempty"`
	Value     string `json:"value,omitempty"`
	StateCode string `json:"state_code,omitempty"`
}

// Applicant represents an applicant. Used as a request and response object.
type Applicant struct {
	ID             string     `json:"id,omitempty"`
	CreatedAt      string     `json:"created_at,omitempty"`
	Href           string     `json:"href,omitempty"`
	Title          string     `json:"title,omitempty"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	MiddleName     string     `json:"middle_name,omitempty"`
	Email          string     `json:"email"`
	Gender         string     `json:"gender,omitempty"`
	Dob            string     `json:"dob,omitempty"`
	Telephone      string     `json:"telephone,omitempty"` // Landline phone
	Mobile         string     `json:"mobile,omitempty"`
	Country        string     `json:"country,omitempty"` // 3 letter country code for verification
	MothersMaiden  string     `json:"mothers_maiden_name,omitempty"`
	PrevLastName   string     `json:"previous_last_name,omitempty"`
	Nationality    string     `json:"nationality,omitempty"`
	CountryOfBirth string     `json:"country_of_birth,omitempty"`
	TownOfBirth    string     `json:"town_of_birth,omitempty"`
	Addresses      []Address  `json:"addresses,omitempty"`
	IDNumbers      []IDNumber `json:"id_numbers,omitempty"`
	Sandbox        bool       `json:"sandbox,omitempty"`
}

// Check presents a check. Used as a request and response struct.
type Check struct {
	ID                           string   `json:"id,omitempty"`
	CreatedAt                    string   `json:"created_at,omitempty"`
	Href                         string   `json:"href,omitempty"`
	Type                         string   `json:"type,omitempty"`
	Result                       string   `json:"result,omitempty"`
	Status                       string   `json:"status,omitempty"`
	DownloadURI                  string   `json:"download_uri,omitempty"`
	FormURI                      string   `json:"form_uri,omitempty"`
	RedirectURI                  string   `json:"redirect_uri,omitempty"`
	Reports                      []Report `json:"reports,omitempty"`
	ReportTypeGroups             []string `json:"report_type_groups,omitempty"`
	CriminalHistoryReportDetails Hash     `json:"criminal_history_report_details,omitempty"`
	Tags                         []string `json:"tags,omitempty"`
	SuppressFormEmails           bool     `json:"suppress_form_emails,omitempty"`
	Async                        bool     `json:"async,omitempty"`
	Consider                     []string `json:"consider,omitempty"`
}

// Event is the request body received from Onfido via their webhook
type Event struct {
	Payload *EventPayload `json:"payload"`
}

// String returns the onfido event as a JSON string
func (e *Event) String() string {
	bys, err := json.Marshal(e)
	if err != nil {
		log.Errorf("Error marshalling to string: err: %v", err)
		return ""
	}
	return string(bys)
}

// Bind implements the render.Binder interface
func (e *Event) Bind(r *http.Request) error {
	return nil
}

// EventPayload is the payload of the OnfidoEvent
type EventPayload struct {
	ResourceType string              `json:"resource_type"`
	Action       string              `json:"action"`
	Object       *EventPayloadObject `json:"object"`
}

// Bind implements the render.Binder interface
func (e *EventPayload) Bind(r *http.Request) error {
	return nil
}

// EventPayloadObject is the object within the payload of the OnfidoEvent
type EventPayloadObject struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	CompletedAt string `json:"completed_at"`
	Href        string `json:"href"`
}

// Bind implements the render.Binder interface
func (e *EventPayloadObject) Bind(r *http.Request) error {
	return nil
}

// CreateApplicant calls Onfido to create a new applicant.  This is the first
// step of the KYC process.
func (o *OnfidoAPI) CreateApplicant(applicant *Applicant) (*Applicant, error) {
	endpointURI := "v2/applicants"

	bys, err := o.rest.SendRequest(endpointURI, http.MethodPost, nil, applicant)
	if err != nil {
		return nil, err
	}

	applicantResp := &Applicant{}
	err = json.Unmarshal(bys, applicantResp)
	if err != nil {
		return nil, err
	}
	return applicantResp, nil
}

// Token is the token response from the SDK token API
type Token struct {
	Token string `json:"token,omitempty"`
}

// GenerateSDKToken retrieves a new JWT token from Onfido given the applicant ID
// Should be called after CreateApplicant has occurred.
func (o *OnfidoAPI) GenerateSDKToken(applicantID string, referrer string) (string, error) {
	endpointURI := "v2/sdk_token"

	tokenRequestInput := struct {
		ApplicantID string `json:"applicant_id"`
		Referrer    string `json:"referrer"`
	}{
		ApplicantID: applicantID,
		Referrer:    referrer,
	}

	bys, err := o.rest.SendRequest(endpointURI, http.MethodPost, nil, tokenRequestInput)
	if err != nil {
		return "", err
	}

	tokenResp := &Token{}
	err = json.Unmarshal(bys, tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.Token, nil
}

// CreateCheck calls Onfido to start the check process. This will be called after
// the user has gone through the steps via the SDK to capture face and document data.
func (o *OnfidoAPI) CreateCheck(applicantID string, check *Check) (*Check, error) {
	endpointURI := fmt.Sprintf("v2/applicants/%v/checks", applicantID)

	bys, err := o.rest.SendRequest(endpointURI, http.MethodPost, nil, check)
	if err != nil {
		return nil, err
	}

	checkResp := &Check{}
	err = json.Unmarshal(bys, checkResp)
	if err != nil {
		return nil, err
	}
	return checkResp, nil
}

// RetrieveCheckFromHref retrieves a Check from a URL returned by Onfido.
func (o *OnfidoAPI) RetrieveCheckFromHref(url string) (*Check, error) {
	bys, err := o.rest.SendRequestToURL(url, http.MethodGet, nil, nil)
	if err != nil {
		return nil, err
	}

	checkResp := &Check{}
	err = json.Unmarshal(bys, checkResp)
	if err != nil {
		return nil, err
	}
	return checkResp, nil
}

// RetrieveReport retrieves a Report given the Check ID and Report ID
func (o *OnfidoAPI) RetrieveReport(checkID string, reportID string) (*Report, error) {
	endpointURI := fmt.Sprintf("v2/checks/%v/reports/%v", checkID, reportID)

	bys, err := o.rest.SendRequest(endpointURI, http.MethodGet, nil, nil)
	if err != nil {
		return nil, err
	}

	reportResp := &Report{}
	err = json.Unmarshal(bys, reportResp)
	if err != nil {
		return nil, err
	}
	return reportResp, nil
}
