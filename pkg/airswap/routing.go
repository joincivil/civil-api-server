package airswap

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

// defaultKrakenPollFreqSecs is the how often the price will get updated in seconds
const defaultKrakenPollFreqSecs = 30

// EnableAirswapRouting adds Airswap routes
func EnableAirswapRouting(router chi.Router) {

	// TODO(dankins): finalize these parameters, maybe via configuration?
	totalOffering := 34000000.0
	totalRaiseUSD := 20000000.0
	startingPrice := 0.2
	pricingManager := NewPricingManager(totalOffering, totalRaiseUSD, startingPrice)
	pairPricing := NewKrakenPairPricing(defaultKrakenPollFreqSecs)

	handlers := &Handlers{Pricing: pricingManager, Conversion: pairPricing}

	router.Group(func(r chi.Router) {
		r.Use(InternalOnlyMiddleware())
		r.Post("/airswap", handlers.GetOrder)
	})
}

// InternalOnlyMiddleware returns 403 Forbidden if it receives traffic from the Load Balancer
func InternalOnlyMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			header := r.Header.Get("X-Forwarded-For")

			if header != "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				respBody := fmt.Sprintf("forbidden")
				_, _ = w.Write([]byte(respBody)) // nolint: gosec
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
