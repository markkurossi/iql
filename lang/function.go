//
// Copyright (c) 2021 Markku Rossi
//
// All rights reserved.
//

package lang

import (
	"fmt"

	"github.com/markkurossi/iql/types"
)

// Function implements a function.
type Function struct {
	Name         string
	Args         []FunctionArg
	RetType      types.Type
	Ret          Expr
	Impl         FunctionImpl
	MinArgs      int
	MaxArgs      int
	FirstBound   int
	IsIdempotent IsIdempotent
}

func (f *Function) String() string {
	if f.Impl != nil {
		return fmt.Sprintf("builtin %s", f.Name)
	}
	return fmt.Sprintf("%s(%v) %s", f.Name, f.Args, f.RetType)
}

// FunctionArg defines function arguments for user-defined
// functions. Builtin functions verify function parameter types
// dynamically.
type FunctionArg struct {
	Name string
	Type types.Type
}

// FunctionImpl implements the built-in IQL functions.
type FunctionImpl func(args []Expr, row *Row, rows []*Row) (types.Value, error)

// IsIdempotent tests if the function is idempotent when applied to
// its arguments.
type IsIdempotent func(args []Expr) bool

func idempotentTrue(args []Expr) bool {
	return true
}

func idempotentFalse(args []Expr) bool {
	return false
}

func idempotentArgs(args []Expr) bool {
	for _, arg := range args {
		if !arg.IsIdempotent() {
			return false
		}
	}
	return true
}

func createFunction(f *Function) error {
	_, ok := builtInsByName[f.Name]
	if ok {
		return fmt.Errorf("function already defined: %s", f.Name)
	}
	builtInsByName[f.Name] = f
	return nil
}

func dropFunction(name string, ifExists bool) error {
	f, ok := builtInsByName[name]
	if !ok {
		if ifExists {
			return nil
		}
		return fmt.Errorf("unknown function: %s", name)
	}
	if f.Impl != nil {
		return fmt.Errorf("can't drop builtin function: %s", name)
	}
	delete(builtInsByName, name)
	return nil
}
