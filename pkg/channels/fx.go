package channels

import (
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/go-common/pkg/newsroom"
	"go.uber.org/fx"
)

// ChannelModule builds channel services
var ChannelModule = fx.Options(
	fx.Provide(
		NewDBPersister,
		NewService,
	),
)

// RuntimeModule builds channel services with concrete implementations
var RuntimeModule = fx.Options(
	fx.Provide(
		func(dbPersister *DBPersister) Persister {
			return dbPersister
		},
		func(newsroomService *newsroom.Service) NewsroomHelper {
			return newsroomService
		},
		func(userService *users.UserService) UserEthAddressGetter {
			return userService
		},
	),
)
