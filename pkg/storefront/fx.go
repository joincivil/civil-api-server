package storefront

import (
	"go.uber.org/fx"
)

// StorefrontModule is an fx Module
var StorefrontModule = fx.Options(
	fx.Provide(
		// NewService,
		NewKrakenCurrencyConversionWithDefault,
	),
)

// RuntimeModule builds channel services with concrete implementations
var RuntimeModule = fx.Options(
	fx.Provide(
		func(kraken *KrakenCurrencyConversion) CurrencyConversion {
			return kraken
		},
	),
)
