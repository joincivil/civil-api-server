package airswap

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/joincivil/civil-api-server/pkg/storefront"
)

// EnableAirswapRouting adds Airswap routes
func EnableAirswapRouting(router chi.Router, storefrontService *storefront.Service) {

	handlers := &Handlers{StorefrontService: storefrontService}

	router.Group(func(r chi.Router) {
		r.Use(InternalOnlyMiddleware())
		r.Post("/airswap", handlers.GetOrder)
	})

}

// InternalOnlyMiddleware returns 403 Forbidden if it receives traffic from the Load Balancer
// todo(dankins): this should be moved to go-common
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
