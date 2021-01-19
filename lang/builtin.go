//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package lang

import (
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/vt100"
)

// Function implements a function.
type Function struct {
	Name         string
	Args         []FunctionArg
	RetType      types.Type
	Ret          Expr
	Impl         FunctionImpl
	MinArgs      int
	MaxArgs      int
	FirstBound   int
	IsIdempotent IsIdempotent
}

func (f *Function) String() string {
	if f.Impl != nil {
		return fmt.Sprintf("builtin %s", f.Name)
	}
	return fmt.Sprintf("%s(%v) %s", f.Name, f.Args, f.RetType)
}

// FunctionArg defines function arguments for user-defined
// functions. Builtin functions verify function parameter types
// dynamically.
type FunctionArg struct {
	Name string
	Type types.Type
}

// FunctionImpl implements the built-in IQL functions.
type FunctionImpl func(args []Expr, row *Row, rows []*Row) (types.Value, error)

// IsIdempotent tests if the function is idempotent when applied to
// its arguments.
type IsIdempotent func(args []Expr) bool

func idempotentTrue(args []Expr) bool {
	return true
}

func idempotentFalse(args []Expr) bool {
	return false
}

func idempotentArgs(args []Expr) bool {
	for _, arg := range args {
		if !arg.IsIdempotent() {
			return false
		}
	}
	return true
}

var builtIns = []Function{
	// Aggregate functions.
	{
		Name:         "AVG",
		Impl:         builtInAvg,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentTrue,
	},
	{
		Name:         "COUNT",
		Impl:         builtInCount,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentTrue,
	},
	{
		Name:         "MAX",
		Impl:         builtInMax,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentTrue,
	},
	{
		Name:         "MIN",
		Impl:         builtInMin,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentTrue,
	},
	{
		Name:         "SUM",
		Impl:         builtInSum,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentTrue,
	},
	{
		Name:         "NULLIF",
		Impl:         builtInNullIf,
		MinArgs:      2,
		MaxArgs:      2,
		IsIdempotent: idempotentArgs,
	},

	// Mathematical function.
	{
		Name:         "FLOOR",
		Impl:         builtInFloor,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},

	// String functions.
	{
		Name:         "CHAR",
		Impl:         builtInChar,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "CHARINDEX",
		Impl:         builtInCharIndex,
		MinArgs:      2,
		MaxArgs:      3,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "CONCAT",
		Impl:         builtInConcat,
		MinArgs:      2,
		MaxArgs:      math.MaxInt32,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "CONCAT_WS",
		Impl:         builtInConcatWS,
		MinArgs:      3,
		MaxArgs:      math.MaxInt32,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "BASE64ENC",
		Impl:         builtInBase64Enc,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "BASE64DEC",
		Impl:         builtInBase64Dec,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "LASTCHARINDEX",
		Impl:         builtInLastCharIndex,
		MinArgs:      2,
		MaxArgs:      2,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "LEFT",
		Impl:         builtInLeft,
		MinArgs:      2,
		MaxArgs:      2,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "LEN",
		Impl:         builtInLen,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "LOWER",
		Impl:         builtInLower,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "LPAD",
		Impl:         builtInLPad,
		MinArgs:      2,
		MaxArgs:      3,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "LTRIM",
		Impl:         builtInLTrim,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "NCHAR",
		Impl:         builtInNChar,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "REPLICATE",
		Impl:         builtInReplicate,
		MinArgs:      2,
		MaxArgs:      2,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "REVERSE",
		Impl:         builtInReverse,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "RIGHT",
		Impl:         builtInRight,
		MinArgs:      2,
		MaxArgs:      2,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "RTRIM",
		Impl:         builtInRTrim,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "SPACE",
		Impl:         builtInSpace,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "STUFF",
		Impl:         builtInStuff,
		MinArgs:      4,
		MaxArgs:      4,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "SUBSTRING",
		Impl:         builtInSubstring,
		MinArgs:      3,
		MaxArgs:      3,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "TRIM",
		Impl:         builtInTrim,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "UNICODE",
		Impl:         builtInUnicode,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "UPPER",
		Impl:         builtInUpper,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},

	// Datetime functions.
	{
		Name:         "DATEDIFF",
		Impl:         builtInDateDiff,
		MinArgs:      3,
		MaxArgs:      3,
		FirstBound:   1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "DAY",
		Impl:         builtInDay,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "GETDATE",
		Impl:         builtInGetDate,
		MinArgs:      0,
		MaxArgs:      0,
		IsIdempotent: idempotentFalse,
	},
	{
		Name:         "MONTH",
		Impl:         builtInMonth,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},
	{
		Name:         "YEAR",
		Impl:         builtInYear,
		MinArgs:      1,
		MaxArgs:      1,
		IsIdempotent: idempotentArgs,
	},

	// Visualization functions.
	{
		Name:         "HBAR",
		Impl:         builtInHBar,
		MinArgs:      3,
		MaxArgs:      4,
		IsIdempotent: idempotentArgs,
	},
}

func builtInAvg(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	seen := make(map[types.Type]bool)

	var intSum int64
	var floatSum float64
	var count int

	for _, sumRow := range rows {
		val, err := args[0].Eval(sumRow, nil)
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
			seen[types.Int] = true
			intSum += add
			count++

		case types.FloatValue:
			add, err := v.Float()
			if err != nil {
				return nil, err
			}
			seen[types.Float] = true
			floatSum += add
			count++

		default:
			return nil, fmt.Errorf("AVG over %T", val)
		}
	}
	if count == 0 || len(seen) != 1 {
		return types.Null, nil
	}
	if seen[types.Float] {
		return types.FloatValue(floatSum / float64(count)), nil
	}
	return types.IntValue(intSum / int64(count)), nil
}

func builtInCount(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	var count int
	for _, countRow := range rows {
		val, err := args[0].Eval(countRow, nil)
		if err != nil {
			return nil, err
		}
		_, ok := val.(types.NullValue)
		if !ok {
			count++
		}
	}
	return types.IntValue(count), nil
}

func builtInMax(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	seen := make(map[types.Type]bool)

	var intMax int64
	var floatMax float64

	for _, sumRow := range rows {
		val, err := args[0].Eval(sumRow, nil)
		if err != nil {
			return nil, err
		}
		switch v := val.(type) {
		case types.NullValue:

		case types.IntValue:
			ival, err := v.Int()
			if err != nil {
				return nil, err
			}
			if !seen[types.Int] || ival > intMax {
				intMax = ival
			}
			seen[types.Int] = true

		case types.FloatValue:
			fval, err := v.Float()
			if err != nil {
				return nil, err
			}
			if !seen[types.Float] || fval > floatMax {
				floatMax = fval
			}
			seen[types.Float] = true

		default:
			return nil, fmt.Errorf("MAX over %T", val)
		}
	}
	if seen[types.Float] && seen[types.Int] {
		var r float64
		if float64(intMax) > floatMax {
			r = float64(intMax)
		} else {
			r = floatMax
		}
		return types.FloatValue(r), nil
	} else if seen[types.Float] {
		return types.FloatValue(floatMax), nil
	}
	return types.IntValue(intMax), nil

}

func builtInMin(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	seen := make(map[types.Type]bool)

	var intMin int64
	var floatMin float64

	for _, sumRow := range rows {
		val, err := args[0].Eval(sumRow, nil)
		if err != nil {
			return nil, err
		}
		switch v := val.(type) {
		case types.NullValue:

		case types.IntValue:
			ival, err := v.Int()
			if err != nil {
				return nil, err
			}
			if !seen[types.Int] || ival < intMin {
				intMin = ival
			}
			seen[types.Int] = true

		case types.FloatValue:
			fval, err := v.Float()
			if err != nil {
				return nil, err
			}
			if !seen[types.Float] || fval < floatMin {
				floatMin = fval
			}
			seen[types.Float] = true

		default:
			return nil, fmt.Errorf("MIN over %T", val)
		}
	}
	if seen[types.Float] && seen[types.Int] {
		var r float64
		if float64(intMin) < floatMin {
			r = float64(intMin)
		} else {
			r = floatMin
		}
		return types.FloatValue(r), nil
	} else if seen[types.Float] {
		return types.FloatValue(floatMin), nil
	}

	return types.IntValue(intMin), nil
}

func builtInSum(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	seen := make(map[types.Type]bool)

	var intSum int64
	var floatSum float64

	for _, sumRow := range rows {
		val, err := args[0].Eval(sumRow, nil)
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
			seen[types.Int] = true
			intSum += add

		case types.FloatValue:
			add, err := v.Float()
			if err != nil {
				return nil, err
			}
			seen[types.Float] = true
			floatSum += add

		default:
			return nil, fmt.Errorf("SUM over %T", val)
		}
	}
	if seen[types.Float] && seen[types.Int] {
		return types.FloatValue(floatSum + float64(intSum)), nil
	} else if seen[types.Float] {
		return types.FloatValue(floatSum), nil
	}
	return types.IntValue(intSum), nil
}

func builtInNullIf(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	cmp, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	ok, err := types.Equal(val, cmp)
	if err != nil {
		return nil, err
	}
	if ok {
		return types.Null, nil
	}
	return val, nil
}

func builtInFloor(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	switch v := val.(type) {
	case types.IntValue:
		return val, nil

	case types.FloatValue:
		return types.FloatValue(math.Floor(float64(v))), nil

	default:
		return types.Null, nil
	}
}

func builtInChar(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	codeVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	code64, err := codeVal.Int()
	if err != nil {
		return nil, err
	}
	if code64 < 0 || code64 > math.MaxInt32 {
		return types.Null, nil
	}
	code := rune(code64)
	return types.StringValue(code), nil
}

func builtInCharIndex(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	strVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	str := strVal.String()
	searchVal, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	search := searchVal.String()

	var idx int
	if len(args) > 2 {
		idxVal, err := args[2].Eval(row, rows)
		if err != nil {
			return nil, err
		}
		idx64, err := idxVal.Int()
		if err != nil {
			return nil, err
		}

		runes := []rune(str)

		if idx64 < 0 {
			idx = 0
		} else if idx64 > math.MaxInt32 {
			idx = math.MaxInt32
		} else {
			idx = int(idx64)
		}
		if idx > len(runes) {
			idx = len(runes)
		}
	}

	return types.IntValue(idx + strings.Index(str[idx:], search) + 1), nil
}

func builtInConcat(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	var sb strings.Builder
	for i := 0; i < len(args); i++ {
		val, err := args[i].Eval(row, rows)
		if err != nil {
			return nil, err
		}
		_, n := val.(types.NullValue)
		if !n {
			sb.WriteString(val.String())
		}
	}

	return types.StringValue(sb.String()), nil
}

func builtInConcatWS(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	separatorStr, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	var separator string
	_, ok := separatorStr.(types.NullValue)
	if !ok {
		separator = separatorStr.String()
	}

	// Collect non-null parts.
	var parts []string
	for i := 1; i < len(args); i++ {
		val, err := args[i].Eval(row, rows)
		if err != nil {
			return nil, err
		}
		_, n := val.(types.NullValue)
		if !n {
			parts = append(parts, val.String())
		}
	}

	// Construct result string.
	var sb strings.Builder
	for idx, part := range parts {
		if idx > 0 && idx < len(parts) {
			sb.WriteString(separator)
		}
		sb.WriteString(part)
	}

	return types.StringValue(sb.String()), nil
}

func builtInBase64Enc(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	strVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	str := base64.StdEncoding.EncodeToString([]byte(strVal.String()))
	return types.StringValue(str), nil
}

func builtInBase64Dec(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	strVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	bytes, err := base64.StdEncoding.DecodeString(strVal.String())
	if err != nil {
		return nil, err
	}
	return types.StringValue(string(bytes)), nil
}

func builtInLastCharIndex(args []Expr, row *Row, rows []*Row) (
	types.Value, error) {

	strVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	str := strVal.String()
	searchVal, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	search := searchVal.String()

	return types.IntValue(strings.LastIndex(str, search) + 1), nil
}

func builtInLeft(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	strVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	idxVal, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	str := strVal.String()
	idx64, err := idxVal.Int()
	if err != nil {
		return nil, err
	}

	runes := []rune(str)

	var idx int
	if idx64 < 0 {
		idx = 0
	} else if idx64 > math.MaxInt32 {
		idx = math.MaxInt32
	} else {
		idx = int(idx64)
	}
	if idx > len(runes) {
		idx = len(runes)
	}
	return types.StringValue(string(runes[:idx])), nil
}

func builtInLen(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	return types.IntValue(len([]rune(val.String()))), nil
}

func builtInLower(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	return types.StringValue(strings.ToLower(val.String())), nil
}

func builtInLPad(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	strVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	lengthVal, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	length64, err := lengthVal.Int()
	if err != nil {
		return nil, err
	}

	var length int
	if length64 < 0 {
		length = 0
	} else if length64 > math.MaxInt32 {
		length = math.MaxInt32
	} else {
		length = int(length64)
	}

	runes := []rune(strVal.String())
	if length <= len(runes) {
		return types.StringValue(string(runes[:length])), nil
	}

	pad := ' '
	if len(args) == 3 {
		padStr, err := args[2].Eval(row, rows)
		if err != nil {
			return nil, err
		}
		padRunes := []rune(padStr.String())
		if len(padRunes) != 1 {
			return nil, fmt.Errorf("LPAD: invalid padding: '%s'", padStr)
		}
		pad = padRunes[0]
	}

	var result []rune
	for i := 0; i < length-len(runes); i++ {
		result = append(result, pad)
	}
	result = append(result, runes...)

	return types.StringValue(string(result)), nil
}

func builtInSubstring(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	strVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	str := strVal.String()
	idxVal, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	idx64, err := idxVal.Int()
	if err != nil {
		return nil, err
	}
	// Index is 1-based.
	idx64--

	runes := []rune(str)

	var idx int
	if idx64 < 0 {
		idx = 0
	} else if idx64 > math.MaxInt32 {
		idx = math.MaxInt32
	} else {
		idx = int(idx64)
	}
	if idx > len(runes) {
		idx = len(runes)
	}

	lengthVal, err := args[2].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	length64, err := lengthVal.Int()
	if err != nil {
		return nil, err
	}
	if length64 < 0 {
		return nil, fmt.Errorf("SUBSTRING: negative length: %d", length64)
	}
	var length int
	if length64 > math.MaxInt32 {
		length = math.MaxInt32
	} else {
		length = int(length64)
	}
	if idx+length > len(runes) {
		length = len(runes) - idx
	}

	return types.StringValue(string(runes[idx : idx+length])), nil
}

func builtInLTrim(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	return types.StringValue(strings.TrimLeftFunc(val.String(),
		func(r rune) bool {
			return unicode.IsSpace(r)
		})), nil
}

func builtInNChar(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	i, err := val.Int()
	if err != nil {
		return nil, err
	}
	if i < 0 || i > math.MaxInt32 {
		return types.Null, nil
	}
	return types.StringValue(string(rune(i))), nil
}

func builtInReplicate(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	strVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	str := strVal.String()

	countVal, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	count, err := countVal.Int()
	if err != nil {
		return nil, err
	}
	if count < 0 {
		return types.Null, nil
	}

	var sb strings.Builder
	var i int64
	for i = 0; i < count; i++ {
		sb.WriteString(str)
	}

	return types.StringValue(sb.String()), nil
}

func builtInReverse(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	runes := []rune(val.String())
	for i := len(runes) - 1; i >= 0; i-- {
		sb.WriteRune(runes[i])
	}

	return types.StringValue(sb.String()), nil
}

func builtInRight(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	strVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	idxVal, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	str := strVal.String()
	idx64, err := idxVal.Int()
	if err != nil {
		return nil, err
	}

	runes := []rune(str)

	var idx int
	if idx64 < 0 {
		idx = 0
	} else if idx64 > math.MaxInt32 {
		idx = math.MaxInt32
	} else {
		idx = int(idx64)
	}
	if idx > len(runes) {
		idx = len(runes)
	}
	return types.StringValue(string(runes[len(runes)-idx:])), nil
}

func builtInRTrim(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	return types.StringValue(strings.TrimRightFunc(val.String(),
		func(r rune) bool {
			return unicode.IsSpace(r)
		})), nil
}

func builtInSpace(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	countVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	count, err := countVal.Int()
	if err != nil {
		return nil, err
	}
	if count < 0 {
		return types.Null, nil
	}

	var sb strings.Builder
	var i int64
	for i = 0; i < count; i++ {
		sb.WriteRune(' ')
	}

	return types.StringValue(sb.String()), nil
}

func builtInStuff(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	str, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	runes := []rune(str.String())

	startVal, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	start64, err := startVal.Int()
	if err != nil {
		return nil, err
	}
	var start int
	if start64 <= 0 {
		return types.Null, nil
	} else if start64 > math.MaxInt32 {
		start = math.MaxInt32
	} else {
		start = int(start64)
	}
	if start > len(runes) {
		return types.Null, nil
	}
	start--

	countVal, err := args[2].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	count64, err := countVal.Int()
	if err != nil {
		return nil, err
	}
	if count64 < 0 {
		return types.Null, nil
	}
	var count int
	if count64 > math.MaxInt32 {
		count = math.MaxInt32
	} else {
		count = int(count64)
	}
	if start+count > len(runes) {
		count = len(runes) - start
	}

	replaceStr, err := args[3].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	var replace string
	_, ok := replaceStr.(types.NullValue)
	if !ok {
		replace = replaceStr.String()
	}

	var sb strings.Builder
	for i := 0; i < start; i++ {
		sb.WriteRune(runes[i])
	}
	sb.WriteString(replace)
	for i := start + count; i < len(runes); i++ {
		sb.WriteRune(runes[i])
	}

	return types.StringValue(sb.String()), nil
}

func builtInTrim(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	return types.StringValue(strings.TrimSpace(val.String())), nil
}

func builtInUnicode(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	str := val.String()
	if len(str) == 0 {
		return types.Null, nil
	}
	return types.IntValue([]rune(str)[0]), nil
}

func builtInUpper(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	val, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	return types.StringValue(strings.ToUpper(val.String())), nil
}

func builtInDateDiff(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	fromVal, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	from, err := fromVal.Date()
	if err != nil {
		return nil, err
	}
	toVal, err := args[2].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	to, err := toVal.Date()
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(args[0].String()) {
	case "year", "yy", "yyyy":
		return types.IntValue(to.Year() - from.Year()), nil

		// XXX quarter, qq, q
		// XXX month, mm, m
		// XXX dayofyear, dy, y

	case "day", "dd", "d":
		d := to.Truncate(time.Hour * 24).Sub(from.Truncate(time.Hour * 24))
		return types.IntValue(d.Hours() / 24), nil

		// XXX week, wk, ww

	case "hour", "hh":
		d := to.Truncate(time.Hour).Sub(from.Truncate(time.Hour))
		return types.IntValue(d.Hours()), nil

	case "minute", "mi", "n":
		d := to.Truncate(time.Minute).Sub(from.Truncate(time.Minute))
		return types.IntValue(d.Minutes()), nil

	case "second", "ss", "s":
		d := to.Truncate(time.Second).Sub(from.Truncate(time.Second))
		return types.IntValue(d / 1000000000), nil

	case "millisecond", "ms":
		d := to.Truncate(time.Millisecond).Sub(from.Truncate(time.Millisecond))
		return types.IntValue(d / 1000000), nil

	case "microsecond", "mcs":
		d := to.Truncate(time.Microsecond).Sub(from.Truncate(time.Microsecond))
		return types.IntValue(d / 1000), nil

	case "nanosecond", "ns":
		return types.IntValue(to.Sub(from)), nil

	default:
		return nil, fmt.Errorf("invalid datepart: %s", args[0])
	}
}

func builtInDay(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	dateVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	date, err := dateVal.Date()
	if err != nil {
		return nil, err
	}
	return types.IntValue(date.Day()), nil
}

func builtInGetDate(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	return types.DateValue(time.Now()), nil
}

func builtInMonth(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	dateVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	date, err := dateVal.Date()
	if err != nil {
		return nil, err
	}
	return types.IntValue(date.Month()), nil
}

func builtInYear(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	dateVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	date, err := dateVal.Date()
	if err != nil {
		return nil, err
	}
	return types.IntValue(date.Year()), nil
}

func builtInHBar(args []Expr, row *Row, rows []*Row) (types.Value, error) {
	valVal, err := args[0].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	val, err := valVal.Float()
	if err != nil {
		return nil, err
	}
	maxVal, err := args[1].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	max, err := maxVal.Float()
	if err != nil {
		return nil, err
	}
	widthVal, err := args[2].Eval(row, rows)
	if err != nil {
		return nil, err
	}
	width64, err := widthVal.Int()
	if err != nil {
		return nil, err
	}
	var width int
	if width64 > math.MaxInt32 {
		width = math.MaxInt32
	} else {
		width = int(width64)
	}

	pad := ' '
	if len(args) == 4 {
		padVal, err := args[3].Eval(row, rows)
		if err != nil {
			return nil, err
		}
		switch pv := padVal.(type) {
		case types.StringValue:
			runes := []rune(pv)
			if len(runes) != 1 {
				return nil, fmt.Errorf("HBAR: invalid pad string: %s", pv)
			}
			pad = runes[0]

		case types.IntValue:
			pad64, err := padVal.Int()
			if err != nil {
				return nil, err
			}
			if pad64 < 0 || pad64 > math.MaxInt32 {
				return nil, fmt.Errorf("HBAR: invalid pad character: %d", pad64)
			}
			pad = rune(pad64)

		default:
			return nil, fmt.Errorf("HBAR: invalid pad character: %s", pv)
		}
	}

	return types.StringValue(vt100.HBlock(width, val/max, pad)), nil
}

var builtInsByName map[string]*Function

func init() {
	builtInsByName = make(map[string]*Function)
	for idx, bi := range builtIns {
		builtInsByName[bi.Name] = &builtIns[idx]
	}
}

func builtIn(name string) *Function {
	return builtInsByName[name]
}

func createFunction(f *Function) error {
	_, ok := builtInsByName[f.Name]
	if ok {
		return fmt.Errorf("function already defined: %s", f.Name)
	}
	builtInsByName[f.Name] = f
	return nil
}

func dropFunction(name string, ifExists bool) error {
	f, ok := builtInsByName[name]
	if !ok {
		if ifExists {
			return nil
		}
		return fmt.Errorf("unknown function: %s", name)
	}
	if f.Impl != nil {
		return fmt.Errorf("can't drop builtin function: %s", name)
	}
	delete(builtInsByName, name)
	return nil
}
