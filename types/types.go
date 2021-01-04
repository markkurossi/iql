//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/markkurossi/tabulate"
)

// Type specifies language types.
type Type int

// Types.
const (
	Bool Type = iota
	Int
	Float
	Date
	String
	Table
	Any
)

// Literal values.
const (
	True  = "true"
	False = "false"
)

// Date formats.
const (
	DateTimeLayout      = "2006-01-02 15:04:05.999999999"
	DateTimeZoneLayout  = "2006-01-02 15:04:05.999999999 -07:00"
	DateTimeZoneLayout2 = "2006-01-02T15:04:05.999999999 -07:00"
	DateLayout          = "2006-01-02"
	DateLayoutUS        = "01/02/2006"
)

var dateFormats = []string{
	time.RFC3339Nano,
	DateTimeLayout,
	DateTimeZoneLayout,
	DateTimeZoneLayout2,
	DateLayout,
	DateLayoutUS,
}

// ParseDate parses the datetime literal value.
func ParseDate(val string) (time.Time, error) {
	for _, fmt := range dateFormats {
		t, err := time.Parse(fmt, val)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date value: %s", val)
}

// ParseBoolean parses the boolean literal value.
func ParseBoolean(val string) (bool, bool) {
	switch strings.ToLower(val) {
	case True, "on":
		return true, true
	case False, "off":
		return false, true
	default:
		return false, false
	}
}

var types = map[Type]string{
	Bool:   "boolean",
	Int:    "integer",
	Float:  "real",
	Date:   "datetime",
	String: "varchar",
	Table:  "table",
}

func (t Type) String() string {
	name, ok := types[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{Type %d}", t)
}

// Align returns the type specific column alignment type.
func (t Type) Align() tabulate.Align {
	if t == String {
		return tabulate.ML
	}
	return tabulate.MR
}

// CanAssign tests if the argument value can be assigned into a
// variable this type.
func (t Type) CanAssign(v Value) bool {
	switch v.(type) {
	case BoolValue:
		return t == Bool
	case IntValue:
		return t == Int || t == Float
	case FloatValue:
		return t == Int || t == Float
	case DateValue:
		return t == Date
	case StringValue:
		return t == String
	case TableValue:
		return t == Table
	case NullValue:
		return true
	default:
		return false
	}
}
