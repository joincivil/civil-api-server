package graphql

import (
	context "context"
	"fmt"

	log "github.com/golang/glog"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
)

const (
	nrsignupOKResponse     = "ok"
	nrsignupNotImplemented = "not implemented"
)

func (r *mutationResolver) NrsignupSendWelcomeEmail(ctx context.Context) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", fmt.Errorf("Access denied")
	}
	log.Infof("token sub = %v", token.Sub)

	err := r.nrsignupService.SendWelcomeEmail(token.Sub)
	if err != nil {
		return "", err
	}

	return nrsignupOKResponse, nil
}

func (r *mutationResolver) NrsignupRequestGrant(ctx context.Context,
	newsroomData nrsignup.Newsroom) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", fmt.Errorf("Access denied")
	}

	err := r.nrsignupService.RequestGrant(token.Sub, &newsroomData)
	if err != nil {
		return "", err
	}

	return nrsignupOKResponse, nil
}

// Not implemented yet, use the REST endpoint instead.
func (r *mutationResolver) NrsignupApproveGrant(ctx context.Context, approved bool,
	newsroomOwnerUID string) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", fmt.Errorf("Access denied")
	}
	return nrsignupNotImplemented, nil
}

func (r *mutationResolver) NrsignupPollNewsroomDeploy(ctx context.Context,
	txHash string) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", fmt.Errorf("Access denied")
	}
	return "", nil
}

func (r *mutationResolver) NrsignupPollTcrApplication(ctx context.Context,
	txHash string) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", fmt.Errorf("Access denied")
	}
	return "", nil
}
