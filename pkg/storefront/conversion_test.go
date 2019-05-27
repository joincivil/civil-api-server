package storefront_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joincivil/civil-api-server/pkg/storefront"
)

const krakenTickerString = `{"error":[],"result":{"XETHZUSD":{"a":["107.90000","243","243.000"],"b":["107.89000","44","44.000"],"c":["107.90000","0.00173880"],"v":["127159.90887629","158785.69230958"],"p":["106.27483","105.94236"],"t":[8592,11132],"l":["102.91000","102.91000"],"h":["109.89000","109.89000"],"o":"103.79000"}}}`
const krakenTickerErrorString = `{"error":["EQuery:Unknown asset pair"]}`

func buildTestKraken(t *testing.T, sendError bool) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if sendError {
			w.Write([]byte(krakenTickerErrorString)) // nolint: errcheck
		} else {
			w.Write([]byte(krakenTickerString)) // nolint: errcheck
		}

	}))

	return ts
}

func TestUSDToETH(t *testing.T) {
	conversion := storefront.KrakenCurrencyConversion{
		KrakenURL: "x",
	}

	_, err := conversion.USDToETH()
	if err != storefront.ErrNoPrice {
		t.Fatalf("expecting error to be `ErrNoPrice`")
	}

	conversion.LatestETHUSD = &storefront.KrakenPriceUpdate{
		Price:      100.0,
		LastUpdate: time.Now(),
	}

	price, err := conversion.USDToETH()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 0.01 {
		t.Fatalf("expecting price to be 0.01 but is %v", price)
	}

	conversion.LatestETHUSD = &storefront.KrakenPriceUpdate{
		Price:      100.0,
		LastUpdate: time.Now().Add(-1 * time.Hour),
	}

	_, err = conversion.USDToETH()
	if err != storefront.ErrStalePrice {
		t.Fatalf("expecting error to be `ErrStalePrice`")
	}

}

func TestUpdatePrice(t *testing.T) {
	// set up test server that will get called from `UpdatePrice`
	ts := buildTestKraken(t, false)
	defer ts.Close()

	// construct an instance using the test server
	conversion := storefront.KrakenCurrencyConversion{
		KrakenURL: ts.URL,
	}

	_, err := conversion.USDToETH()
	if err != storefront.ErrNoPrice {
		t.Fatal("expecting error to be ErrNoPrice ")
	}

	conversion.UpdatePrice() // nolint: errcheck

	price, err := conversion.USDToETH()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 0.00927 {
		t.Fatalf("expecting price to be 0.0093 but it is %v", price)
	}

	price = conversion.LatestETHUSD.Price
	if price != 107.9 {
		t.Fatalf("expecting price to be 107.9 but it is %v", price)
	}
}

func TestUpdatePriceError(t *testing.T) {
	// set up test server that will get called from `UpdatePrice`
	ts := buildTestKraken(t, true)
	defer ts.Close()

	// construct an instance using the test server
	conversion := storefront.KrakenCurrencyConversion{
		KrakenURL: ts.URL,
	}

	conversion.UpdatePrice() // nolint: errcheck
	_, err := conversion.USDToETH()
	if err != storefront.ErrNoPrice {
		t.Fatal("expecting error to be ErrNoPrice")
	}

	if conversion.LatestETHUSD != nil {
		t.Fatal("expecting LatestETHUSD to be nil")
	}
}

func TestRoundFloat(t *testing.T) {

	if storefront.RoundFloat(1.18888888, 0) != 1.0 {
		t.Fatal("expecting 1")
	}
	if storefront.RoundFloat(1.18888888, 1) != 1.2 {
		t.Fatal("expecting 1.2")
	}
	if storefront.RoundFloat(1.18888888, 2) != 1.19 {
		t.Fatalf("expecting 1.19")
	}
	if storefront.RoundFloat(1.18888888, 3) != 1.189 {
		t.Fatalf("expecting 1.189")
	}
}
