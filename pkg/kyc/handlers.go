package kyc

// Handlers for onfido/kyc using the go-chi routing framework

// Handler for /v1/kyc/init
// Handler for /v1/kyc/finish
// Handler for the onfido webhook

import (
	"errors"
	"fmt"
	log "github.com/golang/glog"

	"github.com/go-chi/render"
	"net/http"

	"github.com/joincivil/civil-api-server/pkg/utils"
)

const (
	onfidoReferrer = "*://*.civil.co/*"
)

// OnfidoWebhookHandlerConfig is the config for the onfido webhook handler
type OnfidoWebhookHandlerConfig struct {
}

// OnfidoWebhookHandler is the handler for the Onfido webhook handler.
func OnfidoWebhookHandler(config *OnfidoWebhookHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event := &Event{}
		err := render.Bind(r, event)
		if err != nil {
			log.Errorf("Error binding params to Event: err: %v", err)
			err = render.Render(w, r, OkResponseNormal)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		log.Infof("Received onfido event: %v", event.String())

		// TODO(PN): Do something with it here.
		// Email?
		// Update entry KYC entry in db.
	}
}

// InitHandlerRequest is the request data for the KYC init handler request
// Based it on what TF asks for in their form.
type InitHandlerRequest struct {
	FirstName          string `json:"first_name"`
	MiddleName         string `json:"middle_name"`
	LastName           string `json:"last_name"`
	Email              string `json:"email"`
	Profession         string `json:"profession"` // Does this get passed?
	Nationality        string `json:"nationality"`
	CountryOfResidence string `json:"country_of_residence"`
	DateOfBirth        string `json:"dob"`
	BuildingNumber     string `json:"building_number"`
	Street             string `json:"street"`
	AptNumber          string `json:"apt_number"`
	City               string `json:"city"`
	State              string `json:"state"`
	Zipcode            string `json:"zipcode"`
}

// Bind implements the render.Binder interface
func (i *InitHandlerRequest) Bind(r *http.Request) error {
	return nil
}

func (i *InitHandlerRequest) validate(w http.ResponseWriter, r *http.Request) error {
	var err error
	if i.FirstName == "" {
		err = render.Render(w, r, ErrInvalidRequest("first name"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	if i.LastName == "" {
		err = render.Render(w, r, ErrInvalidRequest("last name"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	if i.Email == "" {
		err = render.Render(w, r, ErrInvalidRequest("email"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	if i.DateOfBirth == "" {
		err = render.Render(w, r, ErrInvalidRequest("dob"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	if i.Nationality == "" {
		err = render.Render(w, r, ErrInvalidRequest("nationality"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	if i.CountryOfResidence == "" {
		err = render.Render(w, r, ErrInvalidRequest("country_of_residence"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	return nil
}

// InitHandlerConfig is the config for the KYC init handler
type InitHandlerConfig struct {
	OnfidoAPI *OnfidoAPI
	Emailer   *utils.Emailer
}

// InitHandler is the handler for the KYC init handler.
func InitHandler(config *InitHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &InitHandlerRequest{}
		err := render.Bind(r, req)
		if err != nil {
			log.Errorf("Error binding params to InitHandlerRequest: err: %v", err)
			err = render.Render(w, r, OkResponseNormal)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		// Validate the fields
		err = req.validate(w, r)
		// Already rendered error, return
		if err != nil {
			return
		}

		// Create new applicant in Onfido
		applicant, err := createOnfidoApplicant(config, req)
		if err != nil {
			log.Errorf("Error creating new applicant in Onfido: err: %v", err)
			err = render.Render(w, r, ErrOnfido)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		// Generate a new JWT token for the SDK
		// TODO(PN): Only lasts 90 mins, so maybe need a way to refresh it if user takes a while?
		token, err := fetchNewOnfidoJWTToken(config, applicant)
		if err != nil {
			log.Errorf("Error creating new JWT token from Onfido: err: %v", err)
			err = render.Render(w, r, ErrOnfido)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		// Return the response
		response := &InitHandlerResponse{
			Token:       token,
			ApplicantID: applicant.ID,
		}
		err = render.Render(w, r, response)
		if err != nil {
			log.Errorf("Should have rendered response: err: %v", err)
		}

	}
}

// FinishHandlerConfig is the config for the finish handler
type FinishHandlerConfig struct {
	OnfidoAPI *OnfidoAPI
	Emailer   *utils.Emailer
}

// FinishHandlerRequest is the request for the finish handler
type FinishHandlerRequest struct {
	ApplicantID             string `json:"applicant_id"`
	FacialSimilarityVariant string `json:"fs_variant"`
}

// Bind implements the render.Binder interface
func (i *FinishHandlerRequest) Bind(r *http.Request) error {
	return nil
}

func (i *FinishHandlerRequest) validate(w http.ResponseWriter, r *http.Request) error {
	var err error
	if i.ApplicantID == "" {
		err = render.Render(w, r, ErrInvalidRequest("applicant id"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	return nil
}

// FinishHandler is the handler for the KYC finish handler.
func FinishHandler(config *FinishHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &FinishHandlerRequest{}
		err := render.Bind(r, req)
		if err != nil {
			log.Errorf("Error binding params to FinishHandlerRequest: err: %v", err)
			err = render.Render(w, r, OkResponseNormal)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		// Validate the fields
		err = req.validate(w, r)
		// Already rendered error, return
		if err != nil {
			return
		}

		// Create new check in Onfido
		check, err := createOnfidoCheck(config, req)
		if err != nil {
			log.Errorf("Error creating new check in Onfido: err: %v", err)
			err = render.Render(w, r, ErrOnfido)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		// Return the response
		response := &FinishHandlerResponse{
			Status:  check.Status,
			Reports: check.Reports,
		}
		err = render.Render(w, r, response)
		if err != nil {
			log.Errorf("Should have rendered response: err: %v", err)
		}

	}
}

func createOnfidoCheck(config *FinishHandlerConfig, req *FinishHandlerRequest) (*Check, error) {
	newCheck := &Check{
		Type: CheckTypeExpress,
		Reports: []Report{
			// *kyc.IdentityKycReport,
			*DocumentReport,
			*FacialSimilarityStandardReport,
			// *kyc.WatchlistKycReport,
		},
	}

	returnedCheck, err := config.OnfidoAPI.CreateCheck(req.ApplicantID, newCheck)
	if err != nil {
		return nil, err
	}
	return returnedCheck, nil
}

func createOnfidoApplicant(config *InitHandlerConfig, req *InitHandlerRequest) (*Applicant, error) {
	newAddress := Address{
		FlatNumber:     req.AptNumber,
		BuildingNumber: req.BuildingNumber,
		Street:         req.Street,
		Town:           req.City,
		State:          req.State,
		Postcode:       req.Zipcode,
		Country:        req.CountryOfResidence,
		StartDate:      "2007-01-01",
		EndDate:        "2010-01-01",
	}
	newApplicant := &Applicant{
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		LastName:   req.LastName,
		Email:      req.Email,
		Dob:        req.DateOfBirth,
		Country:    req.CountryOfResidence,
		Addresses:  []Address{newAddress},
	}
	returnedApplicant, err := config.OnfidoAPI.CreateApplicant(newApplicant)
	if err != nil {
		return nil, err
	}
	return returnedApplicant, nil
}

func fetchNewOnfidoJWTToken(config *InitHandlerConfig, applicant *Applicant) (string, error) {
	return config.OnfidoAPI.GenerateSDKToken(applicant.ID, onfidoReferrer)
}

// FinishHandlerResponse is a normal response for the Finish KYC handler
type FinishHandlerResponse struct {
	Status  string   `json:"status"`
	Reports []Report `json:"reports"`
}

// Render implements the Render func on render.Renderer interface.
func (o *FinishHandlerResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.SetContentType(render.ContentTypeJSON)
	render.Status(r, http.StatusOK)
	return nil
}

// InitHandlerResponse is a normal response for the Init KYC handler
type InitHandlerResponse struct {
	Token       string `json:"token"`
	ApplicantID string `json:"applicant_id"`
}

// Render implements the Render func on render.Renderer interface.
func (o *InitHandlerResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.SetContentType(render.ContentTypeJSON)
	render.Status(r, http.StatusOK)
	return nil
}

// OkResponse represents a generic OK message
type OkResponse struct {
	StatusText string `json:"status"`
}

// Render implements the Render func on render.Renderer interface.
func (o *OkResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.SetContentType(render.ContentTypeJSON)
	render.Status(r, http.StatusOK)
	return nil
}

// OkResponseNormal is a normal 200 response
var OkResponseNormal = &OkResponse{StatusText: "ok"}

// ErrResponse represents a generic error message.  Stolen from the go-chi packages examples.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

// Render implements the Render func on render.Renderer interface.
// Stolen from the go-chi packages examples.
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.SetContentType(render.ContentTypeJSON)
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrSomethingBroke is the error in response to any errors from checkbook.io
var ErrSomethingBroke = &ErrResponse{
	HTTPStatusCode: 500,
	StatusText:     "Problem with civil internals",
	AppCode:        800,
}

// ErrOnfido is the error in response to any errors from Onfido
var ErrOnfido = &ErrResponse{
	HTTPStatusCode: 500,
	StatusText:     "Problem with KYC provider",
	AppCode:        801,
	ErrorText:      "Onfido returns an error, unable to kyc right now",
}

// ErrInvalidRequest represents an error response for this handler.
func ErrInvalidRequest(missingField string) render.Renderer {
	var msg string
	if missingField != "" {
		msg = fmt.Sprintf("Missing information in the %v field", missingField)
	} else {
		msg = "Missing the necessary information to send invoice"
	}
	return &ErrResponse{
		HTTPStatusCode: 400,
		StatusText:     msg,
	}
}
