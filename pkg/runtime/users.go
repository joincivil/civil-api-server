package runtime

import (
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/civil-api-server/pkg/users"
	"go.uber.org/fx"
)

// UsersRuntime builds services with concrete implementations
var UsersRuntime = fx.Options(
	fx.Provide(
		func(dbPersister *users.PostgresPersister) users.UserPersister {
			return dbPersister
		},
		func(updater *tokencontroller.Service) users.TokenControllerUpdater {
			return updater
		},
	),
)
