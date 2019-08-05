package payments

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// Payment is a transfer of value from one party to the other
type Payment interface {
	Type() string
}

// PaymentModel defines the GORM model for a payment
type PaymentModel struct {
	ID           string    `gorm:"type:uuid;primary_key"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
	DeletedAt    *time.Time
	PaymentType  string  `gorm:"not null;unique_index:payments_idx_type_reference"`
	Reference    string  `gorm:"not_null;unique_index:payments_idx_type_reference"` // user_id, newsroom smart contract address, group DID
	Status       string  `gorm:"not null"`
	CurrencyCode string  `gorm:"not null"`
	Amount       float64 `gorm:"not null"`
	ExchangeRate float64 `gorm:"not null"`
	Comment      string
	Reaction     string
	Data         postgres.Jsonb
	OwnerID      string `gorm:"not null"`
	OwnerType    string `gorm:"not null"`
	EmailAddress string
}

// TableName returns the gorm table name for Base
func (PaymentModel) TableName() string {
	return "payments"
}

// USDEquivalent returns the gorm table name for Base
func (p PaymentModel) USDEquivalent() float64 {
	return p.Amount * p.ExchangeRate
}

// ModelToInterface accepts a payment model struct and returns the payment interface
func ModelToInterface(model *PaymentModel) (Payment, error) {
	var payment Payment
	switch model.PaymentType {
	case "stripe":
		payment = &StripePayment{
			PaymentModel: *model,
		}
	case "ether":
		payment = &EtherPayment{
			PaymentModel: *model,
		}
	case "token":
		payment = &TokenPayment{
			PaymentModel: *model,
		}
	}
	err := json.Unmarshal(model.Data.RawMessage, payment)
	if err != nil {
		return nil, err
	}
	return payment.(Payment), nil
}

// StripePayment is a payment that is created by Stripe
type StripePayment struct {
	PaymentModel `json:"-"`
	PaymentToken string `gorm:"-"`
	EmailAddress string
	UsdAmount    string `gorm:"-"`
}

// Type is the type of payment for StripePayment
func (p StripePayment) Type() string {
	return "stripe"
}

// EtherPayment is a payment in Ether
type EtherPayment struct {
	PaymentModel   `json:"-"`
	TransactionID  string `gorm:"-"`
	PaymentAddress string `gorm:"-"`
	FromAddress    string `gorm:"-"`
	EthAmount      string `gorm:"-"`
	UsdAmount      string `gorm:"-"`
	EmailAddress   string
}

// Type is the type of payment for EtherPayment
func (p EtherPayment) Type() string {
	return "ether"
}

// TokenPayment is a payment using an ERC20 token
type TokenPayment struct {
	PaymentModel  `json:"-"`
	TransactionID string
	TokenAddress  string
	EmailAddress  string
}

// Type is the type of payment for TokenPayment
func (p TokenPayment) Type() string {
	return "token"
}
