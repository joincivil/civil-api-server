package posts

import "go.uber.org/fx"

// PostModule builds post services
var PostModule = fx.Options(
	fx.Provide(
		NewDBPostPersister,
		NewService,
	),
)
