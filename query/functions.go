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

// Function implements a function.
type Function struct {
	Type       FunctionType
	MinArgs    int
	MaxArgs    int
	Idempotent bool

	// Name of a composite function. This is empty for built-in
	// functions
	Name string
}

// FunctionType specifies function types and built-in functions.
type FunctionType int

// Function types.
const (
	FuncAvg FunctionType = iota
	FuncCount
	FuncMax
	FuncMin
	FuncNullIf
	FuncSum
	FuncComposite
)

var functionTypes = map[FunctionType]string{
	FuncAvg:       "AVG",
	FuncCount:     "COUNT",
	FuncMax:       "MAX",
	FuncMin:       "MIN",
	FuncNullIf:    "NULLIF,",
	FuncSum:       "SUM",
	FuncComposite: "{composite}",
}

func (t FunctionType) String() string {
	name, ok := functionTypes[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{function %d}", t)
}

var builtIns = map[string]*Function{
	"AVG": {
		Type:       FuncAvg,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
	"COUNT": {
		Type:       FuncCount,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
	"MAX": {
		Type:       FuncMax,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
	"MIN": {
		Type:       FuncMin,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
	"NULLIF": {
		Type:       FuncNullIf,
		MinArgs:    2,
		MaxArgs:    2,
		Idempotent: false,
	},
	"SUM": {
		Type:       FuncSum,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
}

func builtIn(name string) *Function {
	return builtIns[name]
}

func (f *Function) String() string {
	if f.Type == FuncComposite {
		return f.Name
	}
	return f.Type.String()
}

// Eval evaluates the function with the arguments and rows.
func (f *Function) Eval(args []Expr, row []types.Row,
	columns [][]types.ColumnSelector, rows [][]types.Row) (types.Value, error) {

	if len(args) < f.MinArgs {
		return nil, fmt.Errorf("%s: too few arguments: got %d, expected %d",
			f, len(args), f.MinArgs)
	}
	if len(args) > f.MaxArgs {
		return nil, fmt.Errorf("%s: too many arguments: got %d, expected %d",
			f, len(args), f.MaxArgs)
	}

	seen := make(map[types.Type]bool)

	switch f.Type {
	case FuncAvg:
		var intSum int64
		var floatSum float64
		var count int

		for _, sumRow := range rows {
			val, err := args[0].Eval(sumRow, columns, nil)
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

	case FuncCount:
		var count int
		for _, countRow := range rows {
			val, err := args[0].Eval(countRow, columns, nil)
			if err != nil {
				return nil, err
			}
			_, ok := val.(types.NullValue)
			if !ok {
				count++
			}
		}
		return types.IntValue(count), nil

	case FuncMax:
		var intMax int64
		var floatMax float64

		for _, sumRow := range rows {
			val, err := args[0].Eval(sumRow, columns, nil)
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

	case FuncMin:
		var intMin int64
		var floatMin float64

		for _, sumRow := range rows {
			val, err := args[0].Eval(sumRow, columns, nil)
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

	case FuncNullIf:
		val, err := args[0].Eval(row, columns, rows)
		if err != nil {
			return nil, err
		}
		cmp, err := args[1].Eval(row, columns, rows)
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

	case FuncSum:
		var intSum int64
		var floatSum float64

		for _, sumRow := range rows {
			val, err := args[0].Eval(sumRow, columns, nil)
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

	default:
		return nil, fmt.Errorf("unknown function: %v", f.Type)
	}
}
