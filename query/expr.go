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
	_ Expr = &Call{}
	_ Expr = &Binary{}
	_ Expr = &And{}
	_ Expr = &Constant{}
	_ Expr = &Reference{}
	_ Expr = &Cast{}
	_ Expr = &Case{}
)

// Row implements a row that is evaluated against the query.
type Row struct {
	Data  []types.Row
	Order []types.Value
}

func (r *Row) String() string {
	return fmt.Sprintf("Row %v %v", r.Data, r.Order)
}

// Expr implements expressions.
type Expr interface {
	Bind(sql *Query) error
	Eval(row *Row, rows []*Row) (types.Value, error)
	IsIdempotent() bool
	String() string
}

// Call implements function call expressions.
type Call struct {
	Name      string
	Arguments []Expr
	Function  *Function
}

// Bind implements the Expr.Bind().
func (call *Call) Bind(sql *Query) error {
	for _, arg := range call.Arguments {
		err := arg.Bind(sql)
		if err != nil {
			return err
		}
	}

	// Resolve function.
	call.Function = builtIn(call.Name)
	if call.Function == nil {
		return fmt.Errorf("undefined function: %s", call.Name)
	}

	return nil
}

// Eval implements the Expr.Eval().
func (call *Call) Eval(row *Row, rows []*Row) (types.Value, error) {

	if len(call.Arguments) < call.Function.MinArgs {
		return nil, fmt.Errorf("%s: too few arguments: got %d, expected %d",
			call.Name, len(call.Arguments), call.Function.MinArgs)
	}
	if len(call.Arguments) > call.Function.MaxArgs {
		return nil, fmt.Errorf("%s: too many arguments: got %d, expected %d",
			call.Name, len(call.Arguments), call.Function.MaxArgs)
	}

	return call.Function.Impl(call.Arguments, row, rows)
}

// IsIdempotent implements the Expr.IsIdempotent().
func (call *Call) IsIdempotent() bool {
	return call.Function.Idempotent
}

func (call *Call) String() string {
	return fmt.Sprintf("%s(%q)", call.Function, call.Arguments)
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
func (b *Binary) Eval(row *Row, rows []*Row) (types.Value, error) {

	left, err := b.Left.Eval(row, rows)
	if err != nil {
		return nil, err
	}
	right, err := b.Right.Eval(row, rows)
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
		case BinAdd:
			return types.StringValue(l + r), nil
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
func (and *And) Eval(row *Row, rows []*Row) (types.Value, error) {

	left, err := and.Left.Eval(row, rows)
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

	right, err := and.Right.Eval(row, rows)
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
func (c *Constant) Eval(row *Row, rows []*Row) (types.Value, error) {

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
	index   ColumnIndex
	binding *Binding
	public  bool
	bound   bool
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

// ColumnIndex identifies a query Row index.
type ColumnIndex struct {
	Source int
	Column int
	Type   types.Type
}

func (idx ColumnIndex) String() string {
	return fmt.Sprintf("%d.%d", idx.Source, idx.Column)
}

// Bind implements the Expr.Bind().
func (ref *Reference) Bind(sql *Query) error {
	r, err := sql.resolveName(ref.Reference, false)
	if err != nil {
		return err
	}
	ref.index = r.index
	ref.binding = r.binding
	ref.bound = true

	return nil
}

// Eval implements the Expr.Eval().
func (ref *Reference) Eval(row *Row, rows []*Row) (types.Value, error) {

	if !ref.bound {
		return nil, fmt.Errorf("unbound identifier '%s'", ref.Reference)
	}
	if ref.binding != nil {
		return ref.binding.Value, nil
	}

	col := row.Data[ref.index.Source][ref.index.Column]

	switch ref.index.Type {
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

// Cast implements type cast expressions.
type Cast struct {
	Expr Expr
	Type types.Type
}

// Bind implements the Expr.Bind().
func (c *Cast) Bind(sql *Query) error {
	return c.Expr.Bind(sql)
}

// Eval implements the Expr.Eval().
func (c *Cast) Eval(row *Row, rows []*Row) (types.Value, error) {

	val, err := c.Expr.Eval(row, rows)
	if err != nil {
		return nil, err
	}
	switch c.Type {
	case types.Bool:
		v, err := val.Bool()
		if err != nil {
			return nil, err
		}
		return types.BoolValue(v), nil

	case types.Int:
		v, err := val.Int()
		if err != nil {
			return nil, err
		}
		return types.IntValue(v), nil

	case types.Float:
		v, err := val.Float()
		if err != nil {
			return nil, err
		}
		return types.FloatValue(v), nil

	case types.String:
		return types.StringValue(val.String()), nil

	default:
		return nil, fmt.Errorf("CAST(%s AS %s) not supported", c.Expr, c.Type)
	}
}

// IsIdempotent implements the Expr.IsIdempotent().
func (c *Cast) IsIdempotent() bool {
	return c.Expr.IsIdempotent()
}

func (c *Cast) String() string {
	return fmt.Sprintf("CAST(%s AS %s)", c.Expr, c.Type)
}

// Case implements case expressions.
type Case struct {
	Input    Expr
	Branches []Branch
	Else     Expr
}

// Branch implements a case branch.
type Branch struct {
	When Expr
	Then Expr
}

// Bind implements the Expr.Bind().
func (c *Case) Bind(sql *Query) error {
	if c.Input != nil {
		if err := c.Input.Bind(sql); err != nil {
			return err
		}
	}
	for _, b := range c.Branches {
		if err := b.When.Bind(sql); err != nil {
			return err
		}
		if err := b.Then.Bind(sql); err != nil {
			return err
		}
	}
	if c.Else != nil {
		return c.Else.Bind(sql)
	}
	return nil
}

// Eval implements the Expr.Eval().
func (c *Case) Eval(row *Row, rows []*Row) (types.Value, error) {

	var input types.Value
	var err error

	if c.Input != nil {
		input, err = c.Input.Eval(row, rows)
		if err != nil {
			return nil, err
		}
	}

	for _, b := range c.Branches {
		val, err := b.When.Eval(row, rows)
		if err != nil {
			return nil, err
		}
		var bval bool

		if input != nil {
			bval, err = types.Equal(input, val)
		} else {
			bval, err = val.Bool()
		}
		if err != nil {
			return nil, err
		}

		if bval {
			return b.Then.Eval(row, rows)
		}
	}
	if c.Else != nil {
		return c.Else.Eval(row, rows)
	}
	return types.Null, nil
}

// IsIdempotent implements the Expr.IsIdempotent().
func (c *Case) IsIdempotent() bool {
	if c.Input != nil && !c.Input.IsIdempotent() {
		return false
	}
	for _, b := range c.Branches {
		if !b.When.IsIdempotent() {
			return false
		}
		if !b.Then.IsIdempotent() {
			return false
		}
	}
	if c.Else != nil && !c.Else.IsIdempotent() {
		return false
	}
	return true
}

func (c *Case) String() string {
	return fmt.Sprintf("CASE %s %v ELSE %s END", c.Input, c.Branches, c.Else)
}
