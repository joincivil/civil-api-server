package runtime

import (
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/newsroom"
	"go.uber.org/fx"
)

// NewsroomRuntime provides concrete implementations for newsroom services
var NewsroomRuntime = fx.Options(
	fx.Provide(
		func(cachingService *newsrooms.CachingService) newsrooms.Service {
			return cachingService
		},
		func(config *utils.GraphQLConfig) newsrooms.ToolsConfig {
			return newsrooms.ToolsConfig{
				ApplicationTokens: config.TcrApplicationTokens,
				RescueAddress:     config.FastPassRescueMultisig,
			}
		},
		func(ipfs *shell.Shell) newsroom.IPFSHelper {
			return ipfs
		},
	),
)
