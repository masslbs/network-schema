// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"
	"slices"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/objects"
)

func (p *Patcher) patchAccount(patch Patch) error {
	if patch.Path.AccountAddr == nil {
		return fmt.Errorf("account patch needs an address")
	}

	acc, exists := p.shop.Accounts.Get(patch.Path.AccountAddr.Address[:])

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if exists {
				return fmt.Errorf("account already exists")
			}
			var newAcc objects.Account
			if err := masscbor.Unmarshal(patch.Value, &newAcc); err != nil {
				return fmt.Errorf("failed to unmarshal account: %w", err)
			}
			if err := p.validator.Struct(newAcc); err != nil {
				return err
			}
			return p.shop.Accounts.Insert(patch.Path.AccountAddr.Address[:], newAcc)
		} else if len(patch.Path.Fields) == 2 && patch.Path.Fields[0] == "KeyCards" {
			if !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeAccount, Path: patch.Path}
			}

			var keyCard objects.PublicKey
			if err := masscbor.Unmarshal(patch.Value, &keyCard); err != nil {
				return fmt.Errorf("failed to unmarshal keyCard: %w", err)
			}

			i, err := indexFromAny(patch.Path.Fields[1], len(acc.KeyCards))
			if err != nil {
				return fmt.Errorf("invalid keyCards index: %w", err)
			}
			acc.KeyCards = slices.Insert(acc.KeyCards, i, keyCard)

			return p.shop.Accounts.Insert(patch.Path.AccountAddr.Address[:], acc)
		}
		return fmt.Errorf("add operation not supported for account fields")

	case AppendOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeAccount, Path: patch.Path}
		}

		if len(patch.Path.Fields) != 1 || patch.Path.Fields[0] != "KeyCards" {
			return fmt.Errorf("unsupported field for append operation: %v", patch.Path.Fields)
		}

		var keyCard objects.PublicKey
		if err := masscbor.Unmarshal(patch.Value, &keyCard); err != nil {
			return fmt.Errorf("failed to unmarshal keyCard: %w", err)
		}

		acc.KeyCards = append(acc.KeyCards, keyCard)
		return p.shop.Accounts.Insert(patch.Path.AccountAddr.Address[:], acc)

	case RemoveOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeAccount, Path: patch.Path}
		}

		if len(patch.Path.Fields) == 0 {
			return p.shop.Accounts.Delete(patch.Path.AccountAddr.Address[:])
		}

		if len(patch.Path.Fields) != 2 || patch.Path.Fields[0] != "KeyCards" {
			return fmt.Errorf("invalid KeyCards path")
		}

		i, err := indexFromAny(patch.Path.Fields[1], len(acc.KeyCards))
		if err != nil {
			return fmt.Errorf("invalid keyCards index: %w", err)
		}

		acc.KeyCards = slices.Delete(acc.KeyCards, i, i+1)
		return p.shop.Accounts.Insert(patch.Path.AccountAddr.Address[:], acc)

	case ReplaceOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeAccount, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			var newAcc objects.Account
			if err := masscbor.Unmarshal(patch.Value, &newAcc); err != nil {
				return fmt.Errorf("failed to unmarshal account: %w", err)
			}
			if err := p.validator.Struct(newAcc); err != nil {
				return err
			}
			return p.shop.Accounts.Insert(patch.Path.AccountAddr.Address[:], newAcc)
		}
		return fmt.Errorf("replace operation not supported for account fields")
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
}
