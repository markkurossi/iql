//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"
	"strings"

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

// Declare declares the name with type.
func (scope *Scope) Declare(name string, t types.Type) error {
	name = strings.ToUpper(name)

	b := scope.Get(name)
	if b != nil {
		return fmt.Errorf("identifier '%s' already declared", name)
	}
	scope.Symbols[name] = &Binding{
		Type:  t,
		Value: types.Null,
	}

	return nil
}

// Set sets the binding for the name.
func (scope *Scope) Set(name string, v types.Value) error {
	name = strings.ToUpper(name)

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

// Get gets the name binding from the scope.
func (scope *Scope) Get(name string) *Binding {
	name = strings.ToUpper(name)

	for s := scope; s != nil; s = s.Parent {
		b, ok := s.Symbols[name]
		if ok {
			return b
		}
	}
	return nil
}

// Binding symbol binding.
type Binding struct {
	Type  types.Type
	Value types.Value
}
