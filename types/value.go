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
	"strings"
	"time"
)

var (
	_ Value = BoolValue(false)
	_ Value = IntValue(0)
	_ Value = FloatValue(0.0)
	_ Value = DateValue(time.Unix(0, 0))
	_ Value = StringValue("")
	_ Value = TableValue{}
	_ Value = &FormattedValue{}

	// Null value specifies a non-existing value.
	Null Value = NullValue{}
)

const (
	defaultFloatFormat = "%g"
)

// Value implements expression values.
type Value interface {
	Type() Type
	Date() (time.Time, error)
	Bool() (bool, error)
	Int() (int64, error)
	Float() (float64, error)
	String() string
}

// Equal tests if the argument values are equal.
func Equal(value1, value2 Value) (bool, error) {
	switch v1 := value1.(type) {
	case BoolValue:
		v2, err := value2.Bool()
		if err != nil {
			return false, nil
		}
		return v1 == BoolValue(v2), nil

	case IntValue:
		v2, err := value2.Int()
		if err != nil {
			return false, nil
		}
		return v1 == IntValue(v2), nil

	case FloatValue:
		v2, err := value2.Float()
		if err != nil {
			return false, nil
		}
		return v1 == FloatValue(v2), nil

	case DateValue:
		v2, err := value2.Date()
		if err != nil {
			return false, nil
		}
		return v1.Equal(DateValue(v2)), nil

	case StringValue:
		return v1 == StringValue(value2.String()), nil

	default:
		return false, fmt.Errorf("types.Equal: invalid type: %T", value1)
	}
}

// Compare compares two values. It returns -1, 0, 1 if the value 1 is
// smaller, equal, or greater than the value 2 respectively.
func Compare(value1, value2 Value) (int, error) {
	_, n1 := value1.(NullValue)
	_, n2 := value2.(NullValue)
	if n1 && n2 {
		return 0, nil
	} else if n1 {
		return -1, nil
	} else if n2 {
		return 1, nil
	}

	switch v1 := value1.(type) {
	case BoolValue:
		v2, ok := value2.(BoolValue)
		if !ok {
			return -1, nil
		}
		if v1 == v2 {
			return 0, nil
		}
		if !v1 {
			return -1, nil
		}
		return 1, nil

	case IntValue:
		v2, ok := value2.(IntValue)
		if !ok {
			return -1, nil
		}
		if v1 < v2 {
			return -1, nil
		}
		if v1 > v2 {
			return 1, nil
		}
		return 0, nil

	case FloatValue:
		v2, ok := value2.(FloatValue)
		if !ok {
			return -1, nil
		}
		if v1 < v2 {
			return -1, nil
		}
		if v1 > v2 {
			return 1, nil
		}
		return 0, nil

	case DateValue:
		v2, ok := value2.(DateValue)
		if !ok {
			return -1, nil
		}
		if v1.Equal(v2) {
			return 0, nil
		}
		if v1.Before(v2) {
			return -1, nil
		}
		return 1, nil

	case StringValue:
		v2, ok := value2.(StringValue)
		if !ok {
			return -1, nil
		}
		return strings.Compare(v1.String(), v2.String()), nil

	default:
		return -1, fmt.Errorf("types.Compare: invalid type: %T", value1)
	}
}

// BoolValue implements boolean values.
type BoolValue bool

// Type implements the Value.Type().
func (v BoolValue) Type() Type {
	return Bool
}

// Date implements the Value.Date().
func (v BoolValue) Date() (time.Time, error) {
	return time.Time{}, fmt.Errorf("bool used as date")
}

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

// Type implements the Value.Type().
func (v IntValue) Type() Type {
	return Int
}

// Date implements the Value.Date().
func (v IntValue) Date() (time.Time, error) {
	return time.Unix(0, int64(v)), nil
}

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

// Type implements the Value.Type().
func (v FloatValue) Type() Type {
	return Float
}

// Date implements the Value.Date().
func (v FloatValue) Date() (time.Time, error) {
	return time.Time{}, fmt.Errorf("float used as date")
}

// Bool implements the Value.Bool().
func (v FloatValue) Bool() (bool, error) {
	return false, fmt.Errorf("float used as bool")
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
	return fmt.Sprintf(defaultFloatFormat, float64(v))
}

// DateValue implements datetime values.
type DateValue time.Time

// Equal tests if the values are equal.
func (v DateValue) Equal(o DateValue) bool {
	return time.Time(v).Equal(time.Time(o))
}

// Before tests if the value v is before the argument value o.
func (v DateValue) Before(o DateValue) bool {
	return time.Time(v).Before(time.Time(o))
}

// Type implements the Value.Type().
func (v DateValue) Type() Type {
	return Date
}

// Date implements the Value.Date().
func (v DateValue) Date() (time.Time, error) {
	return time.Time(v), nil
}

// Bool implements the Value.Bool().
func (v DateValue) Bool() (bool, error) {
	return false, fmt.Errorf("datetime used as bool")
}

// Int implements the Value.Int().
func (v DateValue) Int() (int64, error) {
	return time.Time(v).UnixNano(), nil
}

// Float implements the Value.Float().
func (v DateValue) Float() (float64, error) {
	return 0, fmt.Errorf("datetime used as float")
}

func (v DateValue) String() string {
	return time.Time(v).Format(DateTimeLayout)
}

// StringValue implements string values.
type StringValue string

// Type implements the Value.Type().
func (v StringValue) Type() Type {
	return String
}

// Date implements the Value.Date().
func (v StringValue) Date() (time.Time, error) {
	return ParseDate(string(v))
}

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

// Type implements the Value.Type().
func (v TableValue) Type() Type {
	return Table
}

// Date implements the Value.Date().
func (v TableValue) Date() (time.Time, error) {
	return time.Time{}, fmt.Errorf("table used as date")
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

// Type implements the Value.Type().
func (v NullValue) Type() Type {
	return Any
}

// Date implements the Value.Date().
func (v NullValue) Date() (time.Time, error) {
	return time.Time{}, fmt.Errorf("null used as date")
}

// Bool implements the Value.Bool().
func (v NullValue) Bool() (bool, error) {
	return false, nil
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

// Format implements value formatting options.
type Format struct {
	Float string
}

// FormattedValue implements value by wrapping another value type with
// formatting options.
type FormattedValue struct {
	value  Value
	format *Format
}

// NewFormattedValue creates a new formatted value from the argument
// value and format.
func NewFormattedValue(v Value, f *Format) *FormattedValue {
	return &FormattedValue{
		value:  v,
		format: f,
	}
}

// Type implements the Value.Type().
func (v *FormattedValue) Type() Type {
	return v.value.Type()
}

// Date implements the Value.Date().
func (v *FormattedValue) Date() (time.Time, error) {
	return v.value.Date()
}

// Bool implements the Value.Bool().
func (v *FormattedValue) Bool() (bool, error) {
	return v.value.Bool()
}

// Int implements the Value.Int().
func (v *FormattedValue) Int() (int64, error) {
	return v.value.Int()
}

// Float implements the Value.Float().
func (v *FormattedValue) Float() (float64, error) {
	return v.value.Float()
}

func (v *FormattedValue) String() string {
	switch val := v.value.(type) {
	case FloatValue:
		format := v.format.Float
		if len(format) == 0 {
			format = defaultFloatFormat
		}
		return fmt.Sprintf(format, float64(val))
	default:
		return v.value.String()
	}
}
