package testruntime

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// MockTransactionReader implements eth.TransactionReader interface
type MockTransactionReader struct {
	Transactions map[string]*types.Transaction
	Receipts     map[string]*types.Receipt
}

// AddTransaction adds a transaction
func (r *MockTransactionReader) AddTransaction(txID common.Hash, tx *types.Transaction) {
	r.Transactions[txID.String()] = tx
}

// AddReceipt adds a receipt
func (r *MockTransactionReader) AddReceipt(txID common.Hash, receipt *types.Receipt) {
	r.Receipts[txID.String()] = receipt
}

// TransactionByHash gets the mock tx by hash
func (r *MockTransactionReader) TransactionByHash(ctx context.Context, txHash common.Hash) (tx *types.Transaction, isPending bool, err error) {

	tx, ok := r.Transactions[txHash.String()]
	if !ok {
		return nil, false, ethereum.NotFound
	}

	// if we found a receipt then the tx is not pending
	_, foundReceipt := r.Receipts[txHash.String()]

	return tx, !foundReceipt, nil
}

// TransactionReceipt gets the mock tx receipt
func (r *MockTransactionReader) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	receipt, ok := r.Receipts[txHash.String()]
	if !ok {
		return nil, ethereum.NotFound
	}

	return receipt, nil
}

// NewMockTransactionReader creates a new MockTransactionReader
func NewMockTransactionReader() *MockTransactionReader {
	transactions := make(map[string]*types.Transaction)
	receipts := make(map[string]*types.Receipt)
	return &MockTransactionReader{
		Transactions: transactions,
		Receipts:     receipts,
	}
}
