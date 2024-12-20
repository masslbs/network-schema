package schema

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// PatchHandler defines functions for add/replace/remove on one field.
type PatchHandler struct {
	Add     func(value cbor.RawMessage) error
	Replace func(value cbor.RawMessage) error
	Remove  func() error

	// For handling nested paths like "options/color/variations/pink":
	// GetSubHandler lets us navigate into nested structures
	GetSubHandler func(subfield string) (*PatchHandler, error)
}

// Patchable is an optional interface that returns a field->handler map.
type Patchable interface {
	GetOpHandlers() map[string]PatchHandler
}

// PatchField looks up the handler for 'field' in obj.GetPatchHandlers(),
// then either calls the direct operation or recurses into subfields
func PatchField(obj Patchable, op OpString, fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return fmt.Errorf("no fields specified")
	}

	handlers := obj.GetOpHandlers()
	firstField := fields[0]
	handler, ok := handlers[firstField]
	if !ok {
		return fmt.Errorf("unsupported field: %s", firstField)
	}

	// If we have more fields, we need to recurse using GetSubHandler
	if len(fields) > 1 {
		if handler.GetSubHandler == nil {
			return fmt.Errorf("field %q does not support nested operations", firstField)
		}
		subHandler, err := handler.GetSubHandler(fields[1])
		if err != nil {
			return fmt.Errorf("getting subhandler for %q: %w", fields[1], err)
		}
		fmt.Printf("subhandler for field %s: %+v\n", fields[1], subHandler)
		// Recurse with remaining fields
		return patchWithHandler(subHandler, op, fields[2:], value)
	}

	// Otherwise do the direct operation
	switch op {
	case AddOp:
		if handler.Add == nil {
			return fmt.Errorf("add not supported on field %s", firstField)
		}
		return handler.Add(value)
	case ReplaceOp:
		if handler.Replace == nil {
			return fmt.Errorf("replace not supported on field %s", firstField)
		}
		return handler.Replace(value)
	case RemoveOp:
		if handler.Remove == nil {
			return fmt.Errorf("remove not supported on field %s", firstField)
		}
		return handler.Remove()
	default:
		return fmt.Errorf("unsupported op: %s", op)
	}
}

// Helper function to apply operations using a specific handler
func patchWithHandler(handler *PatchHandler, op OpString, fields []string, value cbor.RawMessage) error {
	if len(fields) > 0 {
		if handler.GetSubHandler == nil {
			return fmt.Errorf("no further subfields supported")
		}
		subHandler, err := handler.GetSubHandler(fields[0])
		if err != nil {
			return err
		}
		return patchWithHandler(subHandler, op, fields[1:], value)
	}
	switch op {
	case AddOp:
		if handler.Add == nil {
			return fmt.Errorf("add not supported")
		}
		return handler.Add(value)
	case ReplaceOp:
		if handler.Replace == nil {
			return fmt.Errorf("replace not supported")
		}
		return handler.Replace(value)
	case RemoveOp:
		if handler.Remove == nil {
			return fmt.Errorf("remove not supported")
		}
		return handler.Remove()
	default:
		return fmt.Errorf("unsupported op: %s", op)
	}
}
