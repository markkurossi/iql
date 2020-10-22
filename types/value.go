//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package types

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	_ Value = BoolValue(false)
	_ Value = IntValue(0)
	_ Value = FloatValue(0.0)
	_ Value = StringValue("")
	_ Value = TableValue{}

	// Null value specifies a non-existing value.
	Null Value = NullValue{}
)

// Value implements expression values.
type Value interface {
	Bool() (bool, error)
	Int() (int64, error)
	Float() (float64, error)
	String() string
}

// Equal tests if the argument values are equal.
func Equal(value1, value2 Value) (bool, error) {
	switch v1 := value1.(type) {
	case BoolValue:
		v2, ok := value2.(BoolValue)
		return ok && v1 == v2, nil

	case IntValue:
		v2, ok := value2.(IntValue)
		return ok && v1 == v2, nil

	case FloatValue:
		v2, ok := value2.(FloatValue)
		return ok && v1 == v2, nil

	case StringValue:
		v2, ok := value2.(StringValue)
		return ok && v1 == v2, nil

	default:
		return false, fmt.Errorf("types.Equal: invalid type: %T", value1)
	}
}

// BoolValue implements boolean values.
type BoolValue bool

// Bool implements the Value.Bool().
func (v BoolValue) Bool() (bool, error) {
	return bool(v), nil
}

// Int implements the Value.Int().
func (v BoolValue) Int() (int64, error) {
	return 0, fmt.Errorf("bool used as int")
}

// Float implements the Value.Float().
func (v BoolValue) Float() (float64, error) {
	return 0, fmt.Errorf("bool used as float")
}

func (v BoolValue) String() string {
	return fmt.Sprintf("%v", bool(v))
}

// IntValue implements integer values.
type IntValue int64

// Bool implements the Value.Bool().
func (v IntValue) Bool() (bool, error) {
	return false, fmt.Errorf("int used as bool")
}

// Int implements the Value.Int().
func (v IntValue) Int() (int64, error) {
	return int64(v), nil
}

// Float implements the Value.Float().
func (v IntValue) Float() (float64, error) {
	return float64(v), nil
}

func (v IntValue) String() string {
	return fmt.Sprintf("%d", v)
}

// FloatValue implements floating point values.
type FloatValue float64

// Bool implements the Value.Bool().
func (v FloatValue) Bool() (bool, error) {
	return false, fmt.Errorf("int used as bool")
}

// Int implements the Value.Int().
func (v FloatValue) Int() (int64, error) {
	return int64(v), nil
}

// Float implements the Value.Float().
func (v FloatValue) Float() (float64, error) {
	return float64(v), nil
}

func (v FloatValue) String() string {
	return fmt.Sprintf("%.2f", float64(v))
}

// StringValue implements string values.
type StringValue string

// Bool implements the Value.Bool().
func (v StringValue) Bool() (bool, error) {
	return false, fmt.Errorf("string used as bool")
}

// Int implements the Value.Int().
func (v StringValue) Int() (int64, error) {
	val, err := strconv.ParseInt(string(v), 10, 64)
	if err != nil {
		panic("StringValue.Int")
	}
	return val, err
}

// Float implements the Value.Float().
func (v StringValue) Float() (float64, error) {
	return strconv.ParseFloat(string(v), 64)
}

func (v StringValue) String() string {
	return string(v)
}

// TableValue implements table values for sources.
type TableValue struct {
	Source Source
}

// Bool implements the Value.Bool().
func (v TableValue) Bool() (bool, error) {
	return false, fmt.Errorf("table used as bool")
}

// Int implements the Value.Int().
func (v TableValue) Int() (int64, error) {
	rows, err := v.Source.Get()
	if err != nil {
		return 0, err
	}
	return int64(len(rows)), nil
}

// Float implements the Value.Float().
func (v TableValue) Float() (float64, error) {
	return 0, fmt.Errorf("table used as float")
}

func (v TableValue) String() string {
	// XXX source names
	return "table"
}

// NullValue implements non-existing value.
type NullValue struct {
}

// Bool implements the Value.Bool().
func (v NullValue) Bool() (bool, error) {
	return false, errors.New("null used as bool")
}

// Int implements the Value.Int().
func (v NullValue) Int() (int64, error) {
	return 0, errors.New("null used as int")
}

// Float implements the Value.Float().
func (v NullValue) Float() (float64, error) {
	return 0, errors.New("null used as float")
}

func (v NullValue) String() string {
	return "null"
}
