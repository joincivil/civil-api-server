package graphql

import (
	context "context"
	"errors"
	"fmt"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/go-common/pkg/email"
)

// MUTATIONS
func (r *mutationResolver) PaymentsCreateEtherPayment(ctx context.Context, postID string, payment payments.EtherPayment) (payments.EtherPayment, error) {

	post, err := r.postService.GetPost(postID)
	if err != nil {
		return payments.EtherPayment{}, errors.New("could not find post")
	}

	channelID := post.GetChannelID()
	tmplData, err2 := r.GetEthPaymentEmailTemplateData(post, payment)
	fmt.Println("tmplData: ", tmplData)
	fmt.Println("err2: ", err2)
	return r.paymentService.CreateEtherPayment(channelID, "posts", postID, payment.TransactionID, payment.EmailAddress)
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

func (r *mutationResolver) GetEthPaymentEmailTemplateData(post posts.Post, payment payments.EtherPayment) (email.TemplateData, error) {
	if post.GetType() == "boost" {
		boost := post.(*posts.Boost)
		fmt.Println("payment: ", payment)
		fmt.Println("payment.Data: ", payment.Data)
		//data :=
		return (email.TemplateData{
			"newsroom_name":        "temp-name",
			"boost_short_desc":     boost.Title,
			"payment_amount_eth":   "iunno",
			"payment_amount_usd":   payment.USDEquivalent(),
			"payment_from_address": "temp-address-1",
			"payment_to_address":   "temp-address-2",
		}), nil
	} else {
		return nil, ErrNotImplemented
	}
}

func (r *mutationResolver) TestLogs(ctx context.Context, postID string, payment payments.EtherPayment) (payments.EtherPayment, error) {

	tmplData, err2 := r.GetEthPaymentEmailTemplateData(post, payment)
	fmt.Println("tmplData: ", tmplData)
	fmt.Println("err2: ", err2)
	return payments.EtherPayment{}, ErrNotImplemented
}
