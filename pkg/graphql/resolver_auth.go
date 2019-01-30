package graphql

import (
	context "context"
	"fmt"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/users"
)

func (r *mutationResolver) AuthSignupEth(ctx context.Context, input users.SignatureInput) (*auth.LoginResponse, error) {
	response, err := r.authService.SignupEth(&input)
	if err != nil {
		if err == fmt.Errorf("User already exists with this address") {
			return nil, err
		}
		return nil, err
	}

	return response, nil
}

func (r *mutationResolver) AuthSignupEmailSend(ctx context.Context, emailAddress string) (*string, error) {
	result, _, err := r.authService.SignupEmailSend(emailAddress)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *mutationResolver) AuthSignupEmailSendForApplication(ctx context.Context, emailAddress string,
	application auth.ApplicationEnum) (*string, error) {
	result, _, err := r.authService.SignupEmailSendForApplication(emailAddress, application)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *mutationResolver) AuthSignupEmailConfirm(ctx context.Context, signupJWT string) (*auth.LoginResponse, error) {
	response, err := r.authService.SignupEmailConfirm(signupJWT)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (r *mutationResolver) AuthLoginEth(ctx context.Context, input users.SignatureInput) (*auth.LoginResponse, error) {
	response, err := r.authService.LoginEth(&input)
	if err != nil {
		return nil, fmt.Errorf("signature invalid or not signed up")
	}

	return response, nil
}
func (r *mutationResolver) AuthLoginEmailSend(ctx context.Context, emailAddress string) (*string, error) {
	result, _, err := r.authService.LoginEmailSend(emailAddress)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *mutationResolver) AuthLoginEmailSendForApplication(ctx context.Context, emailAddress string,
	application auth.ApplicationEnum) (*string, error) {
	result, _, err := r.authService.LoginEmailSendForApplication(emailAddress, application)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *mutationResolver) AuthLoginEmailConfirm(ctx context.Context, loginJWT string) (*auth.LoginResponse, error) {
	response, err := r.authService.LoginEmailConfirm(loginJWT)
	if err != nil {
		return nil, err
	}

	return response, nil
}
