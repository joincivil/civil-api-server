package runtime

import (
	"github.com/ethereum/go-ethereum"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/go-common/pkg/eth"
	"go.uber.org/fx"
)

// PaymentsRuntime builds payment services with concrete implementations
var PaymentsRuntime = fx.Options(
	fx.Provide(
		func(helper *eth.Helper) ethereum.TransactionReader {
			return helper.Blockchain.(ethereum.TransactionReader)
		},
		func(channel *channels.Service) payments.ChannelHelper {
			return channel
		},
		func(ethpay *payments.EthereumPaymentService) payments.EthereumValidator {
			return ethpay
		},
		func(stripe *payments.StripeService) payments.StripeCharger {
			return stripe
		},
	),
)
