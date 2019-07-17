package graphql

import (
	context "context"
	"errors"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/payments"
)

// MUTATIONS
func (r *mutationResolver) PaymentsCreateEtherPayment(ctx context.Context, postID string, payment payments.EtherPayment) (payments.EtherPayment, error) {

	post, err := r.postService.GetPost(postID)
	if err != nil {
		return payments.EtherPayment{}, errors.New("could not find post")
	}

	channelID := post.GetChannelID()

	return r.paymentService.CreateEtherPayment(channelID, "posts", postID, payment.TransactionID)
}

func (r *mutationResolver) PaymentsCreateStripePayment(ctx context.Context, postID string, payment payments.StripePayment) (payments.StripePayment, error) {

	post, err := r.postService.GetPost(postID)
	if err != nil {
		return payments.StripePayment{}, errors.New("could not find post")
	}

	channelID := post.GetChannelID()

	return r.paymentService.CreateStripePayment(channelID, "posts", postID, payment)
}

func (r *mutationResolver) PaymentsCreateTokenPayment(ctx context.Context, postID string, payment payments.TokenPayment) (payments.TokenPayment, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return payments.TokenPayment{}, ErrAccessDenied
	}

	return payments.TokenPayment{}, ErrNotImplemented
}
