package graphqlmain

import (
	"fmt"
	"net/http"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/utils"
)

func healthCheckRouting(router chi.Router) error {
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK")) // nolint: errcheck
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
