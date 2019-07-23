package runtime

import (
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/go-common/pkg/newsroom"
	"go.uber.org/fx"
)

// Module provides concrete implementations
var Module = fx.Options(
	ChannelsRuntime,
	UsersRuntime,
	PaymentsRuntime,
	storefront.RuntimeModule,
	NewsroomRuntime,
	fx.Provide(
		func(ipfs *shell.Shell) newsroom.IPFSHelper {
			return ipfs
		},
	),
)
