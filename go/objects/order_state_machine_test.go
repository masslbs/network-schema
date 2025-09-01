// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func TestOrderStateMachine_ValidateStateTransition(t *testing.T) {
	tests := []struct {
		name         string
		currentState OrderPaymentState
		newState     OrderPaymentState
		expectError  bool
	}{
		// Valid transitions
		{"Open to Locked", OrderPaymentStateOpen, OrderPaymentStateLocked, false},
		{"Open to Canceled", OrderPaymentStateOpen, OrderPaymentStateCanceled, false},
		{"Locked to PaymentChosen", OrderPaymentStateLocked, OrderPaymentStatePaymentChosen, false},
		{"Locked to Open", OrderPaymentStateLocked, OrderPaymentStateOpen, false},
		{"Locked to Canceled", OrderPaymentStateLocked, OrderPaymentStateCanceled, false},
		{"PaymentChosen to Unpaid", OrderPaymentStatePaymentChosen, OrderPaymentStateUnpaid, false},
		{"PaymentChosen to Canceled", OrderPaymentStatePaymentChosen, OrderPaymentStateCanceled, false},
		{"Unpaid to PaymentChosen", OrderPaymentStateUnpaid, OrderPaymentStatePaymentChosen, false},
		{"Unpaid to Paid", OrderPaymentStateUnpaid, OrderPaymentStatePaid, false},
		{"Unpaid to Canceled", OrderPaymentStateUnpaid, OrderPaymentStateCanceled, false},
		{"Unpaid to UnpaidExpired", OrderPaymentStateUnpaid, OrderPaymentStateUnpaidExpired, false},
		{"Unpaid to PaidLate", OrderPaymentStateUnpaid, OrderPaymentStatePaidLate, false},
		{"Canceled to PaidLate", OrderPaymentStateCanceled, OrderPaymentStatePaidLate, false},
		{"UnpaidExpired to PaidLate", OrderPaymentStateUnpaidExpired, OrderPaymentStatePaidLate, false},
		{"Same state", OrderPaymentStateOpen, OrderPaymentStateOpen, false},

		// Invalid transitions
		{"Open to PaymentChosen", OrderPaymentStateOpen, OrderPaymentStatePaymentChosen, true},
		{"Open to Unpaid", OrderPaymentStateOpen, OrderPaymentStateUnpaid, true},
		{"Open to Paid", OrderPaymentStateOpen, OrderPaymentStatePaid, true},
		{"Locked to Unpaid", OrderPaymentStateLocked, OrderPaymentStateUnpaid, true},
		{"PaymentChosen to Paid", OrderPaymentStatePaymentChosen, OrderPaymentStatePaid, true},
		{"Paid to any state", OrderPaymentStatePaid, OrderPaymentStateOpen, true},
		{"PaidLate to any state", OrderPaymentStatePaidLate, OrderPaymentStateOpen, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrderStateTransitions(tt.currentState, tt.newState)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestOrderStateMachine_ValidateStateRequirements(t *testing.T) {

	tests := []struct {
		name        string
		order       *Order
		targetState OrderPaymentState
		expectError bool
	}{
		{
			name: "Open state - valid",
			order: &Order{
				Items: []OrderedItem{{ListingID: 1, Quantity: 1}},
			},
			targetState: OrderPaymentStateOpen,
			expectError: false,
		},
		{
			name: "Locked state - valid with items",
			order: &Order{
				Items: []OrderedItem{{ListingID: 1, Quantity: 1}},
			},
			targetState: OrderPaymentStateLocked,
			expectError: false,
		},
		{
			name: "Locked state - invalid without items",
			order: &Order{
				Items: []OrderedItem{},
			},
			targetState: OrderPaymentStateLocked,
			expectError: true,
		},
		{
			name: "PaymentChosen state - valid with all requirements",
			order: &Order{
				Items: []OrderedItem{{ListingID: 1, Quantity: 1}},
				ChosenPayee: &Payee{
					Address: ChainAddress{ChainID: 1, EthereumAddress: EthereumAddress{common.HexToAddress("0x1111111111111111111111111111111111111111")}},
				},
				ChosenCurrency: &ChainAddress{ChainID: 1, EthereumAddress: EthereumAddress{common.HexToAddress("0x2222222222222222222222222222222222222222")}},
				ShippingAddress: &AddressDetails{
					Name: "Test", Address1: "123 Main St", City: "City",
					PostalCode: "12345", Country: "US", EmailAddress: "test@example.com",
				},
			},
			targetState: OrderPaymentStatePaymentChosen,
			expectError: false,
		},
		{
			name: "PaymentChosen state - missing ChosenPayee",
			order: &Order{
				Items:          []OrderedItem{{ListingID: 1, Quantity: 1}},
				ChosenCurrency: &ChainAddress{ChainID: 1, EthereumAddress: EthereumAddress{common.HexToAddress("0x2222222222222222222222222222222222222222")}},
				ShippingAddress: &AddressDetails{
					Name: "Test", Address1: "123 Main St", City: "City",
					PostalCode: "12345", Country: "US", EmailAddress: "test@example.com",
				},
			},
			targetState: OrderPaymentStatePaymentChosen,
			expectError: true,
		},
		{
			name: "PaymentChosen state - missing ChosenCurrency",
			order: &Order{
				Items: []OrderedItem{{ListingID: 1, Quantity: 1}},
				ChosenPayee: &Payee{
					Address: ChainAddress{ChainID: 1, EthereumAddress: EthereumAddress{common.HexToAddress("0x1111111111111111111111111111111111111111")}},
				},
				ShippingAddress: &AddressDetails{
					Name: "Test", Address1: "123 Main St", City: "City",
					PostalCode: "12345", Country: "US", EmailAddress: "test@example.com",
				},
			},
			targetState: OrderPaymentStatePaymentChosen,
			expectError: true,
		},
		{
			name: "PaymentChosen state - missing addresses",
			order: &Order{
				Items: []OrderedItem{{ListingID: 1, Quantity: 1}},
				ChosenPayee: &Payee{
					Address: ChainAddress{ChainID: 1, EthereumAddress: EthereumAddress{common.HexToAddress("0x1111111111111111111111111111111111111111")}},
				},
				ChosenCurrency: &ChainAddress{ChainID: 1, EthereumAddress: EthereumAddress{common.HexToAddress("0x2222222222222222222222222222222222222222")}},
			},
			targetState: OrderPaymentStatePaymentChosen,
			expectError: true,
		},
		{
			name: "Unpaid state - valid with PaymentDetails",
			order: &Order{
				Items: []OrderedItem{{ListingID: 1, Quantity: 1}},
				ChosenPayee: &Payee{
					Address: ChainAddress{ChainID: 1, EthereumAddress: EthereumAddress{common.HexToAddress("0x1111111111111111111111111111111111111111")}},
				},
				ChosenCurrency: &ChainAddress{ChainID: 1, EthereumAddress: EthereumAddress{common.HexToAddress("0x2222222222222222222222222222222222222222")}},
				ShippingAddress: &AddressDetails{
					Name: "Test", Address1: "123 Main St", City: "City",
					PostalCode: "12345", Country: "US", EmailAddress: "test@example.com",
				},
				PaymentDetails: &PaymentDetails{
					PaymentID:     Hash{1, 2, 3},
					TTL:           3600,
					ListingHashes: [][]byte{{1, 2, 3}},
				},
			},
			targetState: OrderPaymentStateUnpaid,
			expectError: false,
		},
		{
			name: "Paid state - valid with TxDetails",
			order: &Order{
				Items: []OrderedItem{{ListingID: 1, Quantity: 1}},
				ChosenPayee: &Payee{
					Address: ChainAddress{ChainID: 1, EthereumAddress: EthereumAddress{common.HexToAddress("0x1111111111111111111111111111111111111111")}},
				},
				ChosenCurrency: &ChainAddress{ChainID: 1, EthereumAddress: EthereumAddress{common.HexToAddress("0x2222222222222222222222222222222222222222")}},
				ShippingAddress: &AddressDetails{
					Name: "Test", Address1: "123 Main St", City: "City",
					PostalCode: "12345", Country: "US", EmailAddress: "test@example.com",
				},
				PaymentDetails: &PaymentDetails{
					PaymentID:     Hash{1, 2, 3},
					TTL:           3600,
					ListingHashes: [][]byte{{1, 2, 3}},
				},
				TxDetails: &OrderPaid{
					BlockHash: Hash{4, 5, 6},
				},
			},
			targetState: OrderPaymentStatePaid,
			expectError: false,
		},
		{
			name: "Canceled state - valid with CanceledAt",
			order: &Order{
				Items:      []OrderedItem{{ListingID: 1, Quantity: 1}},
				CanceledAt: &time.Time{},
			},
			targetState: OrderPaymentStateCanceled,
			expectError: false,
		},
		{
			name: "Canceled state - missing CanceledAt",
			order: &Order{
				Items: []OrderedItem{{ListingID: 1, Quantity: 1}},
			},
			targetState: OrderPaymentStateCanceled,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrderStateRequirements(tt.order, tt.targetState)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestOrderStateMachine_CanModifyItems(t *testing.T) {
	tests := []struct {
		name     string
		state    OrderPaymentState
		expected bool
	}{
		{"Open - can modify", OrderPaymentStateOpen, true},
		{"Locked - cannot modify", OrderPaymentStateLocked, false},
		{"PaymentChosen - cannot modify", OrderPaymentStatePaymentChosen, false},
		{"Unpaid - cannot modify", OrderPaymentStateUnpaid, false},
		{"Paid - cannot modify", OrderPaymentStatePaid, false},
		{"Canceled - cannot modify", OrderPaymentStateCanceled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OrderCanModifyItems(tt.state)
			if result != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, result)
			}
		})
	}
}
