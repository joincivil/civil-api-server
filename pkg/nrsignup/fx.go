package nrsignup

import "go.uber.org/fx"

// NrSignupModule returns the services for the nrsignup module
var NrSignupModule = fx.Options(
	fx.Provide(NewService),
)
