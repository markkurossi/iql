//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package types

import (
	"fmt"

	"github.com/markkurossi/tabulate"
)

// Type specifies language types.
type Type int

// Types.
const (
	Bool Type = iota
	Int
	Float
	String
)

// Literal values.
const (
	True  = "true"
	False = "false"
)

var types = map[Type]string{
	Bool:   "bool",
	Int:    "int",
	Float:  "float",
	String: "string",
}

func (t Type) String() string {
	name, ok := types[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{Type %d}", t)
}

// Align returns the type specific column alignment type.
func (t Type) Align() tabulate.Align {
	if t == String {
		return tabulate.ML
	}
	return tabulate.MR
}
