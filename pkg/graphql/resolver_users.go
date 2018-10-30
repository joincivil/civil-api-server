package graphql

import (
	context "context"
	"fmt"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/users"
)

// User is the resolver for the User type
func (r *Resolver) User() graphql.UserResolver {
	return &userResolver{r}
}

// TYPE RESOLVERS

type userResolver struct{ *Resolver }

// Invoices returns a list of `invoicing.PostgresInvoice` created by the user
func (r *userResolver) Invoices(ctx context.Context, obj *users.User) ([]*invoicing.PostgresInvoice, error) {
	return r.invoicePersister.Invoices(obj.UID, obj.Email, "", "")
}

// IsTokenFoundryRegistered determines if the User is registered to buy CVL on TokenFoundry
func (r *userResolver) IsTokenFoundryRegistered(ctx context.Context, obj *users.User) (*bool, error) {
	ok, err := r.tokenFoundry.GetKYCStatus(obj.Email)
	return &ok, err
}

// QUERIES

func (r *queryResolver) CurrentUser(ctx context.Context) (*users.User, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, fmt.Errorf("Access denied")
	}

	return r.userService.GetUser(users.UserCriteria{Email: token.Sub}, true)
}

// MUTATIONS

func (r *mutationResolver) UserSetEthAddress(ctx context.Context, input users.SetEthAddressInput) (*string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, fmt.Errorf("Access denied")
	}

	user, err := r.userService.SetEthAddress(users.UserCriteria{Email: token.Sub}, &input)
	if err != nil {
		return nil, err
	}
	rtn := user.EthAddress
	return &rtn, nil
}

func (r *mutationResolver) UserUpdate(ctx context.Context, uid *string, input *users.UserUpdateInput) (*users.User, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, fmt.Errorf("Access denied")
	}

	user, err := r.userService.UpdateUser(token, *uid, input)
	if err != nil {
		return nil, err
	}
	return user, nil
}
