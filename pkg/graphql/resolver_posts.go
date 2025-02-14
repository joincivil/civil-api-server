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
func (r *queryResolver) PostsGetChildren(ctx context.Context, id string, first *int, after *string) (*graphql.PostResultCursor, error) {
	return children(ctx, r.postService, id, first, after)
}

func (r *queryResolver) PostsGetByReference(ctx context.Context, reference string) (posts.Post, error) {
	return r.postService.GetPostByReferenceSafe(reference)
}

func (r *queryResolver) PostsSearch(ctx context.Context, input posts.SearchInput) (*posts.PostSearchResult, error) {

	results, err := r.postService.SearchPosts(&input)

	return results, err
}

func (r *queryResolver) PostsSearchGroupedByChannel(ctx context.Context, input posts.SearchInput) (*posts.PostSearchResult, error) {

	results, err := r.postService.SearchPostsMostRecentPerChannel(&input)

	return results, err
}

func (r *queryResolver) PostsStoryfeed(ctx context.Context, first *int, after *string, filter *posts.StoryfeedFilter) (*graphql.PostResultCursor, error) {

	cursor := defaultPaginationCursor
	var offset int
	var err error
	count := criteriaCount(first)
	// Figure out the pagination index start point if given
	if after != nil && *after != "" {
		offset, cursor, err = paginationOffsetFromCursor(cursor, after)
		if err != nil {
			return nil, err
		}
	}

	results, err := r.postService.SearchPostsRanked(count, offset, filter)
	if err != nil {
		return nil, err
	}

	posts, hasNextPage := postsReturnPosts(results.Posts, count)

	edges := postsBuildEdges(posts, cursor)
	endCursor := postsEndCursor(edges)

	return &graphql.PostResultCursor{
		Edges: edges,
		PageInfo: &graphql.PageInfo{
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		},
	}, err
}

func postsReturnPosts(allPosts []posts.Post,
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

func postsBuildEdges(posts []posts.Post,
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

func postsEndCursor(edges []*graphql.PostEdge) *string {
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
			return nil, ErrNoListingFoundForURL
		}
		channel, err := r.channelService.GetChannelByReference("newsroom", listing.ContractAddress().Hex())
		if err != nil {
			return nil, ErrNoChannelFoundForNewsroom
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
	ErrNotImplemented            = errors.New("field not yet implemented")
	ErrEmptyURLSubmitted         = errors.New("empty url submitted")
	ErrNoListingFoundForURL      = errors.New("no listing found associated with submitted url")
	ErrNoChannelFoundForNewsroom = errors.New("no channel found for listing")
)

// TYPE RESOLVERS
type postResolver struct{ *Resolver }

func (r *postResolver) getChannel(ctx context.Context, post posts.Post) (*channels.Channel, error) {
	return r.channelService.GetChannel(post.GetChannelID())
}

func children(ctx context.Context, postService *posts.Service, postID string, first *int, after *string) (*graphql.PostResultCursor, error) {
	cursor := defaultPaginationCursor
	var offset int
	var err error
	count := criteriaCount(first)
	// Figure out the pagination index start point if given
	if after != nil && *after != "" {
		offset, cursor, err = paginationOffsetFromCursor(cursor, after)
		if err != nil {
			return nil, err
		}
	}

	results, err := postService.SearchChildren(postID, count, offset)
	if err != nil {
		return nil, err
	}

	posts, hasNextPage := postsReturnPosts(results.Posts, count)

	edges := postsBuildEdges(posts, cursor)
	endCursor := postsEndCursor(edges)

	return &graphql.PostResultCursor{
		Edges: edges,
		PageInfo: &graphql.PageInfo{
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		},
	}, err
}

type postBoostResolver struct {
	*Resolver
	*postResolver
}

// NumChildren returns the number of children post of a Boost post
func (r *postBoostResolver) NumChildren(ctx context.Context, post *posts.Boost) (int, error) {
	return r.postService.GetNumChildrenOfPost(post.ID)
}

// Children returns children post of a Boost post
func (r *postBoostResolver) Children(ctx context.Context, post *posts.Boost, first *int, after *string) (*graphql.PostResultCursor, error) {
	if first == nil {
		three := 3
		first = &three
	}
	return children(ctx, r.postService, post.ID, first, after)
}

// Channel returns children post of a Boost post
func (r *postBoostResolver) Channel(ctx context.Context, post *posts.Boost) (*channels.Channel, error) {
	return r.getChannel(ctx, post)
}

// Payments returns payments associated with this Post
func (r *postBoostResolver) Payments(ctx context.Context, boost *posts.Boost) ([]payments.Payment, error) {
	isAdmin := r.isPostChannelAdmin(ctx, boost.ChannelID)
	if !isAdmin {
		return nil, ErrUserNotAuthorized
	}
	return r.paymentService.GetPayments(boost.ID)
}

// GroupedSanitizedPayments returns "sanitized payments" associated with this Post, grouped by channel
func (r *postBoostResolver) GroupedSanitizedPayments(ctx context.Context, boost *posts.Boost) ([]*payments.SanitizedPayment, error) {
	return r.paymentService.GetGroupedSanitizedPayments(boost.ID)
}

// PaymentsTotal is the sum if payments for this Post
func (r *postBoostResolver) PaymentsTotal(ctx context.Context, boost *posts.Boost, currencyCode string) (float64, error) {
	return r.paymentService.TotalPayments(boost.ID, currencyCode)
}

type postExternalLinkResolver struct {
	*Resolver
	*postResolver
}

// Channel returns children post of a ExternalLink post
func (r *postExternalLinkResolver) Channel(ctx context.Context, post *posts.ExternalLink) (*channels.Channel, error) {
	return r.getChannel(ctx, post)
}

// NumChildren returns the number of children post of a ExternalLink post
func (r *postExternalLinkResolver) NumChildren(ctx context.Context, post *posts.ExternalLink) (int, error) {
	return r.postService.GetNumChildrenOfPost(post.ID)
}

// Children returns children post of an ExternalLink post
func (r *postExternalLinkResolver) Children(ctx context.Context, post *posts.ExternalLink, first *int, after *string) (*graphql.PostResultCursor, error) {
	if first == nil {
		three := 3
		first = &three
	}
	return children(ctx, r.postService, post.ID, first, after)
}

// Payments returns payments associated with this Post
func (r *postExternalLinkResolver) Payments(ctx context.Context, post *posts.ExternalLink) ([]payments.Payment, error) {
	isAdmin := r.isPostChannelAdmin(ctx, post.ChannelID)
	if !isAdmin {
		return nil, ErrUserNotAuthorized
	}
	return r.paymentService.GetPayments(post.ID)
}

// GroupedSanitizedPayments returns "cleaned payments" associated with this Post
func (r *postExternalLinkResolver) GroupedSanitizedPayments(ctx context.Context, post *posts.ExternalLink) ([]*payments.SanitizedPayment, error) {
	return r.paymentService.GetGroupedSanitizedPayments(post.ID)
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

// NumChildren returns the number of children post of a Comment post
func (r *postCommentResolver) NumChildren(ctx context.Context, post *posts.Comment) (int, error) {
	return r.postService.GetNumChildrenOfPost(post.ID)
}

// Children returns children post of a Comment post
func (r *postCommentResolver) Children(ctx context.Context, post *posts.Comment, first *int, after *string) (*graphql.PostResultCursor, error) {
	if first == nil {
		three := 3
		first = &three
	}
	return children(ctx, r.postService, post.ID, first, after)
}

// Payments returns payments associated with this Post
func (r *postCommentResolver) Payments(ctx context.Context, post *posts.Comment) ([]payments.Payment, error) {
	isAdmin := r.isPostChannelAdmin(ctx, post.ChannelID)
	if !isAdmin {
		return nil, ErrUserNotAuthorized
	}
	return r.paymentService.GetPayments(post.ID)
}

// GroupedSanitizedPayments returns "sanitized payments" associated with this Post, grouped by channel
func (r *postCommentResolver) GroupedSanitizedPayments(ctx context.Context, boost *posts.Comment) ([]*payments.SanitizedPayment, error) {
	return nil, nil
}

// PaymentsTotal is the sum if payments for this Post
func (r *postCommentResolver) PaymentsTotal(ctx context.Context, comment *posts.Comment, currencyCode string) (float64, error) {
	return r.paymentService.TotalPayments(comment.ID, currencyCode)
}

// SanitizedPayment is a custom resolver for SanitizedPayments (so can get payer channel data)
func (r *Resolver) SanitizedPayment() graphql.SanitizedPaymentResolver {
	return &sanitizedPaymentResolver{Resolver: r}
}

type sanitizedPaymentResolver struct{ *Resolver }

// PayerChannel gets the channel associated with a sanitized payment
func (r *sanitizedPaymentResolver) PayerChannel(ctx context.Context, payment *payments.SanitizedPayment) (*channels.Channel, error) {
	return r.channelService.GetChannel(payment.PayerChannelID)
}

func (r *Resolver) isPostChannelAdmin(ctx context.Context, channelID string) bool {
	token := auth.ForContext(ctx)
	if token == nil {
		return false
	}

	isAdmin, err := r.channelService.IsChannelAdmin(token.Sub, channelID)
	if err != nil {
		return false
	}
	return isAdmin
}
