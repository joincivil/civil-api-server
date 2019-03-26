package graphql

import (
	context "context"
	"fmt"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	cemail "github.com/joincivil/go-common/pkg/email"
)

const (
	// emailTagNewsroomSignup is the email tag to indicate the user is signed up
	// from the newsroom signup module
	emailTagNewsroomSignup cemail.Tag = "Newsroom Signup"

	// emailTagTokenStorefront is the email tag to indicate the user is signed up
	// from the token storefront module
	emailTagTokenStorefront = "Token Storefront"
)

func (r *mutationResolver) addToNewsletterList(emailAddress string,
	application auth.ApplicationEnum) error {
	if r.emailListMembers != nil {
		tags := []cemail.Tag{}

		if application == auth.ApplicationEnumNewsroom {
			tags = append(tags, emailTagNewsroomSignup)

		} else if application == auth.ApplicationEnumStorefront {
			tags = append(tags, emailTagTokenStorefront)
		}

		err := utils.AddToCivilCompanyNewsletterList(r.emailListMembers, emailAddress, tags)
		if err != nil {
			log.Errorf("Error adding to newsletter: err: %v", err)
			return err
		}
	}
	return nil
}

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

func (r *mutationResolver) AuthSignupEmailSend(ctx context.Context, emailAddress string,
	addToMailing *bool) (*string, error) {
	result, _, err := r.authService.SignupEmailSend(emailAddress)
	if err != nil {
		return nil, err
	}

	// if add to list is true, try to add to mailchimp list.
	if *addToMailing {
		_ = r.addToNewsletterList(emailAddress, auth.ApplicationEnumDefault)
	}

	return &result, nil
}

func (r *mutationResolver) AuthSignupEmailSendForApplication(ctx context.Context, emailAddress string,
	application auth.ApplicationEnum, addToMailing *bool) (*string, error) {
	result, _, err := r.authService.SignupEmailSendForApplication(emailAddress, application)
	if err != nil {
		return nil, err
	}

	// if add to list is true, try to add to mailchimp list.
	if *addToMailing {
		_ = r.addToNewsletterList(emailAddress, application)
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
func (r *mutationResolver) AuthLoginEmailSend(ctx context.Context, emailAddress string,
	addToMailing *bool) (*string, error) {
	result, _, err := r.authService.LoginEmailSend(emailAddress)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *mutationResolver) AuthLoginEmailSendForApplication(ctx context.Context, emailAddress string,
	application auth.ApplicationEnum, addToMailing *bool) (*string, error) {
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

func (r *mutationResolver) AuthRefresh(ctx context.Context, token string) (*auth.LoginResponse, error) {
	response, err := r.authService.RefreshAccessToken(token)
	if err != nil {
		return nil, err
	}

	return response, nil

}
