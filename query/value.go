//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"
	"strconv"
)

var (
	_ Value = BoolValue(false)
	_ Value = IntValue(0)
	_ Value = StringValue("")
)

// Value implements expression values.
type Value interface {
	Bool() (bool, error)
	Int() (int, error)
	String() string
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

// String implements the Value.String().
func (v BoolValue) String() string {
	return fmt.Sprintf("%v", bool(v))
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

// String implements the Value.String().
func (v IntValue) String() string {
	return fmt.Sprintf("%d", v)
}

// StringValue implements string values.
type StringValue string

// Bool implements the Value.Bool().
func (v StringValue) Bool() (bool, error) {
	return false, fmt.Errorf("string used as bool")
}

// Int implements the Value.Int().
func (v StringValue) Int() (int, error) {
	return strconv.Atoi(string(v))
}

// String implements the Value.String().
func (v StringValue) String() string {
	return string(v)
}
