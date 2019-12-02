package newsrooms

import (
	"github.com/joincivil/go-common/pkg/newsroom"
	"go.uber.org/fx"
)

// NewsroomModule provides functions to construct newsroom services
var NewsroomModule = fx.Options(
	fx.Provide(
		NewCachingService,
		newsroom.NewService,
		NewTools,
	),
)
