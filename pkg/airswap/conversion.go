package airswap

// PairPricing provides an interface to return the price of a currency pair
type PairPricing interface {
	USDToETH() float64
}

// StaticPairPricing implements StaticPairPricing that returns a static conversion rate
type StaticPairPricing struct {
	PriceOfETH float64
}

// USDToETH returns the price of 1 USD in ETH
func (s *StaticPairPricing) USDToETH() float64 {
	return 1.0 / s.PriceOfETH
}

// TODO(dankins): write a PairPricing that pulls from a webservice
