package testruntime

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/glog"
	"github.com/joho/godotenv"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/runtime"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/email"

	"go.uber.org/fx"
	"os"
)

const (
	sendGridKeyEnvVar = "SENDGRID_TEST_KEY"
)

// TestModule constructs test services
var TestModule = fx.Options(
	runtime.Services,
	runtime.EthRuntime,
	runtime.UsersRuntime,
	runtime.ChannelsRuntime,
	runtime.JsonbRuntime,
	fx.Provide(
		testutils.GetTestDBConnection,
		tokencontroller.NewService,
		NewMockIPFS,
		NewMockPaymentHelper,
		NewMockTransactionReader,
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
		// BuildConfig initializes the config
		func() *utils.GraphQLConfig {
			err := godotenv.Load("../../.env")
			if err != nil {
				glog.Errorf("Invalid graphql config: err: %v\n", err)
				os.Exit(2)
			}
			err = godotenv.Load("../../.env.local")
			if err != nil {
				glog.Errorf("Invalid graphql config: err: %v\n", err)
				os.Exit(2)
			}
			config := &utils.GraphQLConfig{}
			flag.Usage = func() {
				config.OutputUsage()
				os.Exit(0)
			}
			flag.Parse()
			err = config.PopulateFromEnv()
			if err != nil {
				config.OutputUsage()
				glog.Errorf("Invalid graphql config: err: %v\n", err)
				os.Exit(2)
			}
			fmt.Printf("EthereumDefaultPrivateKey %v\n", config.EthereumDefaultPrivateKey)

			return config
		},
		func() newsrooms.ToolsConfig {
			return newsrooms.ToolsConfig{
				ApplicationTokens: 100,
				RescueAddress:     common.HexToAddress("0xddB9e9957452d0E39A5E43Fd1AB4aE818aecC6aD"),
			}
		},

		func() nrsignup.ServiceConfig {
			return nrsignup.ServiceConfig{
				GrantLandingProtoHost: "",
				ParamAddr:             "",
				RegistryListID:        "",
			}
		},
		func() *utils.JwtTokenGenerator {
			return &utils.JwtTokenGenerator{Secret: []byte("test")}
		},
		func() *email.Emailer {
			return email.NewEmailerWithSandbox(os.Getenv(sendGridKeyEnvVar), true)
		},
	),
)
