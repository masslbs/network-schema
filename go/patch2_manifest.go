package schema

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/fxamacker/cbor/v2"
)

// GetOpHandlers returns a map from field-name to PatchHandler, implementing
// add/replace/remove for each top-level field in the Manifest.
func (m *Manifest) GetOpHandlers() map[string]PatchHandler {
	return map[string]PatchHandler{
		"payees": {
			// Add will merge in new payees
			Add: func(value cbor.RawMessage) error {
				var incoming map[string]Payee
				if err := Unmarshal(value, &incoming); err != nil {
					return fmt.Errorf("failed to unmarshal payees to add: %w", err)
				}
				for k, v := range incoming {
					if _, exists := m.Payees[k]; exists {
						return fmt.Errorf("payee %q already exists", k)
					}
					m.Payees[k] = v
				}
				return nil
			},
			// Replace will completely overwrite the map
			Replace: func(value cbor.RawMessage) error {
				var replaced map[string]Payee
				if err := Unmarshal(value, &replaced); err != nil {
					return fmt.Errorf("failed to unmarshal payees for replace: %w", err)
				}
				for k := range m.Payees {
					delete(m.Payees, k)
				}
				for k, v := range replaced {
					m.Payees[k] = v
				}
				return nil
			},
			// Remove clears the entire map
			Remove: func() error {
				for k := range m.Payees {
					delete(m.Payees, k)
				}
				return nil
			},
			GetSubHandler: m.Payees.GetSubHandler,
		},
		"shippingRegions": {
			Add: func(value cbor.RawMessage) error {
				return fmt.Errorf("specify index to add to shippingRegions")
			},
			Replace: func(value cbor.RawMessage) error {
				var incoming map[string]ShippingRegion
				if err := Unmarshal(value, &incoming); err != nil {
					return fmt.Errorf("failed to unmarshal shipping regions for replace: %w", err)
				}
				m.ShippingRegions = incoming
				return nil
			},
			Remove: func() error {
				m.ShippingRegions = make(ShippingRegions)
				return nil
			},
			GetSubHandler: m.ShippingRegions.GetSubHandler,
		},
		"acceptedCurrencies": {
			Add: func(value cbor.RawMessage) error {
				return fmt.Errorf("specify index to add to acceptedCurrencies")
			},
			Replace: func(value cbor.RawMessage) error {
				var replaced ChainAddresses
				if err := Unmarshal(value, &replaced); err != nil {
					return fmt.Errorf("failed to unmarshal chain addresses for replace: %w", err)
				}
				m.AcceptedCurrencies = replaced
				return nil
			},
			Remove: func() error {
				m.AcceptedCurrencies = nil
				return nil
			},
			GetSubHandler: m.AcceptedCurrencies.GetSubHandler,
		},
		"pricingCurrency": {
			// Typically, only “replace” makes sense. We can define add or remove if we want:
			Replace: func(value cbor.RawMessage) error {
				var cur ChainAddress
				if err := Unmarshal(value, &cur); err != nil {
					return fmt.Errorf("failed to unmarshal pricingCurrency: %w", err)
				}
				m.PricingCurrency = cur
				return nil
			},
			// Example of leaving Add/Remove nil if not supported:
			Add:    nil,
			Remove: nil,
		},
	}
}

// Payees is referenced in patch2_manifest.go, lines like m.Payees.GetPatchHandlers()[""]
// or m.Payees.GetSubHandler(...). So we need to implement them:

// GetSubHandler allows patching an individual payee by “payeeKey”.
func (ps Payees) GetSubHandler(payeeKey string) (*PatchHandler, error) {
	_, exists := ps[payeeKey]
	// Return a PatchHandler that can patch subfields of this single Payee
	return &PatchHandler{
		Replace: func(value cbor.RawMessage) error {
			if !exists {
				return fmt.Errorf("payee not found: %s", payeeKey)
			}
			var newPayee Payee
			if err := Unmarshal(value, &newPayee); err != nil {
				return fmt.Errorf("failed to unmarshal payee: %w", err)
			}
			ps[payeeKey] = newPayee
			return nil
		},
		Remove: func() error {
			if !exists {
				return fmt.Errorf("payee not found: %s", payeeKey)
			}
			delete(ps, payeeKey)
			return nil
		},
		Add: func(value cbor.RawMessage) error {
			if exists {
				return fmt.Errorf("payee %s already exists", payeeKey)
			}
			var newPayee Payee
			if err := Unmarshal(value, &newPayee); err != nil {
				return fmt.Errorf("failed to unmarshal payee: %w", err)
			}
			ps[payeeKey] = newPayee
			return nil
		},
		// If you want deeper subfields (like “Address” or “CallAsContract”) you can define:
		GetSubHandler: func(field string) (*PatchHandler, error) {
			return nil, fmt.Errorf("unsupported subfield on Payee: %q", field)
		},
	}, nil
}

func (sr ShippingRegions) GetSubHandler(regionKey string) (*PatchHandler, error) {
	_, exists := sr[regionKey]
	return &PatchHandler{
		Add: func(value cbor.RawMessage) error {
			if exists {
				return fmt.Errorf("shipping region already exists: %s", regionKey)
			}
			var newRegion ShippingRegion
			if err := Unmarshal(value, &newRegion); err != nil {
				return fmt.Errorf("failed to unmarshal shipping region: %w", err)
			}
			sr[regionKey] = newRegion
			return nil
		},
		Replace: func(value cbor.RawMessage) error {
			if !exists {
				return fmt.Errorf("shipping region not found: %s", regionKey)
			}
			var newRegion ShippingRegion
			if err := Unmarshal(value, &newRegion); err != nil {
				return fmt.Errorf("failed to unmarshal shipping region: %w", err)
			}
			sr[regionKey] = newRegion
			return nil
		},
		Remove: func() error {
			if !exists {
				return fmt.Errorf("shipping region not found: %s", regionKey)
			}
			delete(sr, regionKey)
			return nil
		},
	}, nil
}

func (ac *ChainAddresses) GetSubHandler(indexStr string) (*PatchHandler, error) {
	var index int
	if indexStr == "-" {
		// append
		index = len(*ac)
	} else {
		var err error
		index, err = strconv.Atoi(indexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid index: %w", err)
		}
		if index < 0 || index >= len(*ac) {
			return nil, fmt.Errorf("index out of bounds: %d", index)
		}
	}
	return &PatchHandler{
		Add: func(value cbor.RawMessage) error {
			var newCurrency ChainAddress
			if err := Unmarshal(value, &newCurrency); err != nil {
				return fmt.Errorf("failed to unmarshal chain address: %w", err)
			}
			if index == len(*ac) {
				*ac = append(*ac, newCurrency)
			} else {
				*ac = slices.Insert(*ac, index, newCurrency)
			}
			return nil
		},
		Replace: func(value cbor.RawMessage) error {
			var newCurrency ChainAddress
			if err := Unmarshal(value, &newCurrency); err != nil {
				return fmt.Errorf("failed to unmarshal chain address: %w", err)
			}
			(*ac)[index] = newCurrency
			return nil
		},
		Remove: func() error {
			*ac = slices.Delete(*ac, index, index+1)
			return nil
		},
	}, nil
}
