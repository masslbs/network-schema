// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/objects"
)

func (p *Patcher) patchOrder(patch Patch) error {
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("order patch needs an ID")
	}
	objID := *patch.Path.ObjectID
	order, exists := p.shop.Orders.Get(objID)

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if exists {
				return fmt.Errorf("order %d already exists", objID)
			}
			var newOrder objects.Order
			if err := masscbor.Unmarshal(patch.Value, &newOrder); err != nil {
				return fmt.Errorf("failed to unmarshal order: %w", err)
			}
			if err := p.validateOrderReferences(&newOrder); err != nil {
				return err
			}
			if err := p.validator.Struct(newOrder); err != nil {
				return err
			}
			return p.shop.Orders.Insert(objID, newOrder)
		}

		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		return p.addOrderField(&order, patch)

	case AppendOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		return p.appendOrderField(&order, patch)

	case ReplaceOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			var newOrder objects.Order
			if err := masscbor.Unmarshal(patch.Value, &newOrder); err != nil {
				return fmt.Errorf("failed to unmarshal order: %w", err)
			}
			if err := p.validateOrderReferences(&newOrder); err != nil {
				return err
			}
			if err := p.validator.Struct(newOrder); err != nil {
				return err
			}
			return p.shop.Orders.Insert(objID, newOrder)
		}
		return p.replaceOrderField(&order, patch)

	case RemoveOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			return p.shop.Orders.Delete(objID)
		}
		return p.removeOrderField(&order, patch)

	case IncrementOp, DecrementOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		return p.modifyOrderQuantity(&order, patch)
	}

	return fmt.Errorf("unsupported operation: %s", patch.Op)
}

func (p *Patcher) validateOrderReferences(order *objects.Order) error {
	for _, item := range order.Items {
		listing, exists := p.shop.Listings.Get(item.ListingID)
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: Path{ObjectID: &item.ListingID}}
		}

		for _, varID := range item.VariationIDs {
			found := false
			for _, opt := range listing.Options {
				if _, exists := opt.Variations[varID]; exists {
					found = true
					break
				}
			}
			if !found {
				return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: Path{ObjectID: &item.ListingID, Fields: []any{"options", varID}}}
			}
		}
	}

	if order.ChosenPayee != nil {
		found := false
		chainID := order.ChosenPayee.Address.ChainID
		ethAddr := order.ChosenPayee.Address.Address

		// Check if the payee exists in the manifest
		if payeesForChain, chainExists := p.shop.Manifest.Payees[chainID]; chainExists {
			if _, addrExists := payeesForChain[ethAddr]; addrExists {
				found = true
			}
		}

		if !found {
			payeeAddr := common.Address(order.ChosenPayee.Address.Address)
			return ObjectNotFoundError{
				ObjectType: ObjectTypeOrder,
				Path:       Path{Fields: []any{"ChosenPayee", payeeAddr.Bytes()}}}
		}
	}

	if order.ChosenCurrency != nil {
		found := false
		chainID := order.ChosenCurrency.ChainID
		ethAddr := order.ChosenCurrency.Address

		// Check if the currency exists in accepted currencies
		if currenciesForChain, chainExists := p.shop.Manifest.AcceptedCurrencies[chainID]; chainExists {
			if _, addrExists := currenciesForChain[ethAddr]; addrExists {
				found = true
			}
		}

		if !found {
			return ObjectNotFoundError{
				ObjectType: ObjectTypeManifest,
				Path:       Path{Fields: []any{"AcceptedCurrencies", chainID, ethAddr.Bytes()}}}
		}
	}

	return nil
}

var errCannotModdifyCommittedOrder = fmt.Errorf("cannot modify committed order")

func (p *Patcher) addOrderField(order *objects.Order, patch Patch) error {
	if len(patch.Path.Fields) == 0 {
		return fmt.Errorf("field path required for add operation")
	}

	switch patch.Path.Fields[0] {
	case "CanceledAt":
		if order.CanceledAt != nil {
			return fmt.Errorf("canceledAt already set")
		}
		var canceledAt time.Time
		if err := masscbor.Unmarshal(patch.Value, &canceledAt); err != nil {
			return fmt.Errorf("failed to unmarshal canceledAt: %w", err)
		}
		order.CanceledAt = &canceledAt
	case "Items":
		if order.State >= objects.OrderStateCommitted {
			return errCannotModdifyCommittedOrder
		}

		if len(patch.Path.Fields) < 2 {
			var item objects.OrderedItem
			if err := masscbor.Unmarshal(patch.Value, &item); err != nil {
				return fmt.Errorf("failed to unmarshal order item: %w", err)
			}
			listing, exists := p.shop.Listings.Get(item.ListingID)
			if !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: Path{ObjectID: &item.ListingID}}
			}
			for _, varID := range item.VariationIDs {
				found := false
				for _, opt := range listing.Options {
					if _, exists := opt.Variations[varID]; exists {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("variation %s not found in listing %d", varID, item.ListingID)
				}
			}
			order.Items = append(order.Items, item)
			return nil
		}

		var item objects.OrderedItem
		if err := masscbor.Unmarshal(patch.Value, &item); err != nil {
			return fmt.Errorf("failed to unmarshal order item: %w", err)
		}

		listing, exists := p.shop.Listings.Get(item.ListingID)
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: Path{ObjectID: &item.ListingID}}
		}

		for _, varID := range item.VariationIDs {
			found := false
			for _, opt := range listing.Options {
				if _, exists := opt.Variations[varID]; exists {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("variation %s not found in listing %d", varID, item.ListingID)
			}
		}

		idx, err := indexFromAny(patch.Path.Fields[1], len(order.Items))
		if err != nil {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		order.Items = slices.Insert(order.Items, idx, item)

	case "ShippingAddress":
		if order.ShippingAddress != nil {
			return fmt.Errorf("shipping address already set")
		}
		var shippingAddress objects.AddressDetails
		if err := masscbor.Unmarshal(patch.Value, &shippingAddress); err != nil {
			return fmt.Errorf("failed to unmarshal shipping address: %w", err)
		}
		order.ShippingAddress = &shippingAddress
	case "InvoiceAddress":
		if order.InvoiceAddress != nil {
			return fmt.Errorf("invoice address already set")
		}
		var invoiceAddress objects.AddressDetails
		if err := masscbor.Unmarshal(patch.Value, &invoiceAddress); err != nil {
			return fmt.Errorf("failed to unmarshal invoice address: %w", err)
		}
		order.InvoiceAddress = &invoiceAddress
	case "PaymentDetails":
		if order.PaymentDetails != nil {
			return fmt.Errorf("payment details already set")
		}
		var paymentDetails objects.PaymentDetails
		if err := masscbor.Unmarshal(patch.Value, &paymentDetails); err != nil {
			return fmt.Errorf("failed to unmarshal payment details: %w", err)
		}
		order.PaymentDetails = &paymentDetails
	case "ChosenPayee":
		if order.ChosenPayee != nil {
			return fmt.Errorf("chosen payee already set")
		}
		var payee objects.Payee
		if err := masscbor.Unmarshal(patch.Value, &payee); err != nil {
			return fmt.Errorf("failed to unmarshal payee: %w", err)
		}
		order.ChosenPayee = &payee
	case "ChosenCurrency":
		if order.ChosenCurrency != nil {
			return fmt.Errorf("chosen currency already set")
		}
		var currency objects.ChainAddress
		if err := masscbor.Unmarshal(patch.Value, &currency); err != nil {
			return fmt.Errorf("failed to unmarshal currency: %w", err)
		}
		order.ChosenCurrency = &currency
	case "TxDetails":
		if order.TxDetails != nil {
			return fmt.Errorf("tx details already set")
		}
		var txDetails objects.OrderPaid
		if err := masscbor.Unmarshal(patch.Value, &txDetails); err != nil {
			return fmt.Errorf("failed to unmarshal tx details: %w", err)
		}
		order.TxDetails = &txDetails
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
	}
	if err := p.validateOrderReferences(order); err != nil {
		return err
	}
	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func (p *Patcher) appendOrderField(order *objects.Order, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "Items":
		if order.State >= objects.OrderStateCommitted {
			return errCannotModdifyCommittedOrder
		}
		var item objects.OrderedItem
		if err := masscbor.Unmarshal(patch.Value, &item); err != nil {
			return fmt.Errorf("failed to unmarshal order item: %w", err)
		}
		order.Items = append(order.Items, item)
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
	}
	if err := p.validateOrderReferences(order); err != nil {
		return err
	}
	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func (p *Patcher) replaceOrderField(order *objects.Order, patch Patch) error {
	nFields := len(patch.Path.Fields)

	switch patch.Path.Fields[0] {
	case "CanceledAt":
		if order.CanceledAt == nil {
			return fmt.Errorf("CanceledAt not set")
		}
		var canceledAt time.Time
		if err := masscbor.Unmarshal(patch.Value, &canceledAt); err != nil {
			return fmt.Errorf("failed to unmarshal canceledAt: %w", err)
		}
		order.CanceledAt = &canceledAt

	case "State":
		// TODO: check newState transitions
		var newState objects.OrderState
		if err := masscbor.Unmarshal(patch.Value, &newState); err != nil {
			return fmt.Errorf("failed to unmarshal order state: %w", err)
		}

		// TODO: check state transitions
		order.State = newState

	case "ChosenPayee":
		var payee objects.Payee
		if err := masscbor.Unmarshal(patch.Value, &payee); err != nil {
			return fmt.Errorf("failed to unmarshal payee: %w", err)
		}

		// Check if the payee exists in the manifest
		found := false
		if chainPayees, exists := p.shop.Manifest.Payees[payee.Address.ChainID]; exists {
			_, found = chainPayees[payee.Address.Address]
		}

		if !found {
			return fmt.Errorf("payee not found in manifest payees")
		}
		order.ChosenPayee = &payee

	case "ChosenCurrency":
		var currency objects.ChainAddress
		if err := masscbor.Unmarshal(patch.Value, &currency); err != nil {
			return fmt.Errorf("failed to unmarshal currency: %w", err)
		}

		// Check if the currency exists in accepted currencies
		found := false
		if chainCurrencies, exists := p.shop.Manifest.AcceptedCurrencies[currency.ChainID]; exists {
			_, found = chainCurrencies[currency.Address]
		}

		if !found {
			return fmt.Errorf("currency not found in accepted currencies")
		}
		order.ChosenCurrency = &currency

	case "PaymentDetails":
		var details objects.PaymentDetails
		if err := masscbor.Unmarshal(patch.Value, &details); err != nil {
			return fmt.Errorf("failed to unmarshal PaymentDetails: %w", err)
		}
		order.PaymentDetails = &details

	case "TxDetails":
		var details objects.OrderPaid
		if err := masscbor.Unmarshal(patch.Value, &details); err != nil {
			return fmt.Errorf("failed to unmarshal TxDetails: %w", err)
		}
		order.TxDetails = &details

	case "Items":
		if order.State >= objects.OrderStateCommitted {
			return errCannotModdifyCommittedOrder
		}
		switch {
		case nFields == 1:
			// Replace entire items array
			var items []objects.OrderedItem
			if err := masscbor.Unmarshal(patch.Value, &items); err != nil {
				return fmt.Errorf("failed to unmarshal items: %w", err)
			}
			order.Items = items

		case nFields >= 2:
			// Get the item index
			index, err := indexFromAny(patch.Path.Fields[1], len(order.Items))
			if err != nil {
				return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
			}

			if nFields == 2 {
				// Replace entire item at index
				var item objects.OrderedItem
				if err := masscbor.Unmarshal(patch.Value, &item); err != nil {
					return fmt.Errorf("failed to unmarshal item: %w", err)
				}
				order.Items[index] = item
			} else if patch.Path.Fields[2] == "Quantity" {
				// Replace just the newValue
				var newValue uint32
				if err := masscbor.Unmarshal(patch.Value, &newValue); err != nil {
					return fmt.Errorf("failed to unmarshal quantity: %w", err)
				}
				order.Items[index].Quantity = newValue
			} else {
				return fmt.Errorf("unsupported field: %s", patch.Path.Fields[2])
			}
		default:
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}

	case "InvoiceAddress":
		if order.InvoiceAddress == nil {
			return fmt.Errorf("InvoiceAddress not set")
		}

		switch {
		case nFields == 1:
			var newAddress objects.AddressDetails
			if err := masscbor.Unmarshal(patch.Value, &newAddress); err != nil {
				return fmt.Errorf("failed to unmarshal invoice address: %w", err)
			}
			order.InvoiceAddress = &newAddress
		case nFields == 2:
			switch patch.Path.Fields[1] {
			case "Name":
				var newName string
				if err := masscbor.Unmarshal(patch.Value, &newName); err != nil {
					return fmt.Errorf("failed to unmarshal invoice address Name: %w", err)
				}
				order.InvoiceAddress.Name = newName
			default:
				return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
			}
		default:
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}

	case "ShippingAddress":
		if order.ShippingAddress == nil {
			return fmt.Errorf("ShippingAddress not set")
		}
		switch {
		case nFields == 1:
			var newAddress objects.AddressDetails
			if err := masscbor.Unmarshal(patch.Value, &newAddress); err != nil {
				return fmt.Errorf("failed to unmarshal ShippingAddress: %w", err)
			}
			order.ShippingAddress = &newAddress
		default:
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
	}
	if err := p.validateOrderReferences(order); err != nil {
		return err
	}
	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func (p *Patcher) removeOrderField(order *objects.Order, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "Items":
		if order.State >= objects.OrderStateCommitted {
			return errCannotModdifyCommittedOrder
		}
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid items path")
		}
		index, err := indexFromAny(patch.Path.Fields[1], len(order.Items))
		if err != nil {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		order.Items = slices.Delete(order.Items, index, index+1)

	case "ShippingAddress":
		if order.ShippingAddress == nil {
			return fmt.Errorf("ShippingAddress not set")
		}
		order.ShippingAddress = nil

	case "InvoiceAddress":
		if order.InvoiceAddress == nil {
			return fmt.Errorf("InvoiceAddress not set")
		}
		order.InvoiceAddress = nil

	default:
		return fmt.Errorf("cannot remove field: %s", patch.Path.Fields[0])
	}

	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func (p *Patcher) modifyOrderQuantity(order *objects.Order, patch Patch) error {
	if order.State >= objects.OrderStateCommitted {
		return errCannotModdifyCommittedOrder
	}
	index, err := checkPathAndIndex(order, patch.Path.Fields)
	if err != nil {
		return err
	}

	var value uint32
	if err := masscbor.Unmarshal(patch.Value, &value); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	if patch.Op == IncrementOp {
		order.Items[index].Quantity += value
	} else {
		if value > order.Items[index].Quantity {
			return fmt.Errorf("cannot decrement below zero")
		}
		order.Items[index].Quantity -= value
	}

	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func checkPathAndIndex(existing *objects.Order, fields []any) (int, error) {
	if len(fields) != 3 || fields[0] != "Items" || fields[2] != "Quantity" {
		return 0, fmt.Errorf("incr/decr only works on path: [Items, x, Quantity]")
	}
	index, err := indexFromAny(fields[1], len(existing.Items))
	if err != nil {
		return 0, err
	}
	return index, nil
}
