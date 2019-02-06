package graphql

import (
	context "context"
)

// StorefrontEthPrice returns the current price of ETH in USD
func (r *queryResolver) StorefrontEthPrice(ctx context.Context) (*float64, error) {
	price, err := r.storefrontService.ConvertETHToUSD()
	return &price, err
}

// StorefrontCvlPrice returns the current price of CVL in USD
func (r *queryResolver) StorefrontCvlPrice(ctx context.Context) (*float64, error) {
	price := r.storefrontService.GetQuote(1)
	return &price, nil
}

// User is the resolver for the User type
func (r *queryResolver) StorefrontCvlQuoteUsd(ctx context.Context, usdToSpend float64) (*float64, error) {
	price := r.storefrontService.GetTokensToBuy(usdToSpend)
	return &price, nil
}

// User is the resolver for the User type
func (r *queryResolver) StorefrontCvlQuoteTokens(ctx context.Context, tokensToBuy float64) (*float64, error) {
	price := r.storefrontService.GetQuote(tokensToBuy)
	return &price, nil
}
