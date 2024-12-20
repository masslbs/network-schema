package schema

import (
	"fmt"
	"math/big"

	"github.com/fxamacker/cbor/v2"
)

var bigZero = big.NewInt(0)

func (l *Listing) GetOpHandlers() map[string]PatchHandler {
	return map[string]PatchHandler{
		"options": {
			// Original handlers for top-level options operations
			Add: func(value cbor.RawMessage) error {
				var incoming map[string]ListingOption
				if err := Unmarshal(value, &incoming); err != nil {
					return err
				}
				for k, v := range incoming {
					if _, exists := l.Options[k]; exists {
						return fmt.Errorf("option %q already exists", k)
					}
					l.Options[k] = v
				}
				return nil
			},
			Replace: func(value cbor.RawMessage) error {
				var replaced map[string]ListingOption
				if err := Unmarshal(value, &replaced); err != nil {
					return err
				}
				l.Options = replaced
				return nil
			},
			Remove: func() error {
				l.Options = make(map[string]ListingOption)
				return nil
			},

			// New handler for nested operations
			GetSubHandler: func(optionKey string) (*PatchHandler, error) {
				// option, exists := l.Options[optionKey]
				// if !exists {
				// 	// For Add operations, we'll create the option
				// 	l.Options[optionKey] = ListingOption{
				// 		Variations: make(map[string]ListingVariation),
				// 	}
				// 	option = l.Options[optionKey]
				// }

				return &PatchHandler{
					// Handle direct operations on a specific option
					Replace: func(value cbor.RawMessage) error {
						var newOption ListingOption
						if err := Unmarshal(value, &newOption); err != nil {
							return err
						}
						l.Options[optionKey] = newOption
						return nil
					},
					Remove: func() error {
						delete(l.Options, optionKey)
						return nil
					},

					// Handle nested fields within an option
					GetSubHandler: func(field string) (*PatchHandler, error) {
						switch field {
						case "title":
							return &PatchHandler{
								Replace: func(value cbor.RawMessage) error {
									var title string
									if err := Unmarshal(value, &title); err != nil {
										return err
									}
									opt := l.Options[optionKey]
									opt.Title = title
									l.Options[optionKey] = opt
									return nil
								},
							}, nil
						}
						return nil, fmt.Errorf("unsupported option field: %s", field)
					},
				}, nil
			},
		},
		// ... other fields ...
	}
}

// Let ListingMetadata be patchable. If leftover fields are empty, we add/replace/remove the entire struct.
// If leftover fields remain, we can define sub-fields in GetPatchHandlers to handle "title", "images", etc.
func (m *ListingMetadata) GetPatchHandlers() map[string]PatchHandler {
	return map[string]PatchHandler{
		// If leftover is empty => entire struct add/replace/remove.
		// But if leftover is non-empty => each field is a subobject or direct field.

		// For "title" we can do direct add/replace/remove:
		"title": {
			Add: func(value cbor.RawMessage) error {
				if m.Title != "" {
					return fmt.Errorf("title already set, cannot add again")
				}
				var newTitle string
				if err := Unmarshal(value, &newTitle); err != nil {
					return err
				}
				m.Title = newTitle
				return nil
			},
			Replace: func(value cbor.RawMessage) error {
				var newTitle string
				if err := Unmarshal(value, &newTitle); err != nil {
					return err
				}
				m.Title = newTitle
				return nil
			},
			Remove: func() error {
				m.Title = ""
				return nil
			},
		},

		"images": {
			Add: func(value cbor.RawMessage) error {
				// Could parse a single string or a list of strings
				var one string
				if err := Unmarshal(value, &one); err == nil {
					m.Images = append(m.Images, one)
					return nil
				}
				var many []string
				if err := Unmarshal(value, &many); err != nil {
					return err
				}
				m.Images = append(m.Images, many...)
				return nil
			},
			Replace: func(value cbor.RawMessage) error {
				var newArr []string
				if err := Unmarshal(value, &newArr); err != nil {
					return err
				}
				m.Images = newArr
				return nil
			},
			Remove: func() error {
				m.Images = nil
				return nil
			},
		},
	}
}

func (v *ListingVariation) GetPatchHandlers() map[string]PatchHandler {
	return map[string]PatchHandler{
		"variationInfo": {
			Replace: func(value cbor.RawMessage) error {
				var newVariationInfo ListingMetadata
				if err := Unmarshal(value, &newVariationInfo); err != nil {
					return err
				}
				v.VariationInfo = newVariationInfo
				return nil
			},
		},
	}
}
