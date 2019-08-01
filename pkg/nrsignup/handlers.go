package nrsignup

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/golang/glog"

	"github.com/go-chi/chi"

	"github.com/joincivil/civil-api-server/pkg/utils"
)

const (
	// GrantApproveTokenURLParam is the name of the URL param to embed into
	// the URI
	GrantApproveTokenURLParam = "grantApproveToken"

	invalidTokenResponse = "Hmm, sorry, but the token is invalid. Could be expired, missing data or tampered with." // nolint: gosec
	errorResponse        = "Uh oh, there appears to be some internal problem, contact the devs."
	okResponse           = "Thanks, it worked! Grant approval was successfully set to %v."
)

// NewsroomSignupApproveGrantConfig is a struct to pass configuration to
// NewsroomSignupApproveGrantHandler
type NewsroomSignupApproveGrantConfig struct {
	NrsignupService *Service
	TokenGenerator  *utils.JwtTokenGenerator
}

// NewsroomSignupApproveGrantHandler is a REST endpoint handler for approving grants
// by the Civil Council/Foundation. Used to embed links into emails before a Foundation
// UI is created.
func NewsroomSignupApproveGrantHandler(config *NewsroomSignupApproveGrantConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token := chi.URLParam(r, GrantApproveTokenURLParam)

		status, approved := HandleGrantApproval(
			token,
			config.TokenGenerator,
			config.NrsignupService,
		)

		// Invalid/error responses
		if status == GrantApprovalStatusInvalidToken {
			w.Write([]byte(invalidTokenResponse)) // nolint: errcheck
			return
		} else if status == GrantApprovalStatusError {
			w.Write([]byte(errorResponse)) // nolint: errcheck
			return
		}

		// OK response
		w.Write([]byte(fmt.Sprintf(okResponse, approved))) // nolint: errcheck
	}
}

// GrantApprovalStatus is a status enum that is returned by the HandleGrantApproval function.
type GrantApprovalStatus int

const (
	// GrantApprovalStatusOK means the transaction was successful.
	GrantApprovalStatusOK GrantApprovalStatus = iota
	// GrantApprovalStatusInvalidToken means the token was invalid or missing data
	GrantApprovalStatusInvalidToken
	// GrantApprovalStatusError means there was an error
	GrantApprovalStatusError
)

// HandleGrantApproval handles the business logic for setting the approval status of
// a grant.
func HandleGrantApproval(token string, tokenGenerator *utils.JwtTokenGenerator,
	nrsignupService *Service) (GrantApprovalStatus, bool) {
	// Validate the token
	claims, err := tokenGenerator.ValidateToken(token)
	if err != nil {
		return GrantApprovalStatusInvalidToken, false
	}

	// Extract the newsroom owner UID and verify the approval value from the token
	newsroomOwnerUIDAndApproved, ok := claims["sub"]
	if !ok {
		log.Errorf("No sub claim in token")
		return GrantApprovalStatusInvalidToken, false
	}

	splits := strings.Split(newsroomOwnerUIDAndApproved.(string), ":")
	if len(splits) != 2 {
		log.Errorf(
			"No approved value in sub value: %v",
			newsroomOwnerUIDAndApproved.(string),
		)
		return GrantApprovalStatusInvalidToken, false
	}

	newsroomOwnerUID := splits[0]
	approvedVal := splits[1]

	var approvedErr error
	var approved bool

	if approvedVal == ApprovedSubValue {
		approved = true

	} else if approvedVal != RejectedSubValue {
		log.Errorf(
			"No valid approved value in sub value: %v",
			newsroomOwnerUIDAndApproved.(string),
		)
		return GrantApprovalStatusInvalidToken, false
	}

	// Approve/reject the grant
	approvedErr = nrsignupService.ApproveGrant(newsroomOwnerUID, approved)
	if approvedErr != nil {
		log.Errorf("Error approving/rejecting the grant: err: %v", approvedErr)
		return GrantApprovalStatusError, false
	}

	return GrantApprovalStatusOK, approved
}
