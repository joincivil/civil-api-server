package payments

import (
	"github.com/ethereum/go-ethereum"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/go-common/pkg/eth"
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

// RuntimeModule builds channel services with concrete implementations
var RuntimeModule = fx.Options(
	fx.Provide(
		func(helper *eth.Helper) ethereum.TransactionReader {
			return helper.Blockchain.(ethereum.TransactionReader)
		},
		func(channel *channels.Service) ChannelHelper {
			return channel
		},
		func(ethpay *EthereumPaymentService) EthereumValidator {
			return ethpay
		},
		func(stripe *StripeService) StripeCharger {
			return stripe
		},
	),
)
