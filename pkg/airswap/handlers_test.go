package airswap_test

import (
	"math/big"
	"testing"

	"github.com/joincivil/civil-api-server/pkg/airswap"
	"github.com/joincivil/civil-api-server/pkg/storefront"
)

func TestSetConvertMakerAmtToTakerAmt(t *testing.T) {
	totalOffering := 34000000.0
	totalRaiseUSD := 20000000.0
	startingPrice := 0.2
	pricingManager := storefront.NewPricingManager(totalOffering, totalRaiseUSD, startingPrice)
	// set 1 ETH = 1 USD to make it easier to think about
	currencyConversion := &storefront.StaticCurrencyConversion{PriceOfETH: 1.0}
	storefrontService := &storefront.Service{Pricing: pricingManager, CurrencyConversion: currencyConversion}

	handlers := &airswap.Handlers{StorefrontService: storefrontService}

	// looking to buy 1 CVL
	result, err := handlers.ConvertMakerAmtToTakerAmt(big.NewInt(1e18 * 1).String())

	if err != nil {
		t.Fatalf("not expecting error: %v", err)
	}

	// cost for 1 CVL should be 0.2 ETH
	expected := big.NewInt(1e18 * 0.2).String()
	if result != expected {
		t.Fatalf("expected result to be %v but was %v", result, expected)
	}

}
