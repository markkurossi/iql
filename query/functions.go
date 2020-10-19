//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"
)

// Function implements a function.
type Function struct {
	Type       FunctionType
	MinArgs    int
	MaxArgs    int
	Idempotent bool

	// Name of a composite function. This is empty for built-in
	// functions
	Name string
}

func (f *Function) String() string {
	if f.Type == FuncComposite {
		return f.Name
	}
	return f.Type.String()
}

// FunctionType specifies function types and built-in functions.
type FunctionType int

// Function types.
const (
	FuncSum FunctionType = iota
	FuncAvg
	FuncComposite
)

var functionTypes = map[FunctionType]string{
	FuncSum:       "SUM",
	FuncAvg:       "AVG",
	FuncComposite: "{composite}",
}

func (t FunctionType) String() string {
	name, ok := functionTypes[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{function %d}", t)
}

var builtIns = map[string]*Function{
	"SUM": {
		Type:       FuncSum,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
	"AVG": {
		Type:       FuncAvg,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
}

func builtIn(name string) *Function {
	return builtIns[name]
}
