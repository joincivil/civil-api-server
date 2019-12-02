package graphqlmain

import (
	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/utils"

	cerrors "github.com/joincivil/go-common/pkg/errors"
)

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
