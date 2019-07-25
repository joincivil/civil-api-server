package users

import (
	"go.uber.org/fx"
)

// UserModule builds channel services
var UserModule = fx.Options(
	fx.Provide(
		NewUserService,
		NewPersisterFromGorm,
	),
)
