package runtime

import (
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/go-common/pkg/newsroom"
	"go.uber.org/fx"
)

// ChannelsRuntime builds channel services with concrete implementations
var ChannelsRuntime = fx.Options(
	fx.Provide(
		func(dbPersister *channels.DBPersister) channels.Persister {
			return dbPersister
		},
		func(newsroomService *newsroom.Service) channels.NewsroomHelper {
			return newsroomService
		},
		func(userService *users.UserService) channels.UserEthAddressGetter {
			return userService
		},
		func(stripe *payments.StripeService) channels.StripeConnector {
			return stripe
		},
	),
)
