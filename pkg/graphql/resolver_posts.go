package graphql

import (
	context "context"
	"errors"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
)

// PostBoost is the resolver for the PostBoost type
func (r *Resolver) PostBoost() graphql.PostBoostResolver {
	return &postBoostResolver{Resolver: r, postResolver: &postResolver{r}}
}

// PostExternalLink is the resolver for the PostExternalLink type
func (r *Resolver) PostExternalLink() graphql.PostExternalLinkResolver {
	return &postExternalLinkResolver{Resolver: r, postResolver: &postResolver{r}}
}

// PostComment is the resolver for the PostComment type
func (r *Resolver) PostComment() graphql.PostCommentResolver {
	return &postCommentResolver{Resolver: r, postResolver: &postResolver{r}}
}

// QUERIES

func (r *queryResolver) PostsGet(ctx context.Context, id string) (posts.Post, error) {
	return r.postService.GetPost(id)
}

func (r *queryResolver) PostsSearch(ctx context.Context, input posts.SearchInput) (*posts.PostSearchResult, error) {

	results, err := r.postService.SearchPosts(&input)

	return results, err
}

// MUTATIONS
func (r *mutationResolver) postCreate(ctx context.Context, post posts.Post) (posts.Post, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	isAdmin, err := r.channelService.IsChannelAdmin(token.Sub, post.GetChannelID())
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, ErrAccessDenied
	}

	result, err := r.postService.CreatePost(token.Sub, post)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *mutationResolver) postUpdate(ctx context.Context, postID string, input posts.Post) (posts.Post, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	result, err := r.postService.EditPost(token.Sub, postID, input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *mutationResolver) PostsCreateBoost(ctx context.Context, input posts.Boost) (*posts.Boost, error) {
	post, err := r.postCreate(ctx, input)
	if err != nil {
		return nil, err
	}

	boost := post.(posts.Boost)

	return &boost, nil
}

func (r *mutationResolver) PostsUpdateBoost(ctx context.Context, postID string, input posts.Boost) (*posts.Boost, error) {
	post, err := r.postUpdate(ctx, postID, input)
	if err != nil {
		return nil, err
	}

	boost := post.(posts.Boost)

	return &boost, nil
}

func (r *mutationResolver) PostsCreateComment(ctx context.Context, input posts.Comment) (*posts.Comment, error) {
	post, err := r.postCreate(ctx, input)
	if err != nil {
		return nil, err
	}

	return post.(*posts.Comment), nil
}

func (r *mutationResolver) PostsUpdateComment(ctx context.Context, postID string, input posts.Comment) (*posts.Comment, error) {
	post, err := r.postUpdate(ctx, postID, input)
	if err != nil {
		return nil, err
	}

	return post.(*posts.Comment), nil
}

func (r *mutationResolver) PostsCreateExternalLink(ctx context.Context, input posts.ExternalLink) (*posts.ExternalLink, error) {
	post, err := r.postCreate(ctx, input)
	if err != nil {
		return nil, err
	}

	return post.(*posts.ExternalLink), nil
}

func (r *mutationResolver) PostsUpdateExternalLink(ctx context.Context, postID string, input posts.ExternalLink) (*posts.ExternalLink, error) {
	post, err := r.postUpdate(ctx, postID, input)
	if err != nil {
		return nil, err
	}

	return post.(*posts.ExternalLink), nil
}

// errors
var (
	ErrNotImplemented = errors.New("field not yet implemented")
)

// TYPE RESOLVERS
type postResolver struct{ *Resolver }

func (r *postResolver) getChannel(ctx context.Context, post posts.Post) (*channels.Channel, error) {
	return r.channelService.GetChannel(post.GetChannelID())
}

type postBoostResolver struct {
	*Resolver
	*postResolver
}

// Children returns children post of a Boost post
func (r *postBoostResolver) Children(context.Context, *posts.Boost) ([]posts.Post, error) {
	return nil, ErrNotImplemented
}

// Channel returns children post of a Boost post
func (r *postBoostResolver) Channel(ctx context.Context, post *posts.Boost) (*channels.Channel, error) {
	return r.getChannel(ctx, post)
}

// Payments returns payments associated with this Post
func (r *postBoostResolver) Payments(ctx context.Context, boost *posts.Boost) ([]payments.Payment, error) {
	return r.paymentService.GetPayments(boost.ID)
}

// PaymentsTotal is the sum if payments for this Post
func (r *postBoostResolver) PaymentsTotal(ctx context.Context, boost *posts.Boost, currencyCode string) (float64, error) {
	return r.paymentService.TotalPayments(boost.ID, currencyCode)
}

type postExternalLinkResolver struct {
	*Resolver
	*postResolver
}

// Children returns children post of an ExternalLink post
func (r *postExternalLinkResolver) Children(context.Context, *posts.ExternalLink) ([]posts.Post, error) {
	return nil, ErrNotImplemented
}

// Channel returns children post of a ExternalLink post
func (r *postExternalLinkResolver) Channel(ctx context.Context, post *posts.ExternalLink) (*channels.Channel, error) {
	return r.getChannel(ctx, post)
}

// Payments returns payments associated with this Post
func (r *postExternalLinkResolver) Payments(ctx context.Context, post *posts.ExternalLink) ([]payments.Payment, error) {
	return r.paymentService.GetPayments(post.ID)
}

// PaymentsTotal is the sum if payments for this Post
func (r *postExternalLinkResolver) PaymentsTotal(ctx context.Context, link *posts.ExternalLink, currencyCode string) (float64, error) {
	return r.paymentService.TotalPayments(link.ID, currencyCode)
}

type postCommentResolver struct {
	*Resolver
	*postResolver
}

// Channel returns children post of a ExternalLink post
func (r *postCommentResolver) Channel(ctx context.Context, post *posts.Comment) (*channels.Channel, error) {
	return r.getChannel(ctx, post)
}

// Children returns children post of an Comment post
func (r *postCommentResolver) Children(context.Context, *posts.Comment) ([]posts.Post, error) {
	return nil, ErrNotImplemented
}

// Payments returns payments associated with this Post
func (r *postCommentResolver) Payments(ctx context.Context, post *posts.Comment) ([]payments.Payment, error) {
	return r.paymentService.GetPayments(post.ID)
}

// PaymentsTotal is the sum if payments for this Post
func (r *postCommentResolver) PaymentsTotal(ctx context.Context, comment *posts.Comment, currencyCode string) (float64, error) {
	return r.paymentService.TotalPayments(comment.ID, currencyCode)
}
