package graphql

import (
	context "context"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/users"
)

// User is the resolver for the User type
func (r *Resolver) User() graphql.UserResolver {
	return &userResolver{r}
}

// TYPE RESOLVERS

type userResolver struct{ *Resolver }

// NrStep returns the most recent step the user made in Newsroom signup
func (r *userResolver) NrStep(ctx context.Context, obj *users.User) (*int, error) {
	return &obj.NewsroomStep, nil
}

// NrFurthestStep returns the furthest step the user made in Newsroom signup
func (r *userResolver) NrFurthestStep(ctx context.Context, obj *users.User) (*int, error) {
	return &obj.NewsroomFurthestStep, nil
}

// NrLastSeen returns the timestamp secs from epoch since the user was last seen
// in newsroom signup
func (r *userResolver) NrLastSeen(ctx context.Context, obj *users.User) (*int, error) {
	asInt := int(obj.NewsroomLastSeen)
	return &asInt, nil
}

// Channels returns the channels that the user is a member of
func (r *userResolver) Channels(ctx context.Context, obj *users.User) ([]*channels.ChannelMember, error) {

	return r.channelService.GetUserChannels(obj.UID)
}

// QUERIES

func (r *queryResolver) CurrentUser(ctx context.Context) (*users.User, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	return r.userService.GetUser(users.UserCriteria{UID: token.Sub})
}

// MUTATIONS

func (r *mutationResolver) UserSetEthAddress(ctx context.Context, input users.SignatureInput) (*string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	err := auth.VerifyEthChallengeAndSignature(auth.ChallengeRequest{
		ExpectedPrefix: "I control this address",
		GracePeriod:    5 * 60, // 5 minutes
		InputAddress:   input.Signer,
		InputChallenge: input.Message,
		Signature:      input.Signature,
	})
	if err != nil {
		return nil, err
	}

	user, err := r.userService.SetEthAddress(users.UserCriteria{UID: token.Sub}, input.Signer)
	if err != nil {
		return nil, err
	}
	rtn := user.EthAddress
	return &rtn, nil
}

func (r *mutationResolver) UserUpdate(ctx context.Context, uid *string, input *users.UserUpdateInput) (*users.User, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}
	// If UID is not passed, use the UID in the token
	// TODO(PN): Can remove the uid as an input and just use the uid in the token
	if uid == nil {
		uid = &token.Sub
	}

	if *uid != token.Sub {
		return nil, ErrUserNotAuthorized
	}

	user, err := r.userService.UpdateUser(*uid, input)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// func (r *mutationResolver) SetUserHandleAndEmail(ctx context.Context, input *users.UserUpdateInput) (*users.User, error) {
// 	token := auth.ForContext(ctx)
// 	if token == nil {
// 		return nil, ErrAccessDenied
// 	}

// 	uid := &token.Sub

// 	user, err := r.userService.UpdateUser(*uid, input)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return user, nil
// }
// SetEmailConfirm
