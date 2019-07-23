package runtime

import (
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	"go.uber.org/fx"
)

// NewsroomRuntime provides concrete implementations for newsroom services
var NewsroomRuntime = fx.Options(
	fx.Provide(
		func(cachingService *newsrooms.CachingService) newsrooms.Service {
			return cachingService
		},
	),
)
