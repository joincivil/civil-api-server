package model

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TokenTransferParams are the params to initialize a new TokenTransfer
type TokenTransferParams struct {
	ToAddress    common.Address
	FromAddress  common.Address
	Amount       *big.Int
	TransferDate int64
	BlockNumber  uint64
	TxHash       common.Hash
	TxIndex      uint
	BlockHash    common.Hash
	Index        uint
}

// NewTokenTransfer is a convenience method to init a TokenTransfer struct
func NewTokenTransfer(params *TokenTransferParams) *TokenTransfer {
	return &TokenTransfer{
		toAddress:    params.ToAddress,
		fromAddress:  params.FromAddress,
		amount:       params.Amount,
		transferDate: params.TransferDate,
		blockData: BlockData{
			blockNumber: params.BlockNumber,
			txHash:      params.TxHash.Hex(),
			txIndex:     params.TxIndex,
			blockHash:   params.BlockHash.Hex(),
			index:       params.Index,
		},
	}
}

// TokenTransfer represents a single token transfer made by an individual
type TokenTransfer struct {
	// The address of the purchaser (purchaser wallet addr)
	toAddress common.Address

	// wallet from which the tokens were transferred from (civil wallet)
	fromAddress common.Address

	// amount in gwei, not tokens
	amount *big.Int

	transferDate int64

	blockData BlockData
}

// ToAddress is the address of the purchaser (purchaser wallet)
func (t *TokenTransfer) ToAddress() common.Address {
	return t.toAddress
}

// FromAddress is the address of the token source (civil wallet)
func (t *TokenTransfer) FromAddress() common.Address {
	return t.fromAddress
}

// Amount is the amount of token transferred
// Is in number of gwei, not in token
func (t *TokenTransfer) Amount() *big.Int {
	return t.amount
}

// AmountInToken is the amount in tokens
func (t *TokenTransfer) AmountInToken() *big.Int {
	return t.amount.Quo(t.amount, big.NewInt(1e18))
}

// TransferDate is the purchase date
// Should be based on the block timestamp
func (t *TokenTransfer) TransferDate() int64 {
	return t.transferDate
}

// BlockData has all the block data from the block associated with the event
// NOTE: This is not secured by consensus
func (t *TokenTransfer) BlockData() BlockData {
	return t.blockData
}

// Equals compares this token transfer structs with another for equality
func (t *TokenTransfer) Equals(purchase *TokenTransfer) bool {
	if t.toAddress.Hex() == purchase.ToAddress().Hex() {
		return false
	}
	if t.fromAddress.Hex() == purchase.FromAddress().Hex() {
		return false
	}
	if t.amount.Int64() != purchase.Amount().Int64() {
		return false
	}
	if t.transferDate != purchase.TransferDate() {
		return false
	}
	return true
}
