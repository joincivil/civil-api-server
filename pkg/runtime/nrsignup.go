package runtime

import (
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"go.uber.org/fx"
)

// NrSignupRuntime returns the runtime module for nrsignup
var NrSignupRuntime = fx.Options(
	fx.Provide(
		func(config *utils.GraphQLConfig) nrsignup.ServiceConfig {
			return nrsignup.ServiceConfig{
				GrantLandingProtoHost: config.ApproveGrantProtoHost,
				ParamAddr:             config.ContractAddresses["civilparameterizer"],
				RegistryListID:        config.RegistryAlertsID,
			}
		},
	),
)
