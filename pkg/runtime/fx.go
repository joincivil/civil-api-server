package runtime

import (
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/users"
	"go.uber.org/fx"
)

// Services provide abstract implementations
var Services = fx.Options(
	storefront.RuntimeModule,
	NewsroomRuntime,
	NrSignupRuntime,
	payments.PaymentModule,
	channels.ChannelModule,
	posts.PostModule,
	users.UserModule,
	storefront.StorefrontModule,
	newsrooms.NewsroomModule,
	nrsignup.NrSignupModule,
	jsonstore.JsonbModule,
)

// Module provides concrete implementations
var Module = fx.Options(
	Services,
	EthRuntime,
	ChannelsRuntime,
	UsersRuntime,
	PaymentsRuntime,
	JsonbRuntime,
)
