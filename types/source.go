//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package types

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/markkurossi/tabulate"
	"github.com/markkurossi/vt100"
)

var (
	_ Column = NullColumn{}
	_ Column = ValueColumn{}
	_ Column = StringColumn("")
	_ Column = StringsColumn([]string{})
)

// Source is an interface that defines data input sources.
type Source interface {
	Columns() []ColumnSelector
	Get() ([]Row, error)
}

// Row defines an input data row.
type Row []Column

// ColumnSelector implements data column selector.
type ColumnSelector struct {
	Name Reference
	As   string
	Type Type
}

// IsPublic reports if the column is public and should be included in
// the result set.
func (col ColumnSelector) IsPublic() bool {
	runes := []rune(col.String())
	return len(runes) > 0 && unicode.IsUpper(runes[0])
}

// ResolveValue resolves the column type based on the argument column
// value. This function must be called once for each value and it will
// resolve the most specific column type that is able to represent all
// values.
func (col *ColumnSelector) ResolveValue(val Value) {
	_, ok := val.(NullValue)
	if ok {
		return
	}

	t := val.Type()
	if t > col.Type {
		col.Type = t
	}
	if col.Type > String {
		col.Type = String
	}
}

// ResolveString resolves the column type based on the argument column
// value. This function must be called once for each value and it will
// resolve the most specific column type that is able to represent all
// values.
func (col *ColumnSelector) ResolveString(val string) {
	// Skip empty values.
	if len(val) == 0 {
		return
	}
	for {
		switch col.Type {
		case Bool:
			if val == True || val == False {
				return
			}
			col.Type = Int

		case Int:
			_, err := strconv.ParseInt(val, 10, 64)
			if err == nil {
				return
			}
			col.Type = Float

		case Float:
			_, err := strconv.ParseFloat(val, 64)
			if err == nil {
				return
			}
			col.Type = String

		case String:
			return
		}
	}
}

func (col ColumnSelector) String() string {
	if len(col.As) > 0 {
		return col.As
	}
	return col.Name.Column
}

// Reference implements a reference to table column.
type Reference struct {
	Source string
	Column string
}

// NewReference creates a new column reference for the argument name.
func NewReference(name string) (Reference, error) {
	// XXX escapes
	parts := strings.Split(name, ".")
	switch len(parts) {
	case 1:
		return Reference{
			Column: parts[0],
		}, nil

	case 2:
		return Reference{
			Source: parts[0],
			Column: parts[1],
		}, nil

	default:
		return Reference{}, fmt.Errorf("invalid column reference '%s'", name)
	}
}

// IsAbsolute tests if the reference is an absolute reference
// i.e. specifying both the data source and column.
func (ref *Reference) IsAbsolute() bool {
	return len(ref.Source) > 0
}

func (ref Reference) String() string {
	// XXX escapes
	if len(ref.Source) > 0 {
		return fmt.Sprintf("%s.%s", ref.Source, ref.Column)
	}
	return ref.Column
}

// Column defines a data column.
type Column interface {
	Bool() (Value, error)
	Int() (Value, error)
	Float() (Value, error)
	String() string
}

// NullColumn implements a null-column.
type NullColumn NullValue

// Bool implements the Column.Bool().
func (n NullColumn) Bool() (Value, error) {
	return Null, nil
}

// Int implements the Column.Int().
func (n NullColumn) Int() (Value, error) {
	return Null, nil
}

// Float implements the Column.Float().
func (n NullColumn) Float() (Value, error) {
	return Null, nil
}

func (n NullColumn) String() string {
	return "NULL"
}

// ValueColumn implements column over Value.
type ValueColumn struct {
	v Value
}

// NewValueColumn returns a new ValueColum containing the argument
// value.
func NewValueColumn(v Value) *ValueColumn {
	return &ValueColumn{
		v: v,
	}
}

// Bool implements the Column.Bool().
func (c ValueColumn) Bool() (Value, error) {
	val, err := c.v.Bool()
	if err != nil {
		return nil, err
	}
	return BoolValue(val), nil
}

// Int implements the Column.Int().
func (c ValueColumn) Int() (Value, error) {
	val, err := c.v.Int()
	if err != nil {
		return nil, err
	}
	return IntValue(val), nil
}

// Float implements the Column.Float().
func (c ValueColumn) Float() (Value, error) {
	val, err := c.v.Float()
	if err != nil {
		return nil, err
	}
	return FloatValue(val), nil
}

func (c ValueColumn) String() string {
	return c.v.String()
}

// StringColumn implements a string column.
type StringColumn string

// Bool implements the Column.Bool().
func (s StringColumn) Bool() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	switch s {
	case True:
		return BoolValue(true), nil
	case False:
		return BoolValue(false), nil
	default:
		return nil, fmt.Errorf("string value '%s' used as bool", s)
	}
}

// Int implements the Column.Int().
func (s StringColumn) Int() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	v, err := strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		return nil, err
	}
	return IntValue(v), nil
}

// Float implements the Column.Float().
func (s StringColumn) Float() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	v, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		return nil, err
	}
	return FloatValue(v), nil
}

func (s StringColumn) String() string {
	return string(s)
}

// StringsColumn implements a string array column.
type StringsColumn []string

// Bool implements the Column.Bool().
func (s StringsColumn) Bool() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	return nil, fmt.Errorf("string array used as bool")
}

// Int implements the Column.Int().
func (s StringsColumn) Int() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	return nil, fmt.Errorf("string array used as int")
}

// Float implements the Column.Float().
func (s StringsColumn) Float() (Value, error) {
	if len(s) == 0 {
		return Null, nil
	}
	return nil, fmt.Errorf("string array used as float")
}

func (s StringsColumn) String() string {
	return fmt.Sprintf("%v", []string(s))
}

// Tabulate creates a tabulation table for the data source.
func Tabulate(source Source, style tabulate.Style) (*tabulate.Tabulate, error) {
	rows, err := source.Get()
	if err != nil {
		return nil, err
	}
	tab := tabulate.New(style)
	tab.Measure = func(column string) int {
		w, _ := vt100.DisplayWidth(column)
		return w
	}
	for _, col := range source.Columns() {
		tab.Header(col.String()).SetAlign(col.Type.Align())
	}
	for _, columns := range rows {
		row := tab.Row()
		for _, col := range columns {
			_, ok := col.(NullColumn)
			if ok {
				row.Column("")
			} else {
				row.Column(col.String())
			}
		}
	}
	return tab, nil
}
