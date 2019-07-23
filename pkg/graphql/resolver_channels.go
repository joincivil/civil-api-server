package graphql

import (
	"context"
	"github.com/joincivil/civil-api-server/pkg/posts"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/go-common/pkg/newsroom"
)

// queries
func (r *queryResolver) ChannelsGetByID(ctx context.Context, id string) (*channels.Channel, error) {
	return r.channelService.GetChannel(id)
}
func (r *queryResolver) ChannelsGetByNewsroomAddress(ctx context.Context, contractAddress string) (*channels.Channel, error) {
	return r.channelService.GetChannelByReference("newsroom", contractAddress)
}

func (r *queryResolver) ChannelsGetByHandle(ctx context.Context, handle string) (*channels.Channel, error) {
	return r.channelService.GetChannelByHandle(handle)
}

// mutations
func (r *mutationResolver) ChannelsCreateNewsroomChannel(ctx context.Context, contractAddress string) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	return r.channelService.CreateNewsroomChannel(token.Sub, channels.CreateNewsroomChannelInput{
		ContractAddress: contractAddress,
	})
}

func (r *mutationResolver) ChannelsConnectStripe(ctx context.Context, input channels.ConnectStripeInput) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	return r.channelService.ConnectStripe(token.Sub, input)
}

// Channel is the resolver for the Channel type
func (r *Resolver) Channel() graphql.ChannelResolver {
	return &channelResolver{Resolver: r}
}

type channelResolver struct {
	*Resolver
}

// Newsroom returns newsroom associated with this channel
func (r *channelResolver) Newsroom(ctx context.Context, channel *channels.Channel) (*newsroom.Newsroom, error) {
	if channel.ChannelType != channels.TypeNewsroom {
		return nil, nil
	}

	return r.newsroomService.GetNewsroomByAddress(channel.Reference)
}

func (r *channelResolver) PostsSearch(ctx context.Context, channel *channels.Channel, input posts.SearchInput) (*posts.PostSearchResult, error) {

	input.ChannelID = channel.ID
	results, err := r.postService.SearchPosts(&input)

	return results, err
}

func (r *channelResolver) IsStripeConnected(ctx context.Context, channel *channels.Channel) (bool, error) {
	ch, err := r.channelService.GetChannel(channel.ID)
	if err != nil {
		return false, err
	}

	return ch.StripeAccountID != "", err
}

func (r *channelResolver) CurrentUserIsAdmin(ctx context.Context, channel *channels.Channel) (bool, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return false, nil
	}

	return r.channelService.IsChannelAdmin(token.Sub, channel.ID)
}
