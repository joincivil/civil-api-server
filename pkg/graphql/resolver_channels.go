package graphql

import (
	"context"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/go-common/pkg/newsroom"
)

func (r *mutationResolver) ChannelsCreateNewsroomChannel(ctx context.Context, contractAddress string) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	return r.channelService.CreateNewsroomChannel(token.Sub, channels.CreateNewsroomChannelInput{
		ContractAddress: contractAddress,
	})
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

	return &newsroom.Newsroom{
		ContractAddress: channel.Reference,
	}, nil
}
