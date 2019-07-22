package newsrooms

import (
	"go.uber.org/fx"
)

var RuntimeModule = fx.Options(
	fx.Provide(
		func(cachingService *CachingService) Service {
			return cachingService
		},
	),
)
