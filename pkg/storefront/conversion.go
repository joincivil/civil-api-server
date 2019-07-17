package storefront

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"

	log "github.com/golang/glog"
)

var (
	// ErrNoPrice is returned when a price is requested but none has been set yet
	ErrNoPrice = errors.New("no price data available")
	// ErrStalePrice is returned when a price is requested but it is too old to use
	ErrStalePrice = errors.New("price data is too old")
)

const (
	// stalePriceTimeSecs is the number of seconds a conversion rate is valid
	stalePriceTimeSecs = 120
)

// CurrencyConversion provides an interface to return the price of a currency pair
type CurrencyConversion interface {
	USDToETH() (float64, error)
	ETHToUSD() (float64, error)
}

// StaticCurrencyConversion implements StaticCurrencyConversion that returns a static conversion rate
type StaticCurrencyConversion struct {
	PriceOfETH float64
}

// USDToETH returns the price of 1 USD in ETH
func (s StaticCurrencyConversion) USDToETH() (float64, error) {
	return 1.0 / s.PriceOfETH, nil
}

// ETHToUSD returns the price of 1 ETH in USD
func (s StaticCurrencyConversion) ETHToUSD() (float64, error) {
	return s.PriceOfETH, nil
}

// KrakenCurrencyConversion is a CurrencyConversion that is periodically updated by the Kraken API
type KrakenCurrencyConversion struct {
	LatestETHUSD *KrakenPriceUpdate
	UpdateTicker *time.Ticker
	KrakenURL    string
}

// NewKrakenCurrencyConversion returns a new KrakenCurrencyConversion that updates the price every `frequencySeconds`
func NewKrakenCurrencyConversion(frequencySeconds uint) *KrakenCurrencyConversion {
	k := &KrakenCurrencyConversion{KrakenURL: "https://api.kraken.com/0/public/Ticker"}
	if frequencySeconds > 0 {
		k.PricePolling(frequencySeconds)
	} else {
		err := k.UpdatePrice()
		if err != nil {
			log.Errorf("Error with Kraken UpdatePrice %v", err)
		}
	}

	return k
}

// NewKrakenCurrencyConversionWithDefault creates a new CurrencyConversion with default config
func NewKrakenCurrencyConversionWithDefault() *KrakenCurrencyConversion {
	return NewKrakenCurrencyConversion(30)
}

// KrakenPriceUpdate contains the latest price
type KrakenPriceUpdate struct {
	Price      float64
	LastUpdate time.Time
}

// KrakenTickerResponse is the response that contains either an error or result object
type KrakenTickerResponse struct {
	Error  []string                      `json:"error"`
	Result map[string]KrakenTickerResult `json:"result"`
}

// KrakenTickerResult is the response from a successful ticker query
// https://www.kraken.com/features/api#public-market-data
type KrakenTickerResult struct {
	Ask                   []string `json:"a"`
	Bid                   []string `json:"b"`
	Last                  []string `json:"c"`
	Volume                []string `json:"v"`
	VolumeWeightedAverage []string `json:"p"`
	NumberOfTrades        []int    `json:"t"`
	Low24H                []string `json:"l"`
	High24H               []string `json:"h"`
	OpenPrice             string   `json:"o"`
}

// PricePolling calls UpdatePrice at the specified interval
func (k *KrakenCurrencyConversion) PricePolling(frequencySeconds uint) {
	err := k.UpdatePrice()
	if err != nil {
		log.Errorf("Error with Kraken UpdatePrice %v", err)
	}
	k.UpdateTicker = time.NewTicker(time.Duration(frequencySeconds) * time.Second)
	go func() {
		for range k.UpdateTicker.C {
			err := k.UpdatePrice()
			if err != nil {
				log.Errorf("Error with Kraken UpdatePrice %v", err)
			}
		}
	}()
}

// UpdatePrice queries the kraken API and updates the price
func (k *KrakenCurrencyConversion) UpdatePrice() error {
	res, err := http.Get(k.KrakenURL + "?pair=XETHZUSD")
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var result KrakenTickerResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	if len(result.Error) > 0 {
		log.Errorf("Error updating ETH price from Kraken %v", result.Error)
		return errors.New(result.Error[0])
	}

	price, err := strconv.ParseFloat(result.Result["XETHZUSD"].Last[0], 32)
	if err != nil {
		return err
	}

	k.LatestETHUSD = &KrakenPriceUpdate{Price: RoundFloat(price, 5), LastUpdate: time.Now()}
	log.V(3).Infof("Updated ETHUSD Price: %v\n", k.LatestETHUSD.Price)

	return nil

}

// USDToETH returns the latest price update
func (k *KrakenCurrencyConversion) USDToETH() (float64, error) {
	if k.LatestETHUSD == nil {
		return 0, ErrNoPrice
	}
	if time.Since(k.LatestETHUSD.LastUpdate) > (stalePriceTimeSecs * time.Second) {
		return 0, ErrStalePrice
	}

	return RoundFloat(1/k.LatestETHUSD.Price, 5), nil
}

// ETHToUSD returns the latest price update
func (k *KrakenCurrencyConversion) ETHToUSD() (float64, error) {
	if k.LatestETHUSD == nil {
		return 0, ErrNoPrice
	}
	if time.Since(k.LatestETHUSD.LastUpdate) > (stalePriceTimeSecs * time.Second) {
		return 0, ErrStalePrice
	}

	return RoundFloat(k.LatestETHUSD.Price, 5), nil
}

// RoundFloat rounds a float to the specified number of decimals
func RoundFloat(num float64, places int) float64 {
	return math.Round(num*math.Pow10(places)) / math.Pow10(places)
}
