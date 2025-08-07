// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import (
	"fmt"
	"slices"
)

// validOrderStateTransitions defines the allowed state transitions based on the state machine diagram
var validOrderStateTransitions = map[OrderPaymentState][]OrderPaymentState{
	OrderPaymentStateOpen: {
		OrderPaymentStateLocked,   // freezing items
		OrderPaymentStateCanceled, // timer expired from LOCKED state
	},
	OrderPaymentStateLocked: {
		OrderPaymentStatePaymentChosen, // all required fields set
		OrderPaymentStateOpen,          // unlock (navigate away)
		OrderPaymentStateCanceled,      // timer expired from LOCKED state
	},
	OrderPaymentStatePaymentChosen: {
		OrderPaymentStateCanceled, // navigate away
		OrderPaymentStateUnpaid,   // payment details generated
	},
	OrderPaymentStateUnpaid: {
		OrderPaymentStatePaymentChosen, // changed payment method
		OrderPaymentStatePaid,          // payment detected and confirmed
		OrderPaymentStateCanceled,      // navigate away
		OrderPaymentStateUnpaidExpired, // timeout expired
		OrderPaymentStatePaidLate,      // late payment after timeout
	},
	OrderPaymentStatePaid: {
		// Terminal state - no transitions out
	},
	OrderPaymentStateCanceled: {
		OrderPaymentStatePaidLate, // late payment received on cancelled order
	},
	OrderPaymentStateUnpaidExpired: {
		OrderPaymentStatePaidLate, // late payment after expiry
	},
	OrderPaymentStatePaidLate: {
		// Terminal state requiring manual resolution
	},
}

// ValidateOrderStateTransitions checks if a transition from currentState to newState is allowed
func ValidateOrderStateTransitions(currentState, newState OrderPaymentState) error {
	if currentState == newState {
		return nil // Same state is always valid
	}

	allowedStates, exists := validOrderStateTransitions[currentState]
	if !exists {
		return fmt.Errorf("invalid current state: %s", currentState.String())
	}

	if slices.Contains(allowedStates, newState) {
		return nil
	}

	return fmt.Errorf("invalid state transition from %s to %s", currentState.String(), newState.String())
}

// ValidateOrderStateRequirements checks if the order meets the requirements for the given state
func ValidateOrderStateRequirements(order *Order, targetState OrderPaymentState) error {
	switch targetState {
	case OrderPaymentStateOpen:
		// Open state has no special requirements beyond basic validation
		return nil

	case OrderPaymentStateLocked:
		if len(order.Items) == 0 {
			return fmt.Errorf("locked orders must have items")
		}
		return nil

	case OrderPaymentStatePaymentChosen:
		// Required fields for payment chosen state
		if order.ChosenPayee == nil {
			return fmt.Errorf("ChosenPayee is required for PaymentChosen state")
		}
		if order.ChosenCurrency == nil {
			return fmt.Errorf("ChosenCurrency is required for PaymentChosen state")
		}
		if order.InvoiceAddress == nil && order.ShippingAddress == nil {
			return fmt.Errorf("either InvoiceAddress or ShippingAddress is required for PaymentChosen state")
		}
		return nil

	case OrderPaymentStateUnpaid:
		// Must have payment details
		if order.PaymentDetails == nil {
			return fmt.Errorf("PaymentDetails is required for Unpaid state")
		}
		return ValidateOrderStateRequirements(order, OrderPaymentStatePaymentChosen)

	case OrderPaymentStatePaid:
		// Must have transaction details
		if order.TxDetails == nil {
			return fmt.Errorf("TxDetails is required for Paid state")
		}
		return ValidateOrderStateRequirements(order, OrderPaymentStateUnpaid)

	case OrderPaymentStateCanceled:
		if order.CanceledAt == nil {
			return fmt.Errorf("CanceledAt is required for Canceled state")
		}
		return nil

	case OrderPaymentStateUnpaidExpired:
		// Should have payment details but no transaction
		if order.PaymentDetails == nil {
			return fmt.Errorf("PaymentDetails is required for UnpaidExpired state")
		}
		return nil

	case OrderPaymentStatePaidLate:
		// Must have transaction details
		if order.TxDetails == nil {
			return fmt.Errorf("TxDetails is required for PaidLate state")
		}
		return nil

	default:
		return fmt.Errorf("invalid target state: %s", targetState.String())
	}
}

// OrderCanModifyItems checks if items can be modified in the current state
func OrderCanModifyItems(currentState OrderPaymentState) bool {
	return currentState == OrderPaymentStateOpen
}
