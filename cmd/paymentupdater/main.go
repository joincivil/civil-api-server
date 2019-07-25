package main

import (
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/graphqlmain"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/runtime"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/go-common/pkg/newsroom"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		runtime.Module,
		payments.PaymentModule,
		channels.ChannelModule,
		posts.PostModule,
		users.UserModule,
		storefront.StorefrontModule,
		fx.Provide(
			graphqlmain.NewGorm,
			graphqlmain.BuildConfig,
			graphqlmain.NewETHHelper,
			graphqlmain.NewDeployerContractAddresses,
			newsroom.NewService,
			tokencontroller.NewService,
		),
		fx.Invoke(payments.PaymentUpdaterCron),
	)

	app.Run()
}
