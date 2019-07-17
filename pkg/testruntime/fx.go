package testruntime

import (
	"github.com/ethereum/go-ethereum"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/go-common/pkg/eth"
	"go.uber.org/fx"
)

// TestModule constructs test services
var TestModule = fx.Options(
	payments.PaymentModule,
	fx.Provide(
		testutils.GetTestDBConnection,
		eth.NewSimulatedBackendHelper,
		func(mockTxReader *MockTransactionReader) ethereum.TransactionReader {
			return mockTxReader
		},
		func() storefront.CurrencyConversion {
			return storefront.StaticCurrencyConversion{PriceOfETH: 100}
		},
		func(ethPay *payments.EthereumPaymentService) payments.EthereumValidator {
			return ethPay
		},
		func(paymentHelper *MockPaymentHelper) payments.StripeCharger {
			return paymentHelper
		},
		func(paymentHelper *MockPaymentHelper) payments.ChannelHelper {
			return paymentHelper
		},
		NewMockPaymentHelper,
		NewMockTransactionReader,
	),
	fx.Invoke(CleanDatabase),
)
