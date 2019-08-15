package graphql

import (
	context "context"
	"errors"

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
	if err2 != nil {
		return payments.EtherPayment{}, errors.New("error creating email template data")
	}
	return r.paymentService.CreateEtherPayment(channelID, "posts", postID, payment.TransactionID, payment.EmailAddress, tmplData)
}

func (r *mutationResolver) PaymentsCreateStripePayment(ctx context.Context, postID string, payment payments.StripePayment) (payments.StripePayment, error) {

	post, err := r.postService.GetPost(postID)
	if err != nil {
		return payments.StripePayment{}, errors.New("could not find post")
	}

	channelID := post.GetChannelID()
	tmplData, err2 := r.GetStripePaymentEmailTemplateData(post, payment)
	if err2 != nil {
		return payments.StripePayment{}, errors.New("error creating email template data")
	}
	return r.paymentService.CreateStripePayment(channelID, "posts", postID, payment, tmplData)
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
		channel, err := r.channelService.GetChannel(post.GetChannelID())
		if err != nil {
			return nil, errors.New("could not find channel")
		}
		newsroom, err := r.newsroomService.GetNewsroomByAddress(channel.Reference)
		if err != nil {
			return nil, errors.New("could not find newsroom")
		}
		return (email.TemplateData{
			"newsroom_name":        newsroom.Name,
			"boost_short_desc":     boost.Title,
			"payment_amount_eth":   payment.Amount,
			"payment_amount_usd":   payment.UsdAmount,
			"payment_from_address": payment.FromAddress,
			"payment_to_address":   payment.PaymentAddress,
			"boost_id":             boost.ID,
		}), nil
	}
	return nil, ErrNotImplemented
}

func (r *mutationResolver) GetStripePaymentEmailTemplateData(post posts.Post, payment payments.StripePayment) (email.TemplateData, error) {
	if post.GetType() == "boost" {
		boost := post.(*posts.Boost)
		channel, err := r.channelService.GetChannel(post.GetChannelID())
		if err != nil {
			return nil, errors.New("could not find channel")
		}
		newsroom, err := r.newsroomService.GetNewsroomByAddress(channel.Reference)
		if err != nil {
			return nil, errors.New("could not find newsroom")
		}
		return (email.TemplateData{
			"newsroom_name":      newsroom.Name,
			"boost_short_desc":   boost.Title,
			"payment_amount_usd": payment.Amount,
			"boost_id":           boost.ID,
		}), nil
	}
	return nil, ErrNotImplemented
}

func (r *queryResolver) GetChannelTotalProceeds(ctx context.Context, channelID string) (*payments.ProceedsQueryResult, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}
	isAdmin, err := r.channelService.IsChannelAdmin(token.Sub, channelID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, ErrAccessDenied
	}

	result := r.paymentService.GetChannelTotalProceeds(channelID)
	return result, nil
}
