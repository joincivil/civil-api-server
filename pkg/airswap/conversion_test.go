package airswap_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joincivil/civil-api-server/pkg/airswap"
)

const krakenTickerString = `{"error":[],"result":{"XETHZUSD":{"a":["107.90000","243","243.000"],"b":["107.89000","44","44.000"],"c":["107.90000","0.00173880"],"v":["127159.90887629","158785.69230958"],"p":["106.27483","105.94236"],"t":[8592,11132],"l":["102.91000","102.91000"],"h":["109.89000","109.89000"],"o":"103.79000"}}}`
const krakenTickerErrorString = `{"error":["EQuery:Unknown asset pair"]}`

func buildTestKraken(t *testing.T, sendError bool) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if sendError {
			w.Write([]byte(krakenTickerErrorString))
		} else {
			w.Write([]byte(krakenTickerString))
		}

	}))

	return ts
}

func TestUSDToETH(t *testing.T) {
	conversion := airswap.KrakenPairPricing{
		KrakenURL: "x",
	}

	_, err := conversion.USDToETH()
	if err != airswap.ErrNoPrice {
		t.Fatalf("expecting error to be `ErrNoPrice`")
	}

	conversion.LastestETHUSD = &airswap.KrakenPriceUpdate{
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

	conversion.LastestETHUSD = &airswap.KrakenPriceUpdate{
		Price:      100.0,
		LastUpdate: time.Now().Add(-1 * time.Hour),
	}

	_, err = conversion.USDToETH()
	if err != airswap.ErrStalePrice {
		t.Fatalf("expecting error to be `ErrStalePrice`")
	}

}

func TestUpdatePrice(t *testing.T) {
	// set up test server that will get called from `UpdatePrice`
	ts := buildTestKraken(t, false)
	defer ts.Close()

	// construct an instance using the test server
	conversion := airswap.KrakenPairPricing{
		KrakenURL: ts.URL,
	}

	_, err := conversion.USDToETH()
	if err != airswap.ErrNoPrice {
		t.Fatal("expecting error to be ErrNoPrice ")
	}

	conversion.UpdatePrice()

	price, err := conversion.USDToETH()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 0.0093 {
		t.Fatalf("expecting price to be 107.9 but it is %v", price)
	}

	price = conversion.LastestETHUSD.Price
	if price != 107.9 {
		t.Fatalf("expecting price to be 107.9 but it is %v", price)
	}
}

func TestUpdatePriceError(t *testing.T) {
	// set up test server that will get called from `UpdatePrice`
	ts := buildTestKraken(t, true)
	defer ts.Close()

	// construct an instance using the test server
	conversion := airswap.KrakenPairPricing{
		KrakenURL: ts.URL,
	}

	conversion.UpdatePrice()
	_, err := conversion.USDToETH()
	if err != airswap.ErrNoPrice {
		t.Fatal("expecting error to be ErrNoPrice")
	}

	if conversion.LastestETHUSD != nil {
		t.Fatal("expecting LastestETHUSD to be nil")
	}
}
