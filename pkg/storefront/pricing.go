package storefront

import (
	"math"
	"math/cmplx"
)

// PricingManager provides utilities to price tokens along a linear curve
type PricingManager struct {
	TotalOffering       float64 // the total number of tokens being sold
	TotalRaiseTargetUSD float64 // the amount in USD that will be raised from selling `TotalOffering` tokens
	StartingPrice       float64 // the price of the first token sold
	TokensSold          float64 // TokensSold how far along the price slope we are
}

// NewPricingManager returns an instance of PricingManager
func NewPricingManager(totalOffering float64, totalRaiseUSD float64, startingPrice float64) *PricingManager {
	return &PricingManager{
		TotalOffering:       totalOffering,
		TotalRaiseTargetUSD: totalRaiseUSD,
		StartingPrice:       startingPrice,
	}
}

// GetQuote accepts the number of tokens the user would like to purchase and returns a price in USD
func (m *PricingManager) GetQuote(numTokens float64) float64 {
	// find the base rectangle
	aR := numTokens * m.CalculatePriceAtX(m.TokensSold)

	// find the area of the triangle
	startPrice := m.CalculatePriceAtX(m.TokensSold)
	endPrice := m.CalculatePriceAtX(m.TokensSold + numTokens)
	aT := 0.5 * (endPrice - startPrice) * numTokens

	return math.Round((aR+aT)*10000) / 10000
}

// GetTokensToBuy returns the number of tokens you should buy if you want to spend a certain amount of USD
func (m *PricingManager) GetTokensToBuy(totalSpendUSD float64) float64 {
	slope := m.CalculateSlope()
	// m = slope, s = starting price, a = totalSpendUSD
	// x = number of tokens to buy

	// a = aR + aT
	// aT = 1/2(bh) = 1/2( x * mx )
	// aR = sx
	// a = sx + 1/2(x * mx)
	// 2a = 2sx + mx^2
	// mx^2 + 2sx - 2a = 0
	startingPrice := m.CalculatePriceAtX(m.TokensSold)

	// solve the quadratic
	a := complex(slope, 0)
	b := complex(2*startingPrice, 0)
	c := complex(-2*totalSpendUSD, 0)

	negB := -b
	twoA := 2 * a
	bSquared := b * b
	fourAC := 4 * a * c
	discrim := bSquared - fourAC
	sq := cmplx.Sqrt(discrim)
	xpos := (negB + sq) / twoA
	return math.Round(real(xpos)*100000) / 100000
}

// CalculateSlope determines the slope of the line to get from 0 tokens sold to TotalOffering
func (m *PricingManager) CalculateSlope() float64 {
	// area of rectangle is the total amount raised if we kept the price the same for all tokens
	aR := m.TotalOffering * m.StartingPrice
	// area of triangle is the amount raised from the increase in token price from the start
	aT := m.TotalRaiseTargetUSD - aR

	// totalPriceIncrease is the change in height from the StartingPrice to the LastPrice
	// another way to look at it is the height of aT
	// area of triangle (a) = 1/2(b*h)
	// h = 2a / b
	totalPriceIncrease := (2.0 * aT) / m.TotalOffering

	// slope is the change in price over the number of tokens on offer
	return totalPriceIncrease / m.TotalOffering
}

// CalculatePriceAtX returns the starting price after X tokens sold
func (m *PricingManager) CalculatePriceAtX(tokensSold float64) float64 {

	// y = mx + b
	// m = slope
	// x = number of tokens sold
	// b = starting price

	return m.CalculateSlope()*tokensSold + m.StartingPrice
}

// IncreaseTokensSold increases `TokensSold` by `numTokens`
func (m *PricingManager) IncreaseTokensSold(numTokens float64) {
	m.TokensSold += numTokens
}
