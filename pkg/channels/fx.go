package channels

import (
	"go.uber.org/fx"
)

// ChannelModule builds channel services
var ChannelModule = fx.Options(
	fx.Provide(
		NewDBPersister,
		NewServiceFromConfig,
	),
)
