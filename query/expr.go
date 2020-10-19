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

var (
	_ Expr = &Function{}
	_ Expr = &Binary{}
	_ Expr = &And{}
	_ Expr = &Constant{}
	_ Expr = &Reference{}
)

// Expr implements expressions.
type Expr interface {
	Bind(sql *Query) error
	Eval(row []types.Row, columns [][]types.ColumnSelector, rows [][]types.Row) (
		types.Value, error)
	IsIdempotent() bool
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
	FuncAvg
)

var functionTypes = map[FunctionType]string{
	FuncSum: "SUM",
	FuncAvg: "AVG",
}

var functions = map[string]FunctionType{
	"SUM": FuncSum,
	"AVG": FuncAvg,
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
func (f *Function) Eval(row []types.Row, columns [][]types.ColumnSelector,
	rows [][]types.Row) (types.Value, error) {

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
			case types.NullValue:

			case types.IntValue:
				add, err := v.Int()
				if err != nil {
					return nil, err
				}
				intSum += add

			case types.FloatValue:
				add, err := v.Float()
				if err != nil {
					return nil, err
				}
				floatSum += add

			default:
				return nil, fmt.Errorf("SUM over %T", val)
			}
		}
		if floatSum != 0 {
			return types.FloatValue(floatSum), nil
		}
		return types.IntValue(intSum), nil

	case FuncAvg:
		var intSum int64
		var floatSum float64
		var count int

		for _, sumRow := range rows {
			val, err := f.Arguments[0].Eval(sumRow, columns, nil)
			if err != nil {
				return nil, err
			}
			switch v := val.(type) {
			case types.NullValue:

			case types.IntValue:
				add, err := v.Int()
				if err != nil {
					return nil, err

				}
				intSum += add
				count++

			case types.FloatValue:
				add, err := v.Float()
				if err != nil {
					return nil, err
				}
				floatSum += add
				count++

			default:
				return nil, fmt.Errorf("AVG over %T", val)
			}
		}
		if count == 0 {
			return types.Null, nil
		}
		if floatSum != 0 {
			return types.FloatValue(floatSum / float64(count)), nil
		}
		return types.IntValue(intSum / int64(count)), nil

	default:
		return nil, fmt.Errorf("unknown function: %v", f.Type)
	}
}

// IsIdempotent implements the Expr.IsIdempotent().
func (f *Function) IsIdempotent() bool {
	// XXX need function registry with attributes. Currently all
	// implemented functions are idempotent.
	return true
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
func (b *Binary) Eval(row []types.Row, columns [][]types.ColumnSelector,
	rows [][]types.Row) (types.Value, error) {

	left, err := b.Left.Eval(row, columns, rows)
	if err != nil {
		return nil, err
	}
	right, err := b.Right.Eval(row, columns, rows)
	if err != nil {
		return nil, err
	}

	// Check null values.
	_, lNull := left.(types.NullValue)
	_, rNull := right.(types.NullValue)
	if lNull || rNull {
		switch b.Type {
		case BinEq:
			return types.BoolValue(lNull && rNull), nil
		case BinNeq:
			return types.BoolValue(lNull != rNull), nil
		default:
			return types.Null, nil
		}
	}

	// Resolve operation type.

	var opType types.Type

	switch left.(type) {
	case types.BoolValue:
		switch right.(type) {
		case types.BoolValue:
			opType = types.Bool
		default:
			return nil,
				fmt.Errorf("invalid types: %s(%T) %s %s(%T)",
					left, left, b.Type, right, right)
		}

	case types.IntValue:
		switch right.(type) {
		case types.IntValue:
			opType = types.Int
		case types.FloatValue:
			opType = types.Float
		default:
			return nil,
				fmt.Errorf("invalid types: %s(%T) %s %s(%T)",
					left, left, b.Type, right, right)
		}

	case types.FloatValue:
		switch right.(type) {
		case types.IntValue, types.FloatValue:
			opType = types.Float
		default:
			return nil,
				fmt.Errorf("invalid types: %s(%T) %s %s(%T)",
					left, left, b.Type, right, right)
		}

	case types.StringValue:
		opType = types.String

	default:
		return nil, fmt.Errorf("binary %s(%T) %s %s(%T) not implemented",
			left, left, b.Type, right, right)
	}

	switch opType {
	case types.Bool:
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
			return types.BoolValue(l == r), nil
		case BinNeq:
			return types.BoolValue(l != r), nil
		default:
			return nil, fmt.Errorf("unknown bool binary expression: %s %s %s",
				left, b.Type, right)
		}

	case types.Int:
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
			return types.BoolValue(l == r), nil
		case BinNeq:
			return types.BoolValue(l != r), nil
		case BinLt:
			return types.BoolValue(l < r), nil
		case BinLe:
			return types.BoolValue(l <= r), nil
		case BinGt:
			return types.BoolValue(l > r), nil
		case BinGe:
			return types.BoolValue(l >= r), nil
		case BinMult:
			return types.IntValue(l * r), nil
		case BinDiv:
			return types.IntValue(l / r), nil
		case BinAdd:
			return types.IntValue(l + r), nil
		case BinSub:
			return types.IntValue(l - r), nil
		default:
			return nil, fmt.Errorf("unknown int binary expression: %s %s %s",
				left, b.Type, right)
		}

	case types.Float:
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
			return types.BoolValue(l == r), nil
		case BinNeq:
			return types.BoolValue(l != r), nil
		case BinLt:
			return types.BoolValue(l < r), nil
		case BinGt:
			return types.BoolValue(l > r), nil
		case BinMult:
			return types.FloatValue(l * r), nil
		case BinDiv:
			return types.FloatValue(l / r), nil
		case BinAdd:
			return types.FloatValue(l + r), nil
		case BinSub:
			return types.FloatValue(l - r), nil
		default:
			return nil, fmt.Errorf("unknown float binary expression: %s %s %s",
				left, b.Type, right)
		}

	case types.String:
		l := left.String()
		r := right.String()
		switch b.Type {
		case BinEq:
			return types.BoolValue(l == r), nil
		case BinNeq:
			return types.BoolValue(l != r), nil
		case BinLt:
			return types.BoolValue(l < r), nil
		case BinGt:
			return types.BoolValue(l > r), nil
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

// IsIdempotent implements the Expr.IsIdempotent().
func (b *Binary) IsIdempotent() bool {
	return b.Left.IsIdempotent() && b.Right.IsIdempotent()
}

func (b *Binary) String() string {
	return fmt.Sprintf("%s %s %s", b.Left, b.Type, b.Right)
}

// And implements logical AND expressions.
type And struct {
	Left  Expr
	Right Expr
}

// Bind implements the Expr.Bind().
func (and *And) Bind(sql *Query) error {
	err := and.Left.Bind(sql)
	if err != nil {
		return err
	}
	return and.Right.Bind(sql)
}

// Eval implements the Expr.Eval().
func (and *And) Eval(row []types.Row, columns [][]types.ColumnSelector,
	rows [][]types.Row) (types.Value, error) {

	left, err := and.Left.Eval(row, columns, rows)
	if err != nil {
		return nil, err
	}
	l, err := left.Bool()
	if err != nil {
		return nil, err
	}
	if !l {
		return types.BoolValue(false), nil
	}

	right, err := and.Right.Eval(row, columns, rows)
	if err != nil {
		return nil, err
	}
	r, err := right.Bool()
	if err != nil {
		return nil, err
	}
	return types.BoolValue(r), nil
}

// IsIdempotent implements the Expr.IsIdempotent().
func (and *And) IsIdempotent() bool {
	return and.Left.IsIdempotent() && and.Right.IsIdempotent()
}

func (and *And) String() string {
	return fmt.Sprintf("%s AND %s", and.Left, and.Right)
}

// Constant implements contant expressions.
type Constant struct {
	Value types.Value
}

// Bind implements the Expr.Bind().
func (c *Constant) Bind(sql *Query) error {
	return nil
}

// Eval implements the Expr.Eval().
func (c *Constant) Eval(row []types.Row, columns [][]types.ColumnSelector,
	rows [][]types.Row) (types.Value, error) {

	return c.Value, nil
}

// IsIdempotent implements the Expr.IsIdempotent().
func (c *Constant) IsIdempotent() bool {
	return true
}

func (c *Constant) String() string {
	return c.Value.String()
}

// Reference implements column reference expressions.
type Reference struct {
	types.Reference
	index   columnIndex
	binding *Binding
	public  bool
}

// NewReference creates a new reference for the argument name.
func NewReference(name string) (*Reference, error) {
	r, err := types.NewReference(name)
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
func (ref *Reference) Eval(row []types.Row, columns [][]types.ColumnSelector,
	rows [][]types.Row) (types.Value, error) {

	col := row[ref.index.source][ref.index.column]
	t := columns[ref.index.source][ref.index.column].Type

	switch t {
	case types.Bool:
		return col.Bool()
	case types.Int:
		return col.Int()
	case types.Float:
		return col.Float()
	default:
		return types.StringValue(col.String()), nil
	}
}

// IsIdempotent implements the Expr.IsIdempotent().
func (ref *Reference) IsIdempotent() bool {
	// Variable references are idempotent, column references are not.
	if ref.binding != nil {
		return true
	}
	return false
}
