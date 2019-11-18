package jsonstore

import (
	"go.uber.org/fx"
)

// JsonbModule builds jsonb services
var JsonbModule = fx.Options(
	fx.Provide(
		NewJsonbService,
		NewPersisterFromGorm,
	),
)
