package graphqlmain

import (
	log "github.com/golang/glog"

	"github.com/joincivil/civil-api-server/pkg/utils"

	cerrors "github.com/joincivil/go-common/pkg/errors"
	"github.com/joincivil/go-common/pkg/eth"
)

// NewETHHelper builds a new eth.Helper instance
func NewETHHelper(config *utils.GraphQLConfig) (*eth.Helper, error) {
	if config.EthAPIURL != "" {
		accounts := map[string]string{}
		if config.EthereumDefaultPrivateKey != "" {
			log.Infof("Initialized default Ethereum account\n")
			accounts["default"] = config.EthereumDefaultPrivateKey
		}
		ethHelper, err := eth.NewETHClientHelper(config.EthAPIURL, accounts)
		if err != nil {
			return nil, err
		}
		log.Infof("Connected to Ethereum using %v\n", config.EthAPIURL)
		return ethHelper, nil
	}

	ethHelper, err := eth.NewSimulatedBackendHelper()
	if err != nil {
		return nil, err
	}
	log.Infof("Connected to Ethereum using Simulated Backend\n")
	return ethHelper, nil
}

func initErrorReporter(config *utils.GraphQLConfig) (cerrors.ErrorReporter, error) {
	if config.StackDriverProjectID == "" && config.SentryDsn == "" {
		log.Infof("Enabling null error reporter")
		return &cerrors.NullErrorReporter{}, nil
	}

	errRepConfig := &cerrors.MetaErrorReporterConfig{
		StackDriverProjectID:      config.StackDriverProjectID,
		StackDriverServiceName:    "api-server",
		StackDriverServiceVersion: "1.0",
		SentryDSN:                 config.SentryDsn,
		SentryDebug:               false,
		SentryEnv:                 config.SentryEnv,
		SentryLoggerName:          "api_logger",
		SentryRelease:             "1.0",
		SentrySampleRate:          1.0,
	}
	reporter, err := cerrors.NewMetaErrorReporter(errRepConfig)
	if err != nil {
		log.Errorf("Error creating meta reporter: %v", err)
		return nil, err
	}
	if reporter == nil {
		log.Infof("Enabling null error reporter")
		return &cerrors.NullErrorReporter{}, nil
	}

	return reporter, nil
}
