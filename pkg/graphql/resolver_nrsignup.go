package graphql

import (
	context "context"

	"github.com/joincivil/civil-api-server/pkg/auth"
)

func (r *mutationResolver) NrsignupSendWelcomeEmail(ctx context.Context) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}

	err := r.nrsignupService.SendWelcomeEmail(token.Sub)
	if err != nil {
		return "", err
	}

	return ResponseOK, nil
}

func (r *mutationResolver) NrsignupRequestGrant(ctx context.Context) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}

	err := r.nrsignupService.RequestGrant(token.Sub)
	if err != nil {
		return "", err
	}

	return ResponseOK, nil
}

// Not implemented yet, use the REST endpoint instead.
func (r *mutationResolver) NrsignupApproveGrant(ctx context.Context, approved bool,
	newsroomOwnerUID string) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}
	return ResponseNotImplemented, nil
}

// Not implemented yet
func (r *mutationResolver) NrsignupPollNewsroomDeploy(ctx context.Context,
	txHash string) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}
	return ResponseNotImplemented, nil
}

// Not implemented yet
func (r *mutationResolver) NrsignupPollTcrApplication(ctx context.Context,
	txHash string) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}
	return ResponseNotImplemented, nil
}
