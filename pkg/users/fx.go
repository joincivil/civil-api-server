package users

import (
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"go.uber.org/fx"
)

// UserModule builds channel services
var UserModule = fx.Options(
	fx.Provide(
		NewUserService,
		NewPersisterFromGorm,
	),
)

// RuntimeModule builds services with concrete implementations
var RuntimeModule = fx.Options(
	fx.Provide(
		func(dbPersister *PostgresPersister) UserPersister {
			return dbPersister
		},
		func(updater *tokencontroller.Service) TokenControllerUpdater {
			return updater
		},
	),
)
