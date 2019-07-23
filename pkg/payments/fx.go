package payments

import (
	"go.uber.org/fx"
)

// PaymentModule is an fx Module
var PaymentModule = fx.Options(
	fx.Provide(
		NewEthereumPaymentService,
		NewService,
		NewStripeServiceFromConfig,
	),
)
