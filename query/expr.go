//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"

	"github.com/markkurossi/htmlq/data"
)

var (
	_ Expr = &Function{}
	_ Expr = &Binary{}
	_ Expr = &Constant{}
	_ Expr = &Reference{}
)

// Expr implements expressions.
type Expr interface {
	Bind(sql *Query) error
	Eval(sources []data.Row) (Value, error)
}

// Function implements function expressions.
type Function struct {
	Type       FunctionType
	Arguments  []string
	References []*Reference
}

// FunctionType specifies built-in functions.
type FunctionType int

// Built-in functions.
const (
	FuncCount FunctionType = iota
)

var functions = map[FunctionType]string{
	FuncCount: "count",
}

func (t FunctionType) String() string {
	name, ok := functions[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{function %d}", t)
}

// Bind implements the Expr.Bind().
func (f *Function) Bind(sql *Query) error {
	for _, arg := range f.Arguments {
		ref, err := sql.resolveName(data.Reference{
			Column: arg,
		}, false)
		if err != nil {
			return err
		}
		f.References = append(f.References, ref)
	}
	return nil
}

// Eval implements the Expr.Eval().
func (f *Function) Eval(sources []data.Row) (Value, error) {
	if len(f.Arguments) != 1 {
		return nil, fmt.Errorf("%s: expected one argument, got %d",
			f.Type, len(f.Arguments))
	}
	ref := f.References[0]
	arg0 := sources[ref.index.source][ref.index.column]

	fmt.Printf("%s(%s/%T)\n", f.Type, arg0, arg0)

	switch f.Type {
	case FuncCount:
		return IntValue(arg0.Count()), nil

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
	BinEQ BinaryType = iota
	BinNEQ
	BinLT
	BinGT
	BinAND
)

var binaries = map[BinaryType]string{
	BinEQ:  "=",
	BinNEQ: "<>",
	BinLT:  "<",
	BinGT:  ">",
	BinAND: "AND",
}

func (t BinaryType) String() string {
	name, ok := binaries[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{binary %d}", t)
}

// Bind implements the Expr.Bind().
func (b *Binary) Bind(sql *Query) error {
	err := b.Left.Bind(sql)
	if err != nil {
		return err
	}
	return b.Right.Bind(sql)
}

// Eval implements the Expr.Eval().
func (b *Binary) Eval(sources []data.Row) (Value, error) {
	left, err := b.Left.Eval(sources)
	if err != nil {
		return nil, err
	}
	right, err := b.Right.Eval(sources)
	if err != nil {
		return nil, err
	}

	switch l := left.(type) {
	case BoolValue:
		r, err := right.Bool()
		if err != nil {
			return nil, err
		}
		switch b.Type {
		case BinEQ:
			return BoolValue(bool(l) == r), nil
		case BinNEQ:
			return BoolValue(bool(l) != r), nil
		case BinAND:
			return BoolValue(bool(l) && r), nil
		default:
			return nil, fmt.Errorf("unknown binary expression: %s", b.Type)
		}

	case IntValue:
		r, err := right.Int()
		if err != nil {
			return nil, err
		}
		switch b.Type {
		case BinEQ:
			return BoolValue(int(l) == r), nil
		case BinNEQ:
			return BoolValue(int(l) != r), nil
		case BinLT:
			return BoolValue(int(l) < r), nil
		case BinGT:
			return BoolValue(int(l) > r), nil
		default:
			return nil, fmt.Errorf("unknown binary expression: %s", b.Type)
		}

	case StringValue:
		r := right.String()
		switch b.Type {
		case BinEQ:
			return BoolValue(string(l) == r), nil
		case BinNEQ:
			return BoolValue(string(l) != r), nil
		case BinLT:
			return BoolValue(string(l) < r), nil
		case BinGT:
			return BoolValue(string(l) > r), nil
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

// Bind implements the Expr.Bind().
func (constant *Constant) Bind(sql *Query) error {
	return nil
}

// Eval implements the Expr.Eval().
func (constant *Constant) Eval(sources []data.Row) (Value, error) {
	return constant.Value, nil
}

// Reference implements column reference expressions.
type Reference struct {
	data.Reference
	index  columnIndex
	public bool
}

// NewReference creates a new reference for the argument name.
func NewReference(name string) (*Reference, error) {
	r, err := data.NewReference(name)
	if err != nil {
		return nil, err
	}
	return &Reference{
		Reference: r,
	}, nil
}

type columnIndex struct {
	source int
	column int
}

func (idx columnIndex) String() string {
	return fmt.Sprintf("%d.%d", idx.source, idx.column)
}

// Bind implements the Expr.Bind().
func (ref *Reference) Bind(sql *Query) error {
	r, err := sql.resolveName(ref.Reference, false)
	if err != nil {
		return err
	}
	ref.index = r.index
	return nil
}

// Eval implements the Expr.Eval().
func (ref *Reference) Eval(sources []data.Row) (Value, error) {
	col := sources[ref.index.source][ref.index.column]

	return StringValue(col.String()), nil
}
