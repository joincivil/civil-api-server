package main

import (
	"github.com/joincivil/civil-api-server/pkg/graphqlmain"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/runtime"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/go-common/pkg/newsroom"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		runtime.Module,
		fx.Provide(
			graphqlmain.NewGorm,
			graphqlmain.BuildConfig,
			newsroom.NewService,
			tokencontroller.NewService,
		),
		fx.Invoke(payments.PaymentUpdaterCron),
	)

	app.Run()
}
