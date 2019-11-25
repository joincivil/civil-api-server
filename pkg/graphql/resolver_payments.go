package graphql

import (
	context "context"
	"errors"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/go-common/pkg/email"
)

func (r *mutationResolver) validateUserIsChannelAdmin(ctx context.Context, channelID string) error {
	token := auth.ForContext(ctx)
	if token == nil {
		return ErrAccessDenied
	}
	isAdmin, err := r.channelService.IsChannelAdmin(token.Sub, channelID)
	if err != nil || !isAdmin {
		return ErrAccessDenied
	}
	return nil
}

func (r *queryResolver) validateUserIsChannelAdmin(ctx context.Context, channelID string) error {
	token := auth.ForContext(ctx)
	if token == nil {
		return ErrAccessDenied
	}
	isAdmin, err := r.channelService.IsChannelAdmin(token.Sub, channelID)
	if err != nil || !isAdmin {
		return ErrAccessDenied
	}
	return nil
}

// MUTATIONS

// nolint: dupl
func (r *mutationResolver) PaymentsCreateEtherPayment(ctx context.Context, postID string, payment payments.EtherPayment) (*payments.EtherPayment, error) {

	post, err := r.postService.GetPost(postID)
	if err != nil {
		return &payments.EtherPayment{}, errors.New("could not find post")
	}

	if payment.PayerChannelID != "" {
		err = r.validateUserIsChannelAdmin(ctx, payment.PayerChannelID)
		if err != nil {
			return &payments.EtherPayment{}, err
		}
	}
	if payment.PayerChannelID == "" {
		payment.ShouldPublicize = false
	}

	channelID := post.GetChannelID()
	tmplData, err2 := r.GetEthPaymentEmailTemplateData(post, payment)
	if err2 != nil {
		return &payments.EtherPayment{}, errors.New("error creating email template data")
	}

	p, err := r.paymentService.CreateEtherPayment(channelID, "posts", post.GetType(), postID, payment, tmplData)
	return &p, err
}

// nolint: dupl
func (r *mutationResolver) PaymentsCreateStripePayment(ctx context.Context, postID string, payment payments.StripePayment) (*payments.StripePayment, error) {

	post, err := r.postService.GetPost(postID)
	if err != nil {
		return &payments.StripePayment{}, errors.New("could not find post")
	}

	if payment.PayerChannelID != "" {
		err = r.validateUserIsChannelAdmin(ctx, payment.PayerChannelID)
		if err != nil {
			return &payments.StripePayment{}, err
		}
	}
	if payment.PayerChannelID == "" {
		payment.ShouldPublicize = false
	}

	channelID := post.GetChannelID()
	tmplData, err2 := r.GetStripePaymentEmailTemplateData(post, payment)
	if err2 != nil {
		return &payments.StripePayment{}, errors.New("error creating email template data")
	}
	p, err := r.paymentService.CreateStripePayment(channelID, "posts", post.GetType(), postID, payment, tmplData)
	return &p, err
}

func (r *mutationResolver) PaymentsCreateTokenPayment(ctx context.Context, postID string, payment payments.TokenPayment) (*payments.TokenPayment, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return &payments.TokenPayment{}, ErrAccessDenied
	}

	return &payments.TokenPayment{}, ErrNotImplemented
}

func (r *mutationResolver) GetEthPaymentEmailTemplateData(post posts.Post, payment payments.EtherPayment) (email.TemplateData, error) {
	channel, err := r.channelService.GetChannel(post.GetChannelID())
	if err != nil {
		return nil, errors.New("could not find channel")
	}
	newsroom, err := r.newsroomService.GetNewsroomByAddress(channel.Reference)
	if err != nil {
		return nil, errors.New("could not find newsroom")
	}
	if post.GetType() == posts.TypeBoost {
		boost := post.(*posts.Boost)
		return (email.TemplateData{
			"newsroom_name":        newsroom.Name,
			"boost_short_desc":     boost.Title,
			"payment_amount_eth":   payment.Amount,
			"payment_amount_usd":   payment.UsdAmount,
			"payment_from_address": payment.FromAddress,
			"payment_to_address":   payment.PaymentAddress,
			"boost_id":             boost.ID,
		}), nil
	} else if post.GetType() == posts.TypeExternalLink {
		return (email.TemplateData{
			"newsroom_name":        newsroom.Name,
			"payment_amount_eth":   payment.Amount,
			"payment_amount_usd":   payment.UsdAmount,
			"payment_from_address": payment.FromAddress,
			"payment_to_address":   payment.PaymentAddress,
		}), nil
	}
	return nil, ErrNotImplemented
}

func (r *mutationResolver) GetStripePaymentEmailTemplateData(post posts.Post, payment payments.StripePayment) (email.TemplateData, error) {
	channel, err := r.channelService.GetChannel(post.GetChannelID())
	if err != nil {
		return nil, errors.New("could not find channel")
	}
	newsroom, err := r.newsroomService.GetNewsroomByAddress(channel.Reference)
	if err != nil {
		return nil, errors.New("could not find newsroom")
	}
	if post.GetType() == posts.TypeBoost {
		boost := post.(*posts.Boost)
		return (email.TemplateData{
			"newsroom_name":      newsroom.Name,
			"boost_short_desc":   boost.Title,
			"payment_amount_usd": payment.Amount,
			"boost_id":           boost.ID,
		}), nil
	} else if post.GetType() == posts.TypeExternalLink {
		return (email.TemplateData{
			"newsroom_name":      newsroom.Name,
			"payment_amount_usd": payment.Amount,
		}), nil
	}
	return nil, ErrNotImplemented
}

func (r *queryResolver) GetChannelTotalProceeds(ctx context.Context, channelID string) (*payments.ProceedsQueryResult, error) {
	err := r.validateUserIsChannelAdmin(ctx, channelID)
	if err != nil {
		return nil, err
	}

	result := r.paymentService.GetChannelTotalProceeds(channelID)
	return result, nil
}

func (r *queryResolver) GetChannelTotalProceedsByBoostType(ctx context.Context, channelID string, boostType string) (*payments.ProceedsQueryResult, error) {
	err := r.validateUserIsChannelAdmin(ctx, channelID)
	if err != nil {
		return nil, err
	}

	result := r.paymentService.GetChannelTotalProceedsByBoostType(channelID, boostType)
	return result, nil
}

// PaymentEther is the resolver for the PaymentEther type
func (r *Resolver) PaymentEther() graphql.PaymentEtherResolver {
	return &etherPaymentResolver{Resolver: r, paymentResolver: &paymentResolver{r}}
}

// PaymentStripe is the resolver for the PaymentStripe type
func (r *Resolver) PaymentStripe() graphql.PaymentStripeResolver {
	return &stripePaymentResolver{Resolver: r, paymentResolver: &paymentResolver{r}}
}

// PaymentToken is the resolver for the PaymentToken type
func (r *Resolver) PaymentToken() graphql.PaymentTokenResolver {
	return &tokenPaymentResolver{Resolver: r, paymentResolver: &paymentResolver{r}}
}

// TYPE RESOLVERS
type paymentResolver struct{ *Resolver }

type etherPaymentResolver struct {
	*Resolver
	*paymentResolver
}
type stripePaymentResolver struct {
	*Resolver
	*paymentResolver
}
type tokenPaymentResolver struct {
	*Resolver
	*paymentResolver
}

func (r *etherPaymentResolver) PayerChannel(ctx context.Context, payment *payments.EtherPayment) (*channels.Channel, error) {
	if !payment.ShouldPublicize {
		return nil, nil
	}
	return r.channelService.GetChannel(payment.PayerChannelID)
}

func (r *stripePaymentResolver) PayerChannel(ctx context.Context, payment *payments.StripePayment) (*channels.Channel, error) {
	if !payment.ShouldPublicize {
		return nil, nil
	}
	return r.channelService.GetChannel(payment.PayerChannelID)
}

func (r *tokenPaymentResolver) PayerChannel(ctx context.Context, payment *payments.TokenPayment) (*channels.Channel, error) {
	if !payment.ShouldPublicize {
		return nil, nil
	}
	return r.channelService.GetChannel(payment.PayerChannelID)
}
