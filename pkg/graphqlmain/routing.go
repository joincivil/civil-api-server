package graphqlmain

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/go-common/pkg/email"
	stripe "github.com/stripe/stripe-go"
	webhook "github.com/stripe/stripe-go/webhook"
	"io/ioutil"
	"net/http"
)

func webhookRouting(deps ServerDeps) error {
	deps.Router.Post("/webhook", func(w http.ResponseWriter, r *http.Request) {
		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Error reading request body: %v\n", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		endpointSecret := deps.Config.StripeWebhookSigningSecret
		event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), endpointSecret)

		if err != nil {
			log.Errorf("Error verifying webhook signature: %v\n", err)
			w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
			return
		}

		if event.Type == "payment_intent.succeeded" {
			var paymentIntent stripe.PaymentIntent
			err := json.Unmarshal(event.Data.Raw, &paymentIntent)
			if err != nil {
				log.Errorf("Error parsing webhook JSON: %v\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			ownerID := paymentIntent.Metadata["posts"]
			post, err := deps.PostService.GetPost(ownerID)
			if err != nil {
				log.Errorf("Error getting post from ownerID: %v\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			amount := (float64(paymentIntent.Amount) / 100.0)
			tmplData, err := getStripePaymentEmailTemplateData(post, amount, deps.ChannelService, deps.NewsroomService)
			if err != nil {
				log.Errorf("Error getting email template data: %v\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = deps.PaymentService.ConfirmStripePaymentIntent(paymentIntent.ID, paymentIntent.PaymentMethod.ID, amount, post.GetType(), tmplData)
			if err != nil {
				log.Errorf("Error updating successful payment: %v\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

		} else if event.Type == "payment_intent.payment_failed" {
			var paymentIntent stripe.PaymentIntent
			err := json.Unmarshal(event.Data.Raw, &paymentIntent)
			if err != nil {
				log.Errorf("Error parsing webhook JSON: %v\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_, err = deps.PaymentService.FailStripePaymentIntent(paymentIntent.ID)
			if err != nil {
				log.Errorf("Error updating failed payment: %v\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})
	return nil
}

func getStripePaymentEmailTemplateData(post posts.Post, amount float64, channelService *channels.Service, newsroomService newsrooms.Service) (email.TemplateData, error) {
	channel, err := channelService.GetChannel(post.GetChannelID())
	if err != nil {
		return nil, err
	}
	newsroom, err := (newsroomService).GetNewsroomByAddress(channel.Reference)
	if err != nil {
		return nil, err
	}
	if post.GetType() == posts.TypeBoost {
		boost := post.(*posts.Boost)
		return (email.TemplateData{
			"newsroom_name":      newsroom.Name,
			"boost_short_desc":   boost.Title,
			"payment_amount_usd": amount,
			"boost_id":           boost.ID,
		}), nil
	} else if post.GetType() == posts.TypeExternalLink {
		return (email.TemplateData{
			"newsroom_name":      newsroom.Name,
			"payment_amount_usd": amount,
		}), nil
	}
	return nil, errors.New("NOT IMPLEMENTED")
}

func healthCheckRouting(router chi.Router) error {
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK")) // nolint: errcheck
	})
	return nil
}

func nrsignupRouting(deps ServerDeps) error {

	grantApproveConfig := &nrsignup.NewsroomSignupApproveGrantConfig{
		NrsignupService: deps.NewsroomSignupService,
		TokenGenerator:  deps.JwtGenerator,
	}

	// Set some rate limiters for the invoice handlers
	limiter := tollbooth.NewLimiter(2, nil) // 2 req/sec max
	limiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	limiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	deps.Router.Route(fmt.Sprintf("/%v/nrsignup", invoicingVersion), func(r chi.Router) {
		r.Route(fmt.Sprintf("/grantapprove/{%v}", nrsignup.GrantApproveTokenURLParam), func(r chi.Router) {
			r.Use(tollbooth_chi.LimitHandler(limiter))
			r.Get("/", nrsignup.NewsroomSignupApproveGrantHandler(grantApproveConfig))
		})
	})

	return nil
}
