package nrsignup

import (
	"net/http"
	"strings"

	log "github.com/golang/glog"

	"github.com/go-chi/chi"

	"github.com/joincivil/civil-api-server/pkg/auth"
)

// NewsroomSignupApproveGrantConfig is a struct to pass configuration to
// NewsroomSignupApproveGrantHandler
type NewsroomSignupApproveGrantConfig struct {
	NrsignupService         *Service
	TokenGenerator          *auth.JwtTokenGenerator
	RedirectErrorURL        string
	RedirectInvalidTokenURL string
	RedirectApproveURL      string
	RedirectRejectURL       string
}

// NewsroomSignupApproveGrantHandler is a REST endpoint handler for approving grants
// by the Civil Council/Foundation. Used to embed links into emails before a Foundation
// UI is created.
func NewsroomSignupApproveGrantHandler(config *NewsroomSignupApproveGrantConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "t")

		// Validate the token
		claims, err := config.TokenGenerator.ValidateToken(token)
		if err != nil {
			http.Redirect(w, r, config.RedirectInvalidTokenURL, 301)
			return
		}

		// Extract the newsroom owner UID and verify the approval value from the token
		newsroomOwnerUIDAndApproved, ok := claims["sub"]
		if !ok {
			log.Errorf("No sub claim in token")
			http.Redirect(w, r, config.RedirectInvalidTokenURL, 301)
			return
		}

		splits := strings.Split(newsroomOwnerUIDAndApproved.(string), ":")
		if len(splits) != 2 {
			log.Errorf(
				"No approved value in sub value: %v",
				newsroomOwnerUIDAndApproved.(string),
			)
			http.Redirect(w, r, config.RedirectInvalidTokenURL, 301)
			return
		}

		newsroomOwnerUID := splits[0]
		approvedVal := splits[1]

		var approvedErr error
		var redirectDestURL string

		// Based on approval value, call ApproveGrant with approved or not
		if approvedVal == ApprovedSubValue {
			approvedErr = config.NrsignupService.ApproveGrant(newsroomOwnerUID, true)
			redirectDestURL = config.RedirectApproveURL

		} else if approvedVal == RejectedSubValue {
			approvedErr = config.NrsignupService.ApproveGrant(newsroomOwnerUID, false)
			redirectDestURL = config.RedirectRejectURL

		} else {
			log.Errorf(
				"No valid approved value in sub value: %v",
				newsroomOwnerUIDAndApproved.(string),
			)
			http.Redirect(w, r, config.RedirectInvalidTokenURL, 301)
			return
		}

		if approvedErr != nil {
			log.Errorf(
				"No valid approved value in sub value: %v",
				newsroomOwnerUIDAndApproved.(string),
			)
			http.Redirect(w, r, config.RedirectErrorURL, 301)
			return
		}

		http.Redirect(w, r, redirectDestURL, 301)
	}
}
