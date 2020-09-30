//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"
	"unicode"

	"github.com/markkurossi/tabulate"
)

var (
	_ Expr  = &Function{}
	_ Expr  = &Binary{}
	_ Expr  = &Constant{}
	_ Value = BoolValue(true)
	_ Value = IntValue(42)
)

// ColumnSelector implements data column selector.
type ColumnSelector struct {
	Find  string
	As    string
	Align tabulate.Align
}

// IsPublic reports if the column is public and should be included in
// the result set.
func (col ColumnSelector) IsPublic() bool {
	runes := []rune(col.As)
	return len(runes) > 0 && unicode.IsUpper(runes[0])
}

// Expr implements expressions.
type Expr interface {
	Eval(c Columns) (Value, error)
}

// Value implements expression values.
type Value interface {
	Bool() (bool, error)
	Int() (int, error)
}

// BoolValue implements boolean values.
type BoolValue bool

// Bool implements the Value.Bool().
func (v BoolValue) Bool() (bool, error) {
	return bool(v), nil
}

// Int implements the Value.Int().
func (v BoolValue) Int() (int, error) {
	return 0, fmt.Errorf("bool used as int")
}

// IntValue implements integer values.
type IntValue int

// Bool implements the Value.Bool().
func (v IntValue) Bool() (bool, error) {
	return false, fmt.Errorf("int used as bool")
}

// Int implements the Value.Int().
func (v IntValue) Int() (int, error) {
	return int(v), nil
}

// Function implements function expressions.
type Function struct {
	Type      FunctionType
	Arguments []string
}

// FunctionType specifies built-in functions.
type FunctionType int

// Built-in functions.
const (
	FuncLength FunctionType = iota
)

var functions = map[FunctionType]string{
	FuncLength: "length",
}

func (t FunctionType) String() string {
	name, ok := functions[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{function %d}", t)
}

// Eval implements the Expr.Eval().
func (f *Function) Eval(c Columns) (Value, error) {
	if len(f.Arguments) != 1 {
		return nil, fmt.Errorf("%s: expected one argument, got %d",
			f.Type, len(f.Arguments))
	}
	arg0, ok := c[f.Arguments[0]]
	if !ok {
		return nil, fmt.Errorf("%s: column '%s' not found",
			f.Type, f.Arguments[0])
	}

	switch f.Type {
	case FuncLength:
		return IntValue(arg0.Length()), nil

	default:
		return nil, fmt.Errorf("unknown function: %v", f.Type)
	}
}

// Binary implements binary expressions.
type Binary struct {
	Type  BinaryType
	Left  Expr
	Right Expr
}

// BinaryType specifies binary expression types.
type BinaryType int

// Binary expressions.
const (
	BinLT BinaryType = iota
	BinGT BinaryType = iota
)

var binaries = map[BinaryType]string{
	BinLT: "<",
	BinGT: ">",
}

func (t BinaryType) String() string {
	name, ok := binaries[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{binary %d}", t)
}

// Eval implements the Expr.Eval().
func (b *Binary) Eval(c Columns) (Value, error) {
	left, err := b.Left.Eval(c)
	if err != nil {
		return nil, err
	}
	right, err := b.Right.Eval(c)
	if err != nil {
		return nil, err
	}

	switch l := left.(type) {
	case IntValue:
		r, err := right.Int()
		if err != nil {
			return nil, err
		}
		switch b.Type {
		case BinLT:
			return BoolValue(int(l) < r), nil
		case BinGT:
			return BoolValue(int(l) > r), nil
		default:
			return nil, fmt.Errorf("unknown binary expression: %s", b.Type)
		}

	default:
		return nil,
			fmt.Errorf("invalid types: %s %s %s", left, b.Type, right)
	}
}

// Constant implements contant expressions.
type Constant struct {
	Value Value
}

// Eval implements the Expr.Eval().
func (constant *Constant) Eval(c Columns) (Value, error) {
	return constant.Value, nil
}
