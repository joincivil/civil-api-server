package graphql

import (
	context "context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-events-processor/pkg/utils"
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

func (r *queryResolver) PostsGetByReference(ctx context.Context, reference string) (posts.Post, error) {
	return r.postService.GetPostByReference(reference)
}

func (r *queryResolver) PostsSearch(ctx context.Context, input posts.SearchInput) (*posts.PostSearchResult, error) {

	results, err := r.postService.SearchPosts(&input)

	return results, err
}

func (r *queryResolver) PostsSearchGroupedByChannel(ctx context.Context, input posts.SearchInput) (*posts.PostSearchResult, error) {

	results, err := r.postService.SearchPostsMostRecentPerChannel(&input)

	return results, err
}

func (r *queryResolver) PostsStoryfeed(ctx context.Context, first *int, after *string) (*graphql.PostResultCursor, error) {

	cursor := defaultPaginationCursor
	var offset int
	var err error
	count := r.criteriaCount(first)
	// Figure out the pagination index start point if given
	if after != nil && *after != "" {
		offset, cursor, err = r.paginationOffsetFromCursor(cursor, after)
		if err != nil {
			return nil, err
		}
	}

	results, err := r.postService.SearchPostsRanked(count, offset)
	if err != nil {
		return nil, err
	}

	posts, hasNextPage := r.postsReturnPosts(results.Posts, count)

	edges := r.postsBuildEdges(posts, cursor)
	endCursor := r.postsEndCursor(edges)

	return &graphql.PostResultCursor{
		Edges: edges,
		PageInfo: &graphql.PageInfo{
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		},
	}, err
}

func (r *queryResolver) postsReturnPosts(allPosts []posts.Post,
	count int) ([]posts.Post, bool) {
	allPostsLen := len(allPosts)

	hasNextPage := false
	var posts []posts.Post

	// Figure out the "true" events we want to return.
	// If the posts actually equals what we requested, then we have more results
	// and hasNextPage should be true
	if allPostsLen == count {
		hasNextPage = true
		posts = allPosts[:allPostsLen-1]
	} else {
		posts = allPosts
	}
	return posts, hasNextPage
}

func (r *queryResolver) postsBuildEdges(posts []posts.Post,
	cursor *paginationCursor) []*graphql.PostEdge {

	edges := make([]*graphql.PostEdge, len(posts))

	// Build edges
	// Only support sorted offset until we need other types
	for index, post := range posts {
		cv := cursor.ValueInt()
		newCursor := &paginationCursor{
			typeName: cursor.typeName,
			value:    fmt.Sprintf("%v", cv+index),
		}
		edges[index] = &graphql.PostEdge{
			Cursor: newCursor.Encode(),
			Post:   post,
		}
	}
	return edges
}

func (r *queryResolver) postsEndCursor(edges []*graphql.PostEdge) *string {
	var endCursor *string
	if len(edges) > 0 {
		endCursor = &(edges[len(edges)-1]).Cursor
	}
	return endCursor
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

	return post.(*posts.Boost), nil
}

func (r *mutationResolver) PostsUpdateBoost(ctx context.Context, postID string, input posts.Boost) (*posts.Boost, error) {
	post, err := r.postUpdate(ctx, postID, input)
	if err != nil {
		return nil, err
	}

	return post.(*posts.Boost), nil
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

func (r *mutationResolver) PostsCreateExternalLinkEmbedded(ctx context.Context, input posts.ExternalLink) (*posts.ExternalLink, error) {
	if input.URL != "" {
		cleanedURL, err := utils.CleanURL(input.URL)
		if err != nil {
			return nil, err
		}
		listing, err := r.listingPersister.ListingByCleanedNewsroomURL(cleanedURL)
		if err != nil {
			return nil, err
		}
		channel, err := r.channelService.GetChannelByReference("newsroom", listing.ContractAddress().Hex())
		if err != nil {
			return nil, err
		}
		input.ChannelID = channel.ID
		post, err := r.postService.CreateExternalLinkEmbedded(input)
		if err != nil {
			return nil, err
		}

		return post.(*posts.ExternalLink), nil
	}
	return nil, ErrEmptyURLSubmitted
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
	ErrNotImplemented    = errors.New("field not yet implemented")
	ErrEmptyURLSubmitted = errors.New("empty url submitted")
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

// CleanedPayments returns "cleaned payments" associated with this Post
func (r *postBoostResolver) CleanedPayments(ctx context.Context, boost *posts.Boost) ([]*payments.CleanedPayment, error) {
	return r.paymentService.GetCleanedPayments(boost.ID)
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

// CleanedPayments returns "cleaned payments" associated with this Post
func (r *postExternalLinkResolver) CleanedPayments(ctx context.Context, post *posts.ExternalLink) ([]*payments.CleanedPayment, error) {
	return r.paymentService.GetCleanedPayments(post.ID)
}

// PaymentsTotal is the sum if payments for this Post
func (r *postExternalLinkResolver) PaymentsTotal(ctx context.Context, link *posts.ExternalLink, currencyCode string) (float64, error) {
	return r.paymentService.TotalPayments(link.ID, currencyCode)
}

// OpenGraphData returns the open graph data for this post
func (r *postExternalLinkResolver) OpenGraphData(ctx context.Context, link *posts.ExternalLink) (*graphql.OpenGraphData, error) {
	var ogdata graphql.OpenGraphData
	err := json.Unmarshal(link.OpenGraphData, &ogdata)
	if err != nil {
		return nil, err
	}
	return &ogdata, nil
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

// CleanedPayments returns "cleaned payments" associated with this Post
func (r *postCommentResolver) CleanedPayments(ctx context.Context, boost *posts.Comment) ([]*payments.CleanedPayment, error) {
	return nil, nil
}

// PaymentsTotal is the sum if payments for this Post
func (r *postCommentResolver) PaymentsTotal(ctx context.Context, comment *posts.Comment, currencyCode string) (float64, error) {
	return r.paymentService.TotalPayments(comment.ID, currencyCode)
}

// CleanedPayment is a custom resolver for CleanedPayments (so can get payer channel data)
func (r *Resolver) CleanedPayment() graphql.CleanedPaymentResolver {
	return &cleanedPaymentResolver{Resolver: r}
}

type cleanedPaymentResolver struct{ *Resolver }

func (r *cleanedPaymentResolver) PayerChannel(ctx context.Context, payment *payments.CleanedPayment) (*channels.Channel, error) {
	return r.channelService.GetChannel(payment.PayerChannelID)
}
