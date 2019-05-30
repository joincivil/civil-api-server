package graphql

import (
	context "context"
	"errors"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/posts"
)

// PostBoost is the resolver for the PostBoost type
func (r *Resolver) PostBoost() graphql.PostBoostResolver {
	return &postBoostResolver{r}
}

// PostExternalLink is the resolver for the PostExternalLink type
func (r *Resolver) PostExternalLink() graphql.PostExternalLinkResolver {
	return &postExternalLinkResolver{r}
}

// PostComment is the resolver for the PostComment type
func (r *Resolver) PostComment() graphql.PostCommentResolver {
	return &postCommentResolver{r}
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

func (r *mutationResolver) PostsCreateBoost(ctx context.Context, boost posts.Boost) (*posts.Boost, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	boost.AuthorID = token.Sub
	result, err := r.postService.CreatePost(&boost)
	if err != nil {
		return nil, err
	}

	return result.(*posts.Boost), nil
}

func (r *mutationResolver) PostsCreateComment(ctx context.Context, comment posts.Comment) (*posts.Comment, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	comment.AuthorID = token.Sub
	result, err := r.postService.CreatePost(&comment)
	if err != nil {
		return nil, err
	}

	return result.(*posts.Comment), nil
}
func (r *mutationResolver) PostsCreateExternalLink(ctx context.Context, linkPost posts.ExternalLink) (*posts.ExternalLink, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	linkPost.AuthorID = token.Sub
	result, err := r.postService.CreatePost(&linkPost)
	if err != nil {
		return nil, err
	}

	return result.(*posts.ExternalLink), nil
}

// errors
var (
	ErrNotImplemented = errors.New("field not yet implemented")
)

// TYPE RESOLVERS

type postBoostResolver struct{ *Resolver }

// TotalPaymentsUSD returns the
func (r *postBoostResolver) TotalPaymentsUSD(ctx context.Context, boost *posts.Boost) (float64, error) {
	return r.postService.TotalPaymentsUSD(boost.ID)
}

// Children returns children post of a Boost post
func (r *postBoostResolver) Children(context.Context, *posts.Boost) ([]*posts.Post, error) {
	return nil, ErrNotImplemented
}

type postExternalLinkResolver struct{ *Resolver }

// Children returns children post of an external link post
func (r *postExternalLinkResolver) Children(context.Context, *posts.ExternalLink) ([]*posts.Post, error) {
	return nil, ErrNotImplemented
}

type postCommentResolver struct{ *Resolver }

// Children returns children post of an external link post
func (r *postCommentResolver) Children(context.Context, *posts.Comment) ([]*posts.Post, error) {
	return nil, ErrNotImplemented
}