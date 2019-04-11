package graphqlmain

import (
	"fmt"
	"net/http"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/kyc"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/utils"
	cemail "github.com/joincivil/go-common/pkg/email"
)

func healthCheckRouting(router chi.Router) error {
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK")) // nolint: errcheck
	})
	return nil
}

func invoicingRouting(router chi.Router, client *invoicing.CheckbookIO,
	persister *invoicing.PostgresPersister, emailer *cemail.Emailer, testMode bool) error {
	invoicingConfig := &invoicing.SendInvoiceHandlerConfig{
		CheckbookIOClient: client,
		InvoicePersister:  persister,
		Emailer:           emailer,
		TestMode:          testMode,
	}
	whConfig := &invoicing.CheckbookIOWebhookConfig{
		InvoicePersister: persister,
		Emailer:          emailer,
	}

	// Set some rate limiters for the invoice handlers
	limiter := tollbooth.NewLimiter(2, nil) // 2 req/sec max
	limiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	limiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	cblimiter := tollbooth.NewLimiter(10, nil) // 10 req/sec max
	cblimiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	cblimiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	router.Route(fmt.Sprintf("/%v/invoicing", invoicingVersion), func(r chi.Router) {
		r.Route("/send", func(r chi.Router) {
			r.Use(tollbooth_chi.LimitHandler(limiter))
			r.Post("/", invoicing.SendInvoiceHandler(invoicingConfig))
		})

		r.Route("/cb", func(r chi.Router) {
			r.Use(tollbooth_chi.LimitHandler(cblimiter))
			r.Post("/", invoicing.CheckbookIOWebhookHandler(whConfig))
		})
	})
	return nil
}

func kycRouting(router chi.Router, config *utils.GraphQLConfig, onfido *kyc.OnfidoAPI,
	emailer *cemail.Emailer) error {

	ofConfig := &kyc.OnfidoWebhookHandlerConfig{
		OnfidoWebhookToken: config.OnfidoWebhookToken,
	}

	cblimiter := tollbooth.NewLimiter(10, nil) // 10 req/sec max
	cblimiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	cblimiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	router.Route(fmt.Sprintf("/%v/kyc", invoicingVersion), func(r chi.Router) {
		r.Route("/cb", func(r chi.Router) {
			r.Use(tollbooth_chi.LimitHandler(cblimiter))
			r.Post("/", kyc.OnfidoWebhookHandler(ofConfig))
		})
	})
	return nil
}

func nrsignupRouting(router chi.Router, config *utils.GraphQLConfig,
	nrsignupService *nrsignup.Service, tokenGenerator *auth.JwtTokenGenerator) error {

	grantApproveConfig := &nrsignup.NewsroomSignupApproveGrantConfig{
		NrsignupService: nrsignupService,
		TokenGenerator:  tokenGenerator,
	}

	// Set some rate limiters for the invoice handlers
	limiter := tollbooth.NewLimiter(2, nil) // 2 req/sec max
	limiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	limiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	router.Route(fmt.Sprintf("/%v/nrsignup", invoicingVersion), func(r chi.Router) {
		r.Route(fmt.Sprintf("/grantapprove/{%v}", nrsignup.GrantApproveTokenURLParam), func(r chi.Router) {
			r.Use(tollbooth_chi.LimitHandler(limiter))
			r.Get("/", nrsignup.NewsroomSignupApproveGrantHandler(grantApproveConfig))
		})
	})
	return nil
}
