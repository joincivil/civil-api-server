package graphql

import (
	context "context"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
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

func (r *mutationResolver) NrsignupRequestGrant(ctx context.Context, requested bool) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}

	err := r.nrsignupService.RequestGrant(token.Sub, requested)
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

func (r *mutationResolver) NrsignupUpdateSteps(ctx context.Context,
	input graphql.NrsignupStepsInput) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}

	err := r.nrsignupService.UpdateUserSteps(token.Sub, input.Step,
		input.FurthestStep, input.LastSeen)
	if err != nil {
		return "", err
	}

	return ResponseOK, nil
}

func (r *mutationResolver) NrsignupSaveTxHash(ctx context.Context, txHash string) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}

	err := r.nrsignupService.SaveNewsroomDeployTxHash(token.Sub, txHash)
	if err != nil {
		return "", err
	}

	return ResponseOK, nil
}

func (r *mutationResolver) NrsignupSaveAddress(ctx context.Context, address string) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}

	err := r.nrsignupService.SaveNewsroomAddress(token.Sub, address)
	if err != nil {
		return "", err
	}

	return ResponseOK, nil
}

func (r *mutationResolver) NrsignupSaveNewsroomApplyTxHash(ctx context.Context, txHash string) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}

	err := r.nrsignupService.SaveNewsroomApplyTxHash(token.Sub, txHash)
	if err != nil {
		return "", err
	}

	return ResponseOK, nil
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
