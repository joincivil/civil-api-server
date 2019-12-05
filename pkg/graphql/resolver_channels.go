package graphql

import (
	"context"

	"github.com/joincivil/civil-api-server/pkg/posts"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/go-common/pkg/newsroom"
)

// queries
func (r *queryResolver) ChannelsGetByID(ctx context.Context, id string) (*channels.Channel, error) {
	return r.channelService.GetChannel(id)
}
func (r *queryResolver) ChannelsGetByNewsroomAddress(ctx context.Context, contractAddress string) (*channels.Channel, error) {
	return r.channelService.GetChannelByReference("newsroom", contractAddress)
}
func (r *queryResolver) ChannelsGetByUserID(ctx context.Context, userID string) (*channels.Channel, error) {
	return r.channelService.GetChannelByReference("user", userID)
}

func (r *queryResolver) ChannelsGetByHandle(ctx context.Context, handle string) (*channels.Channel, error) {
	return r.channelService.GetChannelByHandle(handle)
}

func (r *queryResolver) ChannelsIsHandleAvailable(ctx context.Context, handle string) (bool, error) {
	channel, err := r.channelService.GetChannelByHandle(handle)
	if err == nil || channel != nil {
		return false, nil
	}

	return true, nil
}

// mutations
func (r *mutationResolver) ChannelsCreateNewsroomChannel(ctx context.Context, contractAddress string) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	userAddresses, err := r.userService.GetETHAddresses(token.Sub)
	if err != nil {
		return nil, err
	}

	return r.channelService.CreateNewsroomChannel(token.Sub, userAddresses, channels.CreateNewsroomChannelInput{
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

func (r *mutationResolver) ChannelsSetAvatar(ctx context.Context, input channels.SetAvatarInput) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	_, err := r.userService.SetHasSeenUCAvatarPrompt(token.Sub)
	if err != nil {
		return nil, err
	}

	return r.channelService.SetAvatarDataURL(token.Sub, input.ChannelID, input.AvatarDataURL)
}

func (r *mutationResolver) ChannelsSetHandle(ctx context.Context, input channels.SetHandleInput) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	return r.channelService.SetHandle(token.Sub, input.ChannelID, input.Handle)
}

func (r *mutationResolver) ChannelsSetStripeCustomerID(ctx context.Context, input channels.SetStripeCustomerIDInput) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	return r.channelService.SetStripeCustomerID(token.Sub, input.ChannelID, input.StripeCustomerID)
}

func (r *mutationResolver) ChannelsClearStripeCustomerID(ctx context.Context, channelID string) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	return r.channelService.ClearStripeCustomerID(token.Sub, channelID)
}

func (r *mutationResolver) UserChannelSetHandle(ctx context.Context, input channels.UserSetHandleInput) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	channel, err := r.channelService.GetChannelByReference("user", token.Sub)
	if err != nil {
		return nil, err
	}

	return r.channelService.SetHandle(token.Sub, channel.ID, input.Handle)
}

func (r *mutationResolver) ChannelsSetEmail(ctx context.Context, input channels.SetEmailInput) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	return r.channelService.SendEmailConfirmation(token.Sub, input.ChannelID, input.EmailAddress, channels.SetEmailEnumDefault)
}

func (r *mutationResolver) UserChannelSetEmail(ctx context.Context, input channels.SetEmailInput) (*channels.Channel, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	_, err := r.userService.SetHasSeenUCEmailPrompt(token.Sub)
	if err != nil {
		return nil, err
	}
	if input.AddToMailing {
		err = r.addToNewsletterList(input.EmailAddress, auth.ApplicationEnumDefault)
		if err != nil {
			return nil, err
		}
	}
	return r.channelService.SendEmailConfirmation(token.Sub, input.ChannelID, input.EmailAddress, channels.SetEmailEnumDefault)

}

func (r *mutationResolver) ChannelsSetEmailConfirm(ctx context.Context, jwt string) (*channels.SetEmailResponse, error) {
	return r.channelService.SetEmailConfirm(jwt)
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

// Listing returns listing associated with this channel
func (r *channelResolver) Listing(ctx context.Context, channel *channels.Channel) (*model.Listing, error) {
	if channel.ChannelType != channels.TypeNewsroom {
		return nil, nil
	}

	return r.listingPersister.ListingByAddress(common.HexToAddress(channel.Reference))
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

func (r *channelResolver) EmailAddressRestricted(ctx context.Context, channel *channels.Channel) (*string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}
	isAdmin, err := r.channelService.IsChannelAdmin(token.Sub, channel.ID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, ErrAccessDenied
	}

	return &channel.EmailAddress, nil
}

func (r *channelResolver) StripeCustomerIDRestricted(ctx context.Context, channel *channels.Channel) (*string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}
	isAdmin, err := r.channelService.IsChannelAdmin(token.Sub, channel.ID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, ErrAccessDenied
	}

	return &channel.StripeCustomerID, nil
}
