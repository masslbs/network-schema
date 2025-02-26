// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"
	"slices"
	"strconv"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/objects"
)

func (p *Patcher) patchOrder(patch Patch) error {
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("order patch needs an ID")
	}
	objId := *patch.Path.ObjectID
	order, exists := p.shop.Orders.Get(objId)

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if exists {
				return fmt.Errorf("order %d already exists", objId)
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
			return p.shop.Orders.Insert(objId, newOrder)
		}

		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		return p.addOrderField(&order, patch)

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
			return p.shop.Orders.Insert(objId, newOrder)
		}
		return p.replaceOrderField(&order, patch)

	case RemoveOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			return p.shop.Orders.Delete(objId)
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
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: PatchPath{ObjectID: &item.ListingID}}
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
				return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: PatchPath{ObjectID: &item.ListingID, Fields: []string{"options", varID}}}
			}
		}
	}

	if order.ChosenPayee != nil {
		found := false
		for _, payee := range p.shop.Manifest.Payees {
			if payee == *order.ChosenPayee {
				found = true
				break
			}
		}
		if !found {
			return ObjectNotFoundError{
				ObjectType: ObjectTypeOrder,
				Path:       PatchPath{Fields: []string{"chosenPayee", order.ChosenPayee.Address.String()}}}
		}
	}

	if order.ChosenCurrency != nil {
		found := false
		for _, currency := range p.shop.Manifest.AcceptedCurrencies {
			if currency == *order.ChosenCurrency {
				found = true
				break
			}
		}
		if !found {
			return ObjectNotFoundError{
				ObjectType: ObjectTypeManifest,
				Path:       PatchPath{Fields: []string{"acceptedCurrencies", order.ChosenCurrency.String()}}}
		}
	}

	return nil
}

func (p *Patcher) addOrderField(order *objects.Order, patch Patch) error {
	if len(patch.Path.Fields) == 0 {
		return fmt.Errorf("field path required for add operation")
	}

	switch patch.Path.Fields[0] {
	case "items":
		var item objects.OrderedItem
		if err := masscbor.Unmarshal(patch.Value, &item); err != nil {
			return fmt.Errorf("failed to unmarshal order item: %w", err)
		}
		listing, exists := p.shop.Listings.Get(item.ListingID)
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: PatchPath{ObjectID: &item.ListingID}}
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
	case "shippingAddress":
		if order.ShippingAddress != nil {
			return fmt.Errorf("shipping address already set")
		}
		var shippingAddress objects.AddressDetails
		if err := masscbor.Unmarshal(patch.Value, &shippingAddress); err != nil {
			return fmt.Errorf("failed to unmarshal shipping address: %w", err)
		}
		order.ShippingAddress = &shippingAddress
	case "invoiceAddress":
		if order.InvoiceAddress != nil {
			return fmt.Errorf("invoice address already set")
		}
		var invoiceAddress objects.AddressDetails
		if err := masscbor.Unmarshal(patch.Value, &invoiceAddress); err != nil {
			return fmt.Errorf("failed to unmarshal invoice address: %w", err)
		}
		order.InvoiceAddress = &invoiceAddress
	case "paymentDetails":
		if order.PaymentDetails != nil {
			return fmt.Errorf("payment details already set")
		}
		var paymentDetails objects.PaymentDetails
		if err := masscbor.Unmarshal(patch.Value, &paymentDetails); err != nil {
			return fmt.Errorf("failed to unmarshal payment details: %w", err)
		}
		order.PaymentDetails = &paymentDetails
	case "chosenPayee":
		if order.ChosenPayee != nil {
			return fmt.Errorf("chosen payee already set")
		}
		var payee objects.Payee
		if err := masscbor.Unmarshal(patch.Value, &payee); err != nil {
			return fmt.Errorf("failed to unmarshal payee: %w", err)
		}
		order.ChosenPayee = &payee
	case "chosenCurrency":
		if order.ChosenCurrency != nil {
			return fmt.Errorf("chosen currency already set")
		}
		var currency objects.ChainAddress
		if err := masscbor.Unmarshal(patch.Value, &currency); err != nil {
			return fmt.Errorf("failed to unmarshal currency: %w", err)
		}
		order.ChosenCurrency = &currency
	case "txDetails":
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
	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func (p *Patcher) replaceOrderField(order *objects.Order, patch Patch) error {
	nFields := len(patch.Path.Fields)

	switch patch.Path.Fields[0] {
	case "state":
		var state objects.OrderState
		if err := masscbor.Unmarshal(patch.Value, &state); err != nil {
			return fmt.Errorf("failed to unmarshal order state: %w", err)
		}
		order.State = state

	case "chosenPayee":
		var payee objects.Payee
		if err := masscbor.Unmarshal(patch.Value, &payee); err != nil {
			return fmt.Errorf("failed to unmarshal payee: %w", err)
		}
		found := false
		for _, p := range p.shop.Manifest.Payees {
			if p == payee {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("payee not found in manifest payees")
		}
		order.ChosenPayee = &payee

	case "chosenCurrency":
		var currency objects.ChainAddress
		if err := masscbor.Unmarshal(patch.Value, &currency); err != nil {
			return fmt.Errorf("failed to unmarshal currency: %w", err)
		}
		found := false
		for _, c := range p.shop.Manifest.AcceptedCurrencies {
			if c == currency {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("currency not found in accepted currencies")
		}
		order.ChosenCurrency = &currency

	case "paymentDetails":
		var details objects.PaymentDetails
		if err := masscbor.Unmarshal(patch.Value, &details); err != nil {
			return fmt.Errorf("failed to unmarshal payment details: %w", err)
		}
		order.PaymentDetails = &details

	case "txDetails":
		var details objects.OrderPaid
		if err := masscbor.Unmarshal(patch.Value, &details); err != nil {
			return fmt.Errorf("failed to unmarshal tx details: %w", err)
		}
		order.TxDetails = &details

	case "items":

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
			index, err := strconv.Atoi(patch.Path.Fields[1])
			if err != nil {
				return fmt.Errorf("failed to convert index to int: %w", err)
			}
			if index < 0 || index >= len(order.Items) {
				return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
			}

			if nFields == 2 {
				// Replace entire item at index
				var item objects.OrderedItem
				if err := masscbor.Unmarshal(patch.Value, &item); err != nil {
					return fmt.Errorf("failed to unmarshal item: %w", err)
				}
				order.Items[index] = item
			} else if patch.Path.Fields[2] == "quantity" {
				// Replace just the quantity
				var quantity uint32
				if err := masscbor.Unmarshal(patch.Value, &quantity); err != nil {
					return fmt.Errorf("failed to unmarshal quantity: %w", err)
				}
				order.Items[index].Quantity = quantity
			} else {
				return fmt.Errorf("unsupported field: %s", patch.Path.Fields[2])
			}
		default:
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}

	case "invoiceAddress":
		if order.InvoiceAddress == nil {
			return fmt.Errorf("invoice address not set")
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
			case "name":
				var newName string
				if err := masscbor.Unmarshal(patch.Value, &newName); err != nil {
					return fmt.Errorf("failed to unmarshal invoice address name: %w", err)
				}
				order.InvoiceAddress.Name = newName
			default:
				return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
			}
		default:
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}

	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
	}

	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func (p *Patcher) removeOrderField(order *objects.Order, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "items":
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid items path")
		}
		index, err := strconv.Atoi(patch.Path.Fields[1])
		if err != nil {
			return fmt.Errorf("invalid item index: %w", err)
		}
		if index < 0 || index >= len(order.Items) {
			return fmt.Errorf("item index out of bounds: %d", index)
		}

		order.Items = slices.Delete(order.Items, index, index+1)

	case "shippingAddress":
		if order.ShippingAddress == nil {
			return fmt.Errorf("shipping address not set")
		}
		order.ShippingAddress = nil

	case "invoiceAddress":
		if order.InvoiceAddress == nil {
			return fmt.Errorf("invoice address not set")
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

func checkPathAndIndex(existing *objects.Order, fields []string) (int, error) {
	if len(fields) != 3 || fields[0] != "items" || fields[2] != "quantity" {
		return 0, fmt.Errorf("incr/decr only works on path: [items, x, quantity]")
	}
	index, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert index to int: %w", err)
	}
	if index < 0 || index >= len(existing.Items) {
		return 0, fmt.Errorf("index out of bounds: %d", index)
	}
	return index, nil
}
