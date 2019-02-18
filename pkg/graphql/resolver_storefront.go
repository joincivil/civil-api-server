package graphql

import (
	context "context"

	"github.com/joincivil/civil-api-server/pkg/auth"
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

// StorefrontAirswapTxHash handles the transaction hash from the Airswap onComplete when a sale
// is completed
func (r *mutationResolver) StorefrontAirswapTxHash(ctx context.Context, txHash string) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}

	err := r.storefrontService.AirswapOnComplete(token.Sub, txHash)
	if err != nil {
		return ResponseError, err
	}

	return ResponseOK, nil
}

// StorefrontAirswapCancelled handles the Airswap onCancel when a sale is cancelled
func (r *mutationResolver) StorefrontAirswapCancelled(ctx context.Context) (string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return "", ErrAccessDenied
	}

	err := r.storefrontService.AirswapOnCancel(token.Sub)
	if err != nil {
		return ResponseError, err
	}

	return ResponseOK, nil
}
