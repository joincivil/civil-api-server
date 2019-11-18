package runtime

import (
	"go.uber.org/fx"

	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

// JsonbRuntime builds services with concrete implementations
// Functions to convert concrete objects to their interfaces
var JsonbRuntime = fx.Options(
	fx.Provide(
		func(dbPersister *jsonstore.PostgresPersister) jsonstore.JsonbPersister {
			return dbPersister
		},
	),
)
