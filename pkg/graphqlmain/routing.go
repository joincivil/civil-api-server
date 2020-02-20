package graphqlmain

import (
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"net/http"
)

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
