package invoicing

// Handlers for checkbook.io related invoicing using the
// go-chi routing framework

import (
	// "bytes"
	"errors"
	"net/http/httputil"
	"strings"
	// "context"
	"fmt"
	log "github.com/golang/glog"
	"net/http"

	uuid "github.com/satori/go.uuid"
	// "github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/joincivil/civil-api-server/pkg/utils"
)

const (
	// Feature flag to enable/disable the email check
	enableEmailCheck = false

	defaultInvoiceDescription = "Complete Your CVL Token Purchase"
)

var (
	wireTransferAlertRecipientEmails = []string{
		"cvl@civil.co",
		"civil@pipedrivemail.com",
		"peter@civil.co",
	}
	testWireTransferAlertRecipientEmails = []string{
		"peter@civil.co",
	}
)

// Request represents the incoming request for invoicing
type Request struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
	Amount    float64 `json:"amount"`
	// InvoiceDesc string  `json:"invoice_desc"`
	IsCheckbook bool   `json:"is_checkbook"`
	ReferredBy  string `json:"referred_by"`
}

// Bind implements the render.Binder interface
func (e *Request) Bind(r *http.Request) error {
	return nil
}

// Will render out error messages for responses.
func (e *Request) validate(w http.ResponseWriter, r *http.Request) error {
	var err error
	if e.FirstName == "" {
		err = render.Render(w, r, ErrInvalidRequest("first name"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	if e.LastName == "" {
		err = render.Render(w, r, ErrInvalidRequest("last name"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	if e.Email == "" {
		err = render.Render(w, r, ErrInvalidRequest("email"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	if e.Phone == "" {
		err = render.Render(w, r, ErrInvalidRequest("phone"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	// Restrict amount to less than/equal to 10k
	if e.Amount > 10000 {
		err = render.Render(w, r, ErrInvalidRequest("amount"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	return nil
}

// SendInvoiceHandlerConfig configures the SendInvoiceHandler
type SendInvoiceHandlerConfig struct {
	CheckbookIOClient *CheckbookIO
	InvoicePersister  *PostgresPersister
	Emailer           *utils.Emailer
	TestMode          bool
}

// SendInvoiceHandler returns the HTTP handler for the logic for sending an CVL invoice to a user
// who has requested it as well as store some metadata related to this request.
// TODO(PN): Refactor this into smaller pieces
func SendInvoiceHandler(config *SendInvoiceHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := httputil.DumpRequest(r, true)
		if err == nil {
			log.Infof("DEBUG: Request Body: %v", string(b))
		}

		request := &Request{}
		err = render.Bind(r, request)
		if err != nil {
			log.Errorf("Error binding params to Request: err: %v", err)
			err = render.Render(w, r, ErrInvalidRequest(""))
			if err != nil {
				log.Errorf("Error rendering error response: err: %v", err)
			}
			return
		}

		// Validate the request
		err = request.validate(w, r)
		// Already rendered error, return
		if err != nil {
			return
		}

		// If not checkbook and wire transfer, set these values to empty
		if !request.IsCheckbook {
			request.Amount = 0.0
			// request.InvoiceDesc = ""
		}

		existingInvoice := false
		if enableEmailCheck {
			// Check to see if the email address has already been used.
			invoices, ierr := config.InvoicePersister.Invoices("", request.Email, "", "")
			if ierr != nil {
				log.Errorf("Error checking for existing invoices: err: %v", ierr)
				ierr = render.Render(w, r, ErrSomethingBroke)
				if ierr != nil {
					log.Errorf("Error rendering error response: err: %v", ierr)
				}
				return
			}

			if len(invoices) > 0 {
				// If there is an invoice and already has an ID and number, then
				// we already have an invoice for it. If no invoice number, then
				// we might have failed to complete with invoice provider.
				invoice := invoices[0]
				if invoice.InvoiceID != "" {
					err = render.Render(w, r, ErrInvoiceExists)
					if err != nil {
						log.Errorf("Error rendering error response: err: %v", err)
					}
					return
				}
				existingInvoice = true
			}
		}

		// Generate a new referral code for this invoice.
		// If a user has multiple invoices, we will just figure out
		// overall total referrals via that user's email.
		referralCode, err := generateReferralCode()
		if err != nil {
			log.Errorf("Error generating new referrer code: %v", err)
		}

		// If there is referred by code, validate it
		referredBy := ""
		if request.ReferredBy != "" && validReferralCode(request.ReferredBy) {
			referredBy = request.ReferredBy
		} else {
			log.Errorf("Invalid referred by code: %v", request.ReferredBy)
		}

		// Make the request to CheckbookIO for the invoice
		fullName := fmt.Sprintf("%v %v", request.FirstName, request.LastName)

		// Save the user to the store with no invoice id yet
		postgresInvoice := &PostgresInvoice{
			Email:        request.Email,
			Name:         fullName,
			Phone:        request.Phone,
			Amount:       request.Amount,
			StopPoll:     false,
			IsCheckbook:  request.IsCheckbook,
			ReferralCode: referralCode,
			ReferredBy:   referredBy,
		}
		if !existingInvoice {
			err = config.InvoicePersister.SaveInvoice(postgresInvoice)
			if err != nil {
				log.Errorf("Error saving invoice: %v", err)
				err = render.Render(w, r, ErrSomethingBroke)
				if err != nil {
					log.Errorf("Error rendering error response: err: %v", err)
				}
				return
			}
		}

		if request.IsCheckbook {
			// Make request for invoice to checkbook.io
			invoiceRequest := &RequestInvoiceParams{
				Recipient: request.Email,
				Name:      fullName,
				Amount:    request.Amount,
				// Description: request.InvoiceDesc,
				Description: defaultInvoiceDescription,
			}
			invoiceResponse, ierr := config.CheckbookIOClient.RequestInvoice(invoiceRequest)
			if ierr != nil {
				log.Errorf("Error calling checkbookIO: %v", ierr)
				ierr = render.Render(w, r, ErrCheckbookIOInvoicing)
				if ierr != nil {
					log.Errorf("Error rendering checkbook io error: err: %v", ierr)
				}
				return
			}

			postgresInvoice.Amount = invoiceResponse.Amount
			postgresInvoice.InvoiceID = invoiceResponse.ID
			postgresInvoice.InvoiceStatus = invoiceResponse.Status
			postgresInvoice.CheckID = invoiceResponse.CheckID
			postgresInvoice.InvoiceNum = invoiceResponse.Number
			updatedFields := []string{
				"Amount",
				"InvoiceID",
				"InvoiceNum",
				"InvoiceStatus",
				"CheckID",
			}

			// If invoice looks good, store the checkbookIO invoice ID and invoice status
			// If fails here, issues might arise from sending invoice, but not having
			// the invoice ids saved.
			err = config.InvoicePersister.UpdateInvoice(postgresInvoice, updatedFields)
			if err != nil {
				log.Errorf("Error saving invoice: %v", err)
				err = render.Render(w, r, ErrSomethingBroke)
				if err != nil {
					log.Errorf("Error rendering error response: err: %v", err)
				}
				return
			}

		} else {
			// This is a wire transfer request, so email ourselves
			recipients := testWireTransferAlertRecipientEmails
			if !config.TestMode {
				recipients = wireTransferAlertRecipientEmails
			}
			sendWireTransferAlertEmail(config.Emailer, request, recipients)
		}

		// Return the response
		err = render.Render(w, r, OkResponseNormal)
		if err != nil {
			log.Errorf("Should have rendered response: err: %v", err)
		}
	}
}

func sendWireTransferAlertEmail(emailer *utils.Emailer, req *Request, recipientEmails []string) {
	go func() {
		text := fmt.Sprintf(
			"First: %v\nLast:%v\nEmail: %v\nPhone: %v",
			req.FirstName,
			req.LastName,
			req.Email,
			req.Phone,
		)
		html := fmt.Sprintf(
			`<p>First: %v</p>
			<p>Last: %v</p>
			<p>Email: %v</p>
			<p>Phone: %v</p>`,
			req.FirstName,
			req.LastName,
			req.Email,
			req.Phone,
		)
		subject := fmt.Sprintf("Wire Transfer Inquiry: %v %v", req.FirstName, req.LastName)

		emailReq := &utils.SendEmailRequest{
			ToName:    "The Civil Media Company",
			FromName:  "The Civil Media Company",
			FromEmail: "support@civil.co",
			Subject:   subject,
			Text:      text,
			HTML:      html,
		}
		for _, email := range recipientEmails {
			emailReq.ToEmail = email
			err := emailer.SendEmail(emailReq)
			if err != nil {
				log.Errorf("Error sending wire transfer email: err: %v", err)
			}
		}
	}()
}

// CheckUpdate is the request body received from the checkbook.io webhook
type CheckUpdate struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// Will render out error messages for responses.
func (c *CheckUpdate) validate(w http.ResponseWriter, r *http.Request) error {
	var err error
	if c.ID == "" {
		log.Errorf("No check ID received")
		err = render.Render(w, r, ErrInvalidRequest("id"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	if c.Status == "" {
		log.Errorf("No updated status received")
		err = render.Render(w, r, ErrInvalidRequest("status"))
		if err != nil {
			return err
		}
		return errors.New("Invalid Request")
	}
	return nil
}

// Bind implements the render.Binder interface
func (c *CheckUpdate) Bind(r *http.Request) error {
	return nil
}

// CheckbookIOWebhookConfig configures the CheckbookIOWebhook
type CheckbookIOWebhookConfig struct {
	InvoicePersister *PostgresPersister
}

// CheckbookIOWebhookHandler is the handler for the Checkbook.io webhook handler.
// TODO(PN): Refactor this into smaller pieces
func CheckbookIOWebhookHandler(config *CheckbookIOWebhookConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		update := &CheckUpdate{}
		err := render.Bind(r, update)
		if err != nil {
			log.Errorf("Error binding params to CheckUpdate: err: %v", err)
			err = render.Render(w, r, OkResponseNormal)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		// Validate the request
		err = update.validate(w, r)
		// Already rendered error, return
		if err != nil {
			return
		}

		// Normalize the status string (lower, replace spaces with _)
		update.Status = strings.ToLower(update.Status)
		update.Status = strings.Replace(update.Status, " ", "_", -1)

		invoices, err := config.InvoicePersister.Invoices("", "", "", update.ID)
		if err != nil {
			log.Errorf("Error checking for existing invoices: err: %v", err)
			err = render.Render(w, r, OkResponseNormal)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		if len(invoices) == 0 {
			log.Errorf("No invoices found with check ID: %v", update.ID)
			err = render.Render(w, r, OkResponseNormal)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		invoice := invoices[0]
		if invoice.CheckStatus == update.Status {
			log.Info("Check status has not been updated, same check status")
			err = render.Render(w, r, OkResponseNormal)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		nowPaid := false
		if invoice.CheckStatus == CheckStatusUnpaid ||
			invoice.CheckStatus == CheckStatusInProcess {
			if update.Status == CheckStatusPaid {
				log.Infof("setting nowpaid to true")
				nowPaid = true
			}
		}

		invoice.CheckStatus = update.Status
		updatedFields := []string{"CheckStatus"}

		// If a check comes in as paid or void, then update the invoice as
		// paid or canceled
		if update.Status == CheckStatusPaid {
			invoice.InvoiceStatus = InvoiceStatusPaid
			updatedFields = append(updatedFields, "InvoiceStatus")
		} else if update.Status == CheckStatusVoid {
			invoice.InvoiceStatus = InvoiceStatusCanceled
			updatedFields = append(updatedFields, "InvoiceStatus")
		}

		err = config.InvoicePersister.UpdateInvoice(invoice, updatedFields)
		if err != nil {
			log.Errorf("Error saving invoice: %v", err)
			err = render.Render(w, r, ErrSomethingBroke)
			if err != nil {
				log.Errorf("Error rendering error response: err: %v", err)
			}
			return
		}

		if nowPaid {
			// Push a message to pubsub
			log.Infof("Check was just paid, so push message to pubsub")
		}

		err = render.Render(w, r, OkResponseNormal)
		if err != nil {
			log.Errorf("Error rendering response: err: %v", err)
		}
	}
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

// ErrSomethingBroke is the error in response to any errors from checkbook.io
var ErrSomethingBroke = &ErrResponse{
	HTTPStatusCode: 500,
	StatusText:     "Problem with civil internals",
	AppCode:        800,
}

// ErrCheckbookIOInvoicing is the error in response to any errors from checkbook.io
var ErrCheckbookIOInvoicing = &ErrResponse{
	HTTPStatusCode: 500,
	StatusText:     "Problem with invoicing provider",
	AppCode:        801,
	ErrorText:      "CheckbookIO returns an error, unable to invoice right now",
}

// ErrInvoiceExists is the error in response to an invoice already existing
// for a specific email.
var ErrInvoiceExists = &ErrResponse{
	HTTPStatusCode: 400,
	StatusText:     "Invoice with this email already exists",
	AppCode:        802,
}

// Render implements the Render func on render.Renderer interface.
// Stolen from the go-chi packages examples.
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.SetContentType(render.ContentTypeJSON)
	render.Status(r, e.HTTPStatusCode)
	return nil
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

func generateReferralCode() (string, error) {
	code, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return code.String(), nil
}

func validReferralCode(code string) bool {
	_, err := uuid.FromString(code)
	if err != nil {
		log.Errorf("err = %v", err)
		return false
	}
	return true
}
