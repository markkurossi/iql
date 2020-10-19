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
	Table
)

// Literal values.
const (
	True  = "true"
	False = "false"
)

var types = map[Type]string{
	Bool:   "boolean",
	Int:    "integer",
	Float:  "real",
	String: "varchar",
	Table:  "table",
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

// CanAssign tests if the argument value can be assigned into a
// variable this type.
func (t Type) CanAssign(v Value) bool {
	switch v.(type) {
	case BoolValue:
		return t == Bool
	case IntValue:
		return t == Int || t == Float
	case FloatValue:
		return t == Int || t == Float
	case StringValue:
		return t == String
	case TableValue:
		return t == Table
	case NullValue:
		return true
	default:
		return false
	}
}
