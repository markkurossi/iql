//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"

	"github.com/markkurossi/iql/types"
)

// Scope implements name scope.
type Scope struct {
	Parent  *Scope
	Symbols map[string]*Binding
}

// NewScope creates a new name scope.
func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent:  parent,
		Symbols: make(map[string]*Binding),
	}
}

// Get gets the name binding from the scope.
func (scope *Scope) Get(name string) *Binding {
	for s := scope; s != nil; s = s.Parent {
		b, ok := s.Symbols[name]
		if ok {
			return b
		}
	}
	return nil
}

// Declare declares the name with type.
func (scope *Scope) Declare(name string, t types.Type) {
	scope.Symbols[name] = &Binding{
		Type:  t,
		Value: types.Null,
	}
}

// Set sets the binding for the name.
func (scope *Scope) Set(name string, v types.Value) error {
	for s := scope; s != nil; s = s.Parent {
		b, ok := s.Symbols[name]
		if ok {
			// Set new binding for this scope.
			if !b.Type.CanAssign(v) {
				return fmt.Errorf("can't assign '%s' to '%s' variable",
					v, b.Type)
			}
			b.Value = v
			return nil
		}
	}
	return fmt.Errorf("unknown identifier '%s'", name)
}

// Binding symbol binding.
type Binding struct {
	Type  types.Type
	Value types.Value
}
