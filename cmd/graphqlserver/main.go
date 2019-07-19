package main

import (
	"github.com/joincivil/civil-api-server/pkg/graphqlmain"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		graphqlmain.MainModule,
	)

	app.Run()
}
