package invoicing

// Handlers for checkbook.io related invoicing using the
// go-chi routing framework

import (
	"net/url"
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

	defaultInvoiceDescription  = "Complete Your CVL Token Purchase"
	referralEmailTemplateID    = "d-33fbe062ad2d44bdbb4c584f75b9a576"
	postPaymentEmailTemplateID = "d-fc18db3e7e394aad92c0774858f0a1d7"
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

// GenerateReferralCode generates a new referral code
func GenerateReferralCode() (string, error) {
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

// Request represents the incoming request for invoicing
// If IsCheckbook flag is true, will send checkbook io invoice to user.  If
// false, will consider it a wire transfer.
// If IsThirdParty flag is true, will not send to checkbook io or consider it a
// wire transfer.  It will just record the user info and generate a referral code and
// send a referral email.  Use if user was billed via Token Foundry or other third party
// source.
type Request struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
	Amount    float64 `json:"amount"`
	// InvoiceDesc string  `json:"invoice_desc"`
	IsCheckbook  bool   `json:"is_checkbook"`
	IsThirdParty bool   `json:"is_third_party"`
	ReferredBy   string `json:"referred_by"`
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
	// Restrict amount to less than/equal to 10k for checkbook transactions
	if e.IsCheckbook && e.Amount > 10000 {
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

		// If not checkbook and is wire transfer, set these values to empty
		// If third party, save the amount
		if !request.IsCheckbook && !request.IsThirdParty {
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
		referralCode, err := GenerateReferralCode()
		if err != nil {
			log.Errorf("Error generating new referrer code: %v", err)
		}

		// If there is referred by code, validate it
		referredBy := ""
		if request.ReferredBy != "" {
			if validReferralCode(request.ReferredBy) {
				referredBy = request.ReferredBy
			} else {
				log.Errorf("Invalid referred by code: %v", request.ReferredBy)
			}
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
			IsThirdParty: request.IsThirdParty,
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

		if !request.IsThirdParty {
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
				go sendWireTransferAlertEmail(config.Emailer, request, recipients)
				log.Infof("send a wire transfer email")
			}
		}

		// Send the referral email
		go SendReferralProgramEmail(config.Emailer, request, referralCode)

		// Return the response
		err = render.Render(w, r, OkResponseNormal)
		if err != nil {
			log.Errorf("Should have rendered response: err: %v", err)
		}
	}
}

func sendWireTransferAlertEmail(emailer *utils.Emailer, req *Request, recipientEmails []string) {
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
}

// SendReferralProgramEmail sends the referrer email with the given referral code to the
// recipient specified in the request
func SendReferralProgramEmail(emailer *utils.Emailer, req *Request, referralCode string) {
	fullName := fmt.Sprintf("%v %v", req.FirstName, req.LastName)

	templateData := utils.TemplateData{}
	templateData["first_name"] = req.FirstName
	templateData["referral_link"] = referralLinkHTML(referralCode)
	templateData["referral_email"] = referralEmailHTML(referralCode)
	templateData["referral_twitter"] = referralTwitterHTML(referralCode)
	templateData["referral_fb"] = referralFacebookHTML(referralCode)

	emailReq := &utils.SendTemplateEmailRequest{
		ToName:       fullName,
		ToEmail:      req.Email,
		FromName:     "The Civil Media Company",
		FromEmail:    "support@civil.co",
		TemplateID:   referralEmailTemplateID,
		TemplateData: templateData,
		AsmGroupID:   7395,
	}
	err := emailer.SendTemplateEmail(emailReq)
	if err != nil {
		log.Errorf("Error sending referral email: err: %v", err)
	}
}

func referralLinkHTML(referralCode string) string {
	link := referralLink(referralCode)
	return fmt.Sprintf("<a href=\"%v\">%v</a>", link, link)
}

func referralLink(referralCode string) string {
	return fmt.Sprintf("http://civil.co/?referred_by=%v", referralCode)
}

func referralEmailHTML(referralCode string) string {
	return fmt.Sprintf("<a href=\"%v\">Email</a>", referralEmail(referralCode))
}

func referralEmail(referralCode string) string {
	referralLink := referralLink(referralCode)
	subject := "Share Civil, Earn CVL"
	body := fmt.Sprintf("%v", referralLink)
	return fmt.Sprintf("mailto:?body=%v&subject=%v", body, subject)
}

func referralTwitterHTML(referralCode string) string {
	return fmt.Sprintf("<a href=\"%v\">Twitter</a>", referralTwitter(referralCode))
}

func referralTwitter(referralCode string) string {
	referralLink := referralLink(referralCode)
	twitterMsg := fmt.Sprintf(
		"I support #journalism on @Join_Civil -- and so can you. If you contribute $100 to the Civil token sale, "+
			"you'll get an extra $100 of CVL with my referral code until 10/15: %v",
		referralLink,
	)
	escapedTwitterMsg := url.QueryEscape(twitterMsg)
	return fmt.Sprintf("https://twitter.com/home?status=%v", escapedTwitterMsg)
}

func referralFacebookHTML(referralCode string) string {
	return fmt.Sprintf("<a href=\"%v\">Facebook</a>", referralFacebook(referralCode))
}

func referralFacebook(referralCode string) string {
	referralLink := referralLink(referralCode)
	return fmt.Sprintf("https://www.facebook.com/sharer/sharer.php?u=%v", referralLink)
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
	Emailer          *utils.Emailer
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

		// Find the invoice in out DB from the ID
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

		// Has the status changed from unpaid to paid?
		nowPaid := false
		if invoice.InvoiceStatus == InvoiceStatusUnpaid ||
			invoice.InvoiceStatus == InvoiceStatusInProcess {
			if update.Status == CheckStatusPaid {
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

		// Update the invoice and check status
		err = config.InvoicePersister.UpdateInvoice(invoice, updatedFields)
		if err != nil {
			log.Errorf("Error saving invoice: %v", err)
			err = render.Render(w, r, ErrSomethingBroke)
			if err != nil {
				log.Errorf("Error rendering error response: err: %v", err)
			}
			return
		}

		// If the it was an unpaid to paid status, send email
		if nowPaid {
			SendPostPaymentEmail(config.Emailer, invoice.Email, invoice.Name)
			log.Infof("Post payment email sent to %v", invoice.Email)
		}

		err = render.Render(w, r, OkResponseNormal)
		if err != nil {
			log.Errorf("Error rendering response: err: %v", err)
		}
	}
}

// SendPostPaymentEmail sends the post payment instruction email
func SendPostPaymentEmail(emailer *utils.Emailer, email string, name string) {
	templateData := utils.TemplateData{}
	templateData["name"] = name

	emailReq := &utils.SendTemplateEmailRequest{
		ToName:       name,
		ToEmail:      email,
		FromName:     "The Civil Media Company",
		FromEmail:    "support@civil.co",
		TemplateID:   postPaymentEmailTemplateID,
		TemplateData: templateData,
		AsmGroupID:   7395,
	}
	err := emailer.SendTemplateEmail(emailReq)
	if err != nil {
		log.Errorf("Error sending post payment email: err: %v", err)
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
