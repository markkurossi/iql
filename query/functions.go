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
	FuncSum FunctionType = iota
	FuncAvg
	FuncCount
	FuncComposite
)

var functionTypes = map[FunctionType]string{
	FuncSum:       "SUM",
	FuncAvg:       "AVG",
	FuncCount:     "COUNT",
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
	"SUM": {
		Type:       FuncSum,
		MinArgs:    1,
		MaxArgs:    1,
		Idempotent: true,
	},
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

	switch f.Type {
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

	default:
		return nil, fmt.Errorf("unknown function: %v", f.Type)
	}
}
