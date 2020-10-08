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
	Eval(row []data.Row, rows [][]data.Row) (data.Value, error)
	String() string
}

// Function implements function expressions.
type Function struct {
	Type      FunctionType
	Arguments []Expr
}

// FunctionType specifies built-in functions.
type FunctionType int

// Built-in functions.
const (
	FuncSum FunctionType = iota
)

var functions = map[FunctionType]string{
	FuncSum: "sum",
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
		err := arg.Bind(sql)
		if err != nil {
			return err
		}
	}
	return nil
}

// Eval implements the Expr.Eval().
func (f *Function) Eval(row []data.Row, rows [][]data.Row) (data.Value, error) {
	if len(f.Arguments) != 1 {
		return nil, fmt.Errorf("%s: expected one argument, got %d",
			f.Type, len(f.Arguments))
	}
	_, err := f.Arguments[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}

	switch f.Type {
	case FuncSum:
		return nil, fmt.Errorf("%s not implemented yet", f.Type)

	default:
		return nil, fmt.Errorf("unknown function: %v", f.Type)
	}
}

func (f *Function) String() string {
	return fmt.Sprintf("%s(%q)", f.Type, f.Arguments)
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
func (b *Binary) Eval(row []data.Row, rows [][]data.Row) (data.Value, error) {
	left, err := b.Left.Eval(row, rows)
	if err != nil {
		return nil, err
	}
	right, err := b.Right.Eval(row, rows)
	if err != nil {
		return nil, err
	}

	switch l := left.(type) {
	case data.BoolValue:
		r, err := right.Bool()
		if err != nil {
			return nil, err
		}
		switch b.Type {
		case BinEQ:
			return data.BoolValue(bool(l) == r), nil
		case BinNEQ:
			return data.BoolValue(bool(l) != r), nil
		case BinAND:
			return data.BoolValue(bool(l) && r), nil
		default:
			return nil, fmt.Errorf("unknown binary expression: %s", b.Type)
		}

	case data.IntValue:
		r, err := right.Int()
		if err != nil {
			return nil, err
		}
		switch b.Type {
		case BinEQ:
			return data.BoolValue(int64(l) == r), nil
		case BinNEQ:
			return data.BoolValue(int64(l) != r), nil
		case BinLT:
			return data.BoolValue(int64(l) < r), nil
		case BinGT:
			return data.BoolValue(int64(l) > r), nil
		default:
			return nil, fmt.Errorf("unknown binary expression: %s", b.Type)
		}

	case data.StringValue:
		r := right.String()
		switch b.Type {
		case BinEQ:
			return data.BoolValue(string(l) == r), nil
		case BinNEQ:
			return data.BoolValue(string(l) != r), nil
		case BinLT:
			return data.BoolValue(string(l) < r), nil
		case BinGT:
			return data.BoolValue(string(l) > r), nil
		default:
			return nil, fmt.Errorf("unknown binary expression: %s", b.Type)
		}

	default:
		return nil,
			fmt.Errorf("invalid types: %s %s %s", left, b.Type, right)
	}
}

func (b *Binary) String() string {
	return fmt.Sprintf("%s %s %s", b.Left, b.Type, b.Right)
}

// Constant implements contant expressions.
type Constant struct {
	Value data.Value
}

// Bind implements the Expr.Bind().
func (c *Constant) Bind(sql *Query) error {
	return nil
}

// Eval implements the Expr.Eval().
func (c *Constant) Eval(row []data.Row, rows [][]data.Row) (data.Value, error) {
	return c.Value, nil
}

func (c *Constant) String() string {
	return c.Value.String()
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
func (ref *Reference) Eval(row []data.Row, rows [][]data.Row) (
	data.Value, error) {

	col := row[ref.index.source][ref.index.column]

	return data.StringValue(col.String()), nil
}
