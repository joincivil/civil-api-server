package graphql

import (
	"context"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
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
