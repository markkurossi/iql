//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"fmt"
	"strconv"
)

var (
	_ Value = BoolValue(false)
	_ Value = IntValue(0)
	_ Value = FloatValue(0.0)
	_ Value = StringValue("")
)

// Value implements expression values.
type Value interface {
	Bool() (bool, error)
	Int() (int64, error)
	Float() (float64, error)
	String() string
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
	return fmt.Sprintf("%v", float64(v))
}

// StringValue implements string values.
type StringValue string

// Bool implements the Value.Bool().
func (v StringValue) Bool() (bool, error) {
	return false, fmt.Errorf("string used as bool")
}

// Int implements the Value.Int().
func (v StringValue) Int() (int64, error) {
	return strconv.ParseInt(string(v), 10, 64)
}

// Float implements the Value.Float().
func (v StringValue) Float() (float64, error) {
	return strconv.ParseFloat(string(v), 64)
}

func (v StringValue) String() string {
	return string(v)
}
