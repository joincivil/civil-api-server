package airswap

// TokenSupply returns the number of tokens remaining to be sold
func TokenSupply() float64 {
	// TODO(dankins): these should draw from CVLToken wallet addresses
	multisigSupply := 33000000.0
	hotWalletSupply := 1000000.0

	return multisigSupply + hotWalletSupply
}
