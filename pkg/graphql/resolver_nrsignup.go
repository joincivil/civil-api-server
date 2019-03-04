package graphql

import (
	context "context"
	"github.com/joincivil/civil-api-server/pkg/auth"
	model "github.com/joincivil/civil-api-server/pkg/nrsignup"
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

func (r *mutationResolver) NrsignupSaveCharter(ctx context.Context, charterData model.Charter) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}

	err := r.nrsignupService.UpdateCharter(token.Sub, charterData)
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

func (r *queryResolver) NrsignupNewsroom(ctx context.Context) (*model.SignupUserJSONData, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	newsroom, err := r.nrsignupService.RetrieveUserJSONData(token.Sub)

	if err != nil {
		return nil, err
	}

	return newsroom, nil
}
