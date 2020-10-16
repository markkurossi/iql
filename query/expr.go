//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"

	"github.com/markkurossi/iql/data"
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
	Eval(row []data.Row, columns [][]data.ColumnSelector, rows [][]data.Row) (
		data.Value, error)
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

var functionTypes = map[FunctionType]string{
	FuncSum: "SUM",
}

var functions = map[string]FunctionType{
	"SUM": FuncSum,
}

func (t FunctionType) String() string {
	name, ok := functionTypes[t]
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
func (f *Function) Eval(row []data.Row, columns [][]data.ColumnSelector,
	rows [][]data.Row) (data.Value, error) {

	if len(f.Arguments) != 1 {
		return nil, fmt.Errorf("%s: expected one argument, got %d",
			f.Type, len(f.Arguments))
	}
	switch f.Type {
	case FuncSum:
		var intSum int64
		var floatSum float64

		for _, sumRow := range rows {
			val, err := f.Arguments[0].Eval(sumRow, columns, nil)
			if err != nil {
				return nil, err
			}
			switch v := val.(type) {
			case data.IntValue:
				add, err := v.Int()
				if err != nil {
					return nil, err
				}
				intSum += add
			case data.FloatValue:
				add, err := v.Float()
				if err != nil {
					return nil, err
				}
				floatSum += add
			default:
				return nil, fmt.Errorf("sum over %T", val)
			}
		}
		if floatSum != 0 {
			return data.FloatValue(floatSum), nil
		}
		return data.IntValue(intSum), nil

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
	BinEq BinaryType = iota
	BinNeq
	BinLt
	BinLe
	BinGt
	BinGe
	BinAnd
	BinMult
	BinDiv
	BinAdd
	BinSub
)

var binaries = map[BinaryType]string{
	BinEq:   "=",
	BinNeq:  "<>",
	BinLt:   "<",
	BinLe:   "<=",
	BinGt:   ">",
	BinGe:   ">=",
	BinAnd:  "AND",
	BinMult: "*",
	BinDiv:  "/",
	BinAdd:  "+",
	BinSub:  "-",
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
func (b *Binary) Eval(row []data.Row, columns [][]data.ColumnSelector,
	rows [][]data.Row) (data.Value, error) {

	left, err := b.Left.Eval(row, columns, rows)
	if err != nil {
		return nil, err
	}
	right, err := b.Right.Eval(row, columns, rows)
	if err != nil {
		return nil, err
	}

	// Resolve operation type.

	var opType data.ColumnType

	switch left.(type) {
	case data.BoolValue:
		switch right.(type) {
		case data.BoolValue:
			opType = data.ColumnBool
		default:
			return nil,
				fmt.Errorf("invalid types: %s(%T) %s %s(%T)",
					left, left, b.Type, right, right)
		}

	case data.IntValue:
		switch right.(type) {
		case data.IntValue:
			opType = data.ColumnInt
		case data.FloatValue:
			opType = data.ColumnFloat
		default:
			return nil,
				fmt.Errorf("invalid types: %s(%T) %s %s(%T)",
					left, left, b.Type, right, right)
		}

	case data.FloatValue:
		switch right.(type) {
		case data.IntValue, data.FloatValue:
			opType = data.ColumnFloat
		default:
			return nil,
				fmt.Errorf("invalid types: %s(%T) %s %s(%T)",
					left, left, b.Type, right, right)
		}

	case data.StringValue:
		switch right.(type) {
		case data.StringValue:
			opType = data.ColumnString
		default:
			return nil,
				fmt.Errorf("invalid types: %s(%T) %s %s(%T)",
					left, left, b.Type, right, right)
		}
	}

	switch opType {
	case data.ColumnBool:
		l, err := left.Bool()
		if err != nil {
			return nil, err
		}
		r, err := right.Bool()
		if err != nil {
			return nil, err
		}
		switch b.Type {
		case BinEq:
			return data.BoolValue(l == r), nil
		case BinNeq:
			return data.BoolValue(l != r), nil
		case BinAnd:
			return data.BoolValue(l && r), nil
		default:
			return nil, fmt.Errorf("unknown bool binary expression: %s %s %s",
				left, b.Type, right)
		}

	case data.ColumnInt:
		l, err := left.Int()
		if err != nil {
			return nil, err
		}
		r, err := right.Int()
		if err != nil {
			return nil, err
		}
		switch b.Type {
		case BinEq:
			return data.BoolValue(l == r), nil
		case BinNeq:
			return data.BoolValue(l != r), nil
		case BinLt:
			return data.BoolValue(l < r), nil
		case BinLe:
			return data.BoolValue(l <= r), nil
		case BinGt:
			return data.BoolValue(l > r), nil
		case BinGe:
			return data.BoolValue(l >= r), nil
		case BinMult:
			return data.IntValue(l * r), nil
		case BinDiv:
			return data.IntValue(l / r), nil
		case BinAdd:
			return data.IntValue(l + r), nil
		case BinSub:
			return data.IntValue(l - r), nil
		default:
			return nil, fmt.Errorf("unknown int binary expression: %s %s %s",
				left, b.Type, right)
		}

	case data.ColumnFloat:
		l, err := left.Float()
		if err != nil {
			return nil, err
		}
		r, err := right.Float()
		if err != nil {
			return nil, err
		}
		switch b.Type {
		case BinEq:
			return data.BoolValue(l == r), nil
		case BinNeq:
			return data.BoolValue(l != r), nil
		case BinLt:
			return data.BoolValue(l < r), nil
		case BinGt:
			return data.BoolValue(l > r), nil
		case BinMult:
			return data.FloatValue(l * r), nil
		case BinDiv:
			return data.FloatValue(l / r), nil
		default:
			return nil, fmt.Errorf("unknown float binary expression: %s %s %s",
				left, b.Type, right)
		}

	case data.ColumnString:
		l := left.String()
		r := right.String()
		switch b.Type {
		case BinEq:
			return data.BoolValue(l == r), nil
		case BinNeq:
			return data.BoolValue(l != r), nil
		case BinLt:
			return data.BoolValue(l < r), nil
		case BinGt:
			return data.BoolValue(l > r), nil
		default:
			return nil, fmt.Errorf("unknown string binary expression: %s %s %s",
				left, b.Type, right)
		}

	default:
		return nil,
			fmt.Errorf("invalid types: %s(%T) %s %s(%T)",
				left, left, b.Type, right, right)
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
func (c *Constant) Eval(row []data.Row, columns [][]data.ColumnSelector,
	rows [][]data.Row) (data.Value, error) {

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
func (ref *Reference) Eval(row []data.Row, columns [][]data.ColumnSelector,
	rows [][]data.Row) (data.Value, error) {

	col := row[ref.index.source][ref.index.column]
	t := columns[ref.index.source][ref.index.column].Type

	switch t {
	case data.ColumnBool:
		return col.Bool()
	case data.ColumnInt:
		return col.Int()
	case data.ColumnFloat:
		return col.Float()
	default:
		return data.StringValue(col.String()), nil
	}
}
