// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"github.com/ethereum/go-ethereum/common"
)

// Metadata represents any metadata associated with a governance event
type Metadata map[string]interface{}

// BlockData is block data from the block. NOTE: filled in by node, not secured by consensus
// TODO(IS): Instead of intializing this in NewGovernanceEvent, create constructor for this.
type BlockData struct {
	blockNumber uint64
	txHash      string
	txIndex     uint
	blockHash   string
	index       uint
}

// BlockNumber is the block number in which the transaction was included
func (bd *BlockData) BlockNumber() uint64 {
	return bd.blockNumber
}

// TxHash is the hash of the transaction
func (bd *BlockData) TxHash() string {
	return bd.txHash
}

// TxIndex is the index of the transaction in the block
func (bd *BlockData) TxIndex() uint {
	return bd.txIndex
}

// BlockHash is the hash of the block in which the transaction was included
func (bd *BlockData) BlockHash() string {
	return bd.blockHash
}

// Index is the index of the log in the receipt
func (bd *BlockData) Index() uint {
	return bd.index
}

// NewGovernanceEvent is a convenience function to init a new GovernanceEvent
// struct
func NewGovernanceEvent(listingAddr common.Address, senderAddr common.Address,
	metadata Metadata, eventType string, creationDateTs int64,
	lastUpdatedDateTs int64, eventHash string, blockNumber uint64,
	txHash common.Hash, txIndex uint, blockHash common.Hash, index uint) *GovernanceEvent {
	ge := &GovernanceEvent{}
	ge.listingAddress = listingAddr
	ge.senderAddress = senderAddr
	ge.metadata = metadata
	ge.governanceEventType = eventType
	ge.creationDateTs = creationDateTs
	ge.lastUpdatedDateTs = lastUpdatedDateTs
	ge.eventHash = eventHash
	ge.blockData = BlockData{
		blockNumber: blockNumber,
		txHash:      txHash.Hex(),
		txIndex:     txIndex,
		blockHash:   blockHash.Hex(),
		index:       index,
	}
	return ge
}

// GovernanceEvent represents a single governance event made to a listing.  Meant
// to be a central log of these events for audit.
type GovernanceEvent struct {
	listingAddress common.Address

	senderAddress common.Address

	metadata Metadata

	governanceEventType string

	creationDateTs int64

	lastUpdatedDateTs int64

	eventHash string

	blockData BlockData
}

// ListingAddress returns the listing address associated with this event
func (g *GovernanceEvent) ListingAddress() common.Address {
	return g.listingAddress
}

// SenderAddress returns the address of the sender of this event. The sender
// is the address that initiated this event
func (g *GovernanceEvent) SenderAddress() common.Address {
	return g.senderAddress
}

// Metadata returns the Metadata associated with the event. It might be anything
// returned in the event payload
func (g *GovernanceEvent) Metadata() Metadata {
	return g.metadata
}

// GovernanceEventType returns the type of this event
func (g *GovernanceEvent) GovernanceEventType() string {
	return g.governanceEventType
}

// CreationDateTs is the timestamp of creation for this event
func (g *GovernanceEvent) CreationDateTs() int64 {
	return g.creationDateTs
}

// LastUpdatedDateTs is the timestamp of the last update of this event
func (g *GovernanceEvent) LastUpdatedDateTs() int64 {
	return g.lastUpdatedDateTs
}

// SetLastUpdatedDateTs sets the value of the last time this governance event was updated
func (g *GovernanceEvent) SetLastUpdatedDateTs(date int64) {
	g.lastUpdatedDateTs = date
}

// EventHash is the hash from the event
func (g *GovernanceEvent) EventHash() string {
	return g.eventHash
}

// BlockData has all the block data from the block associated with this event.
// NOTE: This is not secured by consensus.
func (g *GovernanceEvent) BlockData() BlockData {
	return g.blockData
}
