package kyc

// Handlers for onfido/kyc REST APIs using the go-chi routing framework

// Handler for the onfido webhook

import (
	"crypto/hmac"
	"crypto/sha1" // nolint: gosec
	"encoding/hex"
	log "github.com/golang/glog"
	"net/http"

	"github.com/go-chi/render"
)

// OnfidoWebhookHandlerConfig is the config for the onfido webhook handler
type OnfidoWebhookHandlerConfig struct {
	OnfidoWebhookToken string
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

		if validOnfidoRequest(r, config, event.String()) {
			log.Errorf("Invalid signature for request: err: %v", err)
			err = render.Render(w, r, ErrInvalidOnfidoWebhookSig)
			if err != nil {
				log.Errorf("Error rendering response: err: %v", err)
			}
			return
		}

		// TODO(PN): Do something with it here.
		// Email?
		// Update entry KYC entry in db.

		err = render.Render(w, r, OkResponseNormal)
		if err != nil {
			log.Errorf("Error rendering response: err: %v", err)
		}
	}
}

func validOnfidoRequest(r *http.Request, config *OnfidoWebhookHandlerConfig,
	eventData string) bool {
	// If no webhook token passed in, then we disable the check
	if config.OnfidoWebhookToken == "" {
		return true
	}

	// Hex of HMAC SHA-1 of content + secret key
	mac := hmac.New(sha1.New, []byte(config.OnfidoWebhookToken))
	_, err := mac.Write([]byte(eventData))
	if err != nil {
		log.Errorf("Error writing hmac: err: %v", err)
		return false
	}
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Pull out X-Signature and compare
	xsig := r.Header.Get("X-Signature")
	if xsig == "" {
		return false
	}
	return expectedSig == xsig
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

// ErrInvalidOnfidoWebhookSig is the error in response to invalid signature on
// webhook
var ErrInvalidOnfidoWebhookSig = &ErrResponse{
	HTTPStatusCode: 401,
	StatusText:     "Invalid signature for event",
	AppCode:        802,
	ErrorText:      "The signatures don't match between payload and x-signature header",
}
