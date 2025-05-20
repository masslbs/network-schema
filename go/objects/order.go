// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	masscbor "github.com/masslbs/network-schema/go/cbor"
)

// Order represents an order placed by a user
type Order struct {
	ID              ObjectID        `validate:"required,gt=0"`
	Items           OrderedItems    `validate:"required"`
	State           OrderState      `validate:"required"`
	InvoiceAddress  *AddressDetails `cbor:",omitempty"`
	ShippingAddress *AddressDetails `cbor:",omitempty"`
	CanceledAt      *time.Time      `cbor:",omitempty"`
	ChosenPayee     *Payee          `cbor:",omitempty"`
	ChosenCurrency  *ChainAddress   `cbor:",omitempty"`
	PaymentDetails  *PaymentDetails `cbor:",omitempty"`
	TxDetails       *OrderPaid      `cbor:",omitempty"`
}

// OrderedItems is a list of items in an order
type OrderedItems []OrderedItem

// OrderedItem represents an item in an order
type OrderedItem struct {
	ListingID    ObjectID `validate:"required,gt=0"`
	VariationIDs []string `cbor:",omitempty"`
	Quantity     uint32   `validate:"required,gt=0"`
}

// OrderValidation validates the state-dependent fields of an order
func OrderValidation(sl validator.StructLevel) {
	order := sl.Current().Interface().(Order)
	switch order.State {
	case OrderStatePaid:
		if order.TxDetails == nil {
			sl.ReportError(order.TxDetails, "TxDetails", "TxDetails", "required", "")
		}
		fallthrough
	case OrderStateUnpaid:
		if order.PaymentDetails == nil {
			sl.ReportError(order.PaymentDetails, "PaymentDetails", "PaymentDetails", "required", "")
		}
		fallthrough
	case OrderStatePaymentChosen:
		if order.ChosenPayee == nil {
			sl.ReportError(order.ChosenPayee, "ChosenPayee", "ChosenPayee", "required", "")
		}
		if order.ChosenCurrency == nil {
			sl.ReportError(order.ChosenCurrency, "ChosenCurrency", "ChosenCurrency", "required", "")
		}
		if order.InvoiceAddress == nil && order.ShippingAddress == nil {
			sl.ReportError(order.InvoiceAddress, "InvoiceAddress", "InvoiceAddress", "either_or", "")
			sl.ReportError(order.ShippingAddress, "ShippingAddress", "ShippingAddress", "either_or", "")
		}
		fallthrough
	case OrderStateCommitted:
		if len(order.Items) == 0 {
			sl.ReportError(order.Items, "Items", "Items", "required", "")
		}
	case OrderStateCanceled:
		if order.CanceledAt == nil {
			sl.ReportError(order.CanceledAt, "CanceledAt", "CanceledAt", "required", "")
		}
	case OrderStateOpen:
		// noop
	default:
		sl.ReportError(order.State, "State", "State", fmt.Sprintf("invalid order state: %d", order.State), "")
	}
}

// OrderState represents the possible states an order can be in
type OrderState uint

const (
	// OrderStateUnspecified is the default and invalid state of an order
	OrderStateUnspecified OrderState = iota
	// OrderStateOpen is the state of an order which is open to being changed
	OrderStateOpen
	// OrderStateCanceled is the state of an order which has been canceled
	OrderStateCanceled
	// OrderStateCommitted is the state of an order which items have been frozen
	OrderStateCommitted
	// OrderStatePaymentChosen is the state of an order which has chosen a payment channel
	OrderStatePaymentChosen
	// OrderStateUnpaid is the state of an order which has not been paid for
	OrderStateUnpaid
	// OrderStatePaid is the state of an order which has been paid for
	OrderStatePaid

	maxOrderState
)

// UnmarshalCBOR implements the cbor.Unmarshaler interface
func (s *OrderState) UnmarshalCBOR(data []byte) error {
	dec := masscbor.DefaultDecoder(bytes.NewReader(data))
	var i uint
	err := dec.Decode(&i)
	if err != nil {
		return err
	}
	if i == uint(OrderStateUnspecified) || i >= uint(maxOrderState) {
		return fmt.Errorf("invalid order state: %d", i)
	}
	*s = OrderState(i)
	return nil
}

// AddressDetails represents shipping or billing address information
type AddressDetails struct {
	Name         string  `validate:"required,notblank"`
	Address1     string  `validate:"required,notblank"`
	Address2     string  `cbor:",omitempty"`
	City         string  `validate:"required,notblank"`
	PostalCode   string  `cbor:",omitempty" validate:"required,notblank"`
	Country      string  `validate:"required,notblank"`
	EmailAddress string  `validate:"required,email"`
	PhoneNumber  *string `cbor:",omitempty" validate:"omitempty,e164"`
}

// PaymentDetails represents the details needed to pay for an order
type PaymentDetails struct {
	PaymentID     Hash
	Total         Uint256
	ListingHashes [][]byte `validate:"required,gt=0"`
	TTL           uint64   `validate:"required,gt=0"` // The time to live in block
	ShopSignature Signature
}

// OrderPaid represents the details of a payment for an order
type OrderPaid struct {
	TxHash    *Hash `cbor:",omitempty"`
	BlockHash Hash  `cbor:",omitempty"`
}

// UnmarshalCBOR implements the cbor.Unmarshaler interface
func (op *OrderPaid) UnmarshalCBOR(data []byte) error {
	var tmp struct {
		TxHash    *Hash `cbor:",omitempty"`
		BlockHash *Hash `cbor:",omitempty"`
	}
	dec := masscbor.DefaultDecoder(bytes.NewReader(data))
	err := dec.Decode(&tmp)
	if err != nil {
		return err
	}
	if tmp.BlockHash == nil {
		return fmt.Errorf("BlockHash must be set")
	}
	op.BlockHash = *tmp.BlockHash
	op.TxHash = tmp.TxHash
	return nil
}
