package payments

import (
	"encoding/json"
	"github.com/go-chi/chi"
	log "github.com/golang/glog"
	stripe "github.com/stripe/stripe-go"
	webhook "github.com/stripe/stripe-go/webhook"
	"io/ioutil"
	"net/http"
)

const (
	stripeSignature = "Stripe-Signature"
)

// WebhookRouting provides an entry point for stripe webhook events
func (s *Service) WebhookRouting(router chi.Router, stripeWebhookSigningSecret string) error {
	router.Post("/webhook", func(w http.ResponseWriter, r *http.Request) {
		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Error reading request body: %v\n", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		endpointSecret := stripeWebhookSigningSecret
		event, err := webhook.ConstructEvent(body, r.Header.Get(stripeSignature), endpointSecret)

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

			err = s.ConfirmStripePaymentIntent(paymentIntent)
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
			_, err = s.FailStripePaymentIntent(paymentIntent.ID)
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
