//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"
	"os"
	"sort"

	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

var (
	_ types.Source = &Query{}
)

// Query implements an IQL query. It also implements data.Source so
// that the query can be used as a nested data source for other
// queries.
type Query struct {
	Select        []ColumnSelector
	From          []SourceSelector
	Into          *Binding
	Where         Expr
	GroupBy       []Expr
	OrderBy       []Order
	Global        *Scope
	fromColumns   map[string]ColumnIndex
	evaluated     bool
	resultColumns []types.ColumnSelector
	result        []types.Row
}

// Order specifies column sorting order.
type Order struct {
	Expr Expr
	Desc bool
}

// NewQuery creates a new query object.
func NewQuery(global *Scope) *Query {
	return &Query{
		Global:      global,
		fromColumns: make(map[string]ColumnIndex),
	}
}

// SysTermOut describes if terminal output is enabled.
func (iql *Query) SysTermOut() bool {
	b := iql.Global.Get(sysTermOut)
	if b == nil {
		panic("system variable TERMOUT not set")
	}
	v, err := b.Value.Bool()
	if err != nil {
		panic(fmt.Sprintf("invalid system variable value: %s", err))
	}
	return v
}

// ColumnSelector defines selected query output columns.
type ColumnSelector struct {
	Expr Expr
	As   string
	Type types.Type
}

// IsPublic reports if the column is public and should be included in
// the result set.
func (col ColumnSelector) IsPublic() bool {
	if len(col.As) == 0 {
		return true
	}
	runes := []rune(col.As)
	return len(runes) > 0 && runes[0] != ','
}

func (col ColumnSelector) String() string {
	if len(col.As) > 0 {
		return fmt.Sprintf("%s AS %s TYPE %s", col.Expr, col.As, col.Type)
	}
	return fmt.Sprintf("%s TYPE %s", col.Expr, col.Type)
}

// SourceSelector defines an input source with an optional name alias.
type SourceSelector struct {
	Source types.Source
	As     string
}

// Columns implements the Source.Columns().
func (iql *Query) Columns() []types.ColumnSelector {
	return iql.resultColumns
}

// Get implements the Source.Get().
func (iql *Query) Get() ([]types.Row, error) {
	if iql.evaluated {
		return iql.result, nil
	}

	// Eval all sources.
	for sourceIdx, from := range iql.From {
		_, err := from.Source.Get()
		if err != nil {
			return nil, err
		}
		if false {
			fmt.Printf("Source %d", sourceIdx)
			if len(from.As) > 0 {
				fmt.Printf("\tAS %s", from.As)
			}
			fmt.Println()
			tab := tabulate.New(tabulate.Unicode)
			tab.Header("Index")
			tab.Header("data.Name")
			tab.Header("data.As")
			tab.Header("Type")

			for idx, col := range from.Source.Columns() {
				row := tab.Row()
				row.Column(fmt.Sprintf("%d", idx))
				row.Column(col.Name.String())
				row.Column(col.As)
				row.Column(col.Type.String())
			}
			tab.Print(os.Stdout)
		}

		// Collect column names.
		for columnIdx, col := range from.Source.Columns() {
			var columnName string
			if len(col.As) > 0 {
				columnName = col.As
			} else {
				columnName = col.Name.Column
			}

			var key string
			if len(from.As) > 0 {
				key = fmt.Sprintf("%s.%s", from.As, columnName)
			} else {
				key = columnName
			}
			iql.fromColumns[key] = ColumnIndex{
				Source: sourceIdx,
				Column: columnIdx,
				Type:   col.Type,
			}
		}
	}

	if len(iql.Select) == 0 {
		// SELECT *, populate iql.Select from source columns.
		for _, f := range iql.From {
			columns := f.Source.Columns()
			for _, col := range columns {
				ref := col.Name
				if len(f.As) != 0 {
					ref.Source = f.As
				}

				iql.Select = append(iql.Select, ColumnSelector{
					Expr: &Reference{
						Reference: ref,
					},
				})
			}
		}
	}

	// Create column info.
	for _, col := range iql.Select {
		if !col.IsPublic() {
			continue
		}
		// Promote expressions to aliases unless explicit aliases are
		// defined.
		var as string
		if len(col.As) > 0 {
			as = col.As
		} else {
			as = col.Expr.String()
		}
		iql.resultColumns = append(iql.resultColumns, types.ColumnSelector{
			Name: types.Reference{
				Column: col.Expr.String(),
			},
			As: as,
		})
	}

	// Bind SELECT expressions.
	var idempotent = true
	for _, sel := range iql.Select {
		if err := sel.Expr.Bind(iql); err != nil {
			return nil, err
		}
		if !sel.Expr.IsIdempotent() {
			idempotent = false
		}
	}
	// Bind WHERE expressions.
	if iql.Where != nil {
		if err := iql.Where.Bind(iql); err != nil {
			return nil, err
		}
	}
	// Bind GROUP BY expressions.
	for _, group := range iql.GroupBy {
		if err := group.Bind(iql); err != nil {
			return nil, err
		}
	}
	// Bind ORDER BY expressions.
	for _, order := range iql.OrderBy {
		if err := order.Expr.Bind(iql); err != nil {
			return nil, err
		}
	}

	var matches []*Row
	err := iql.eval(0, nil, &matches)
	if err != nil {
		return nil, err
	}

	// Group by.
	grouping := NewGrouping()
	for _, match := range matches {
		var key []types.Value
		for _, group := range iql.GroupBy {
			val, err := group.Eval(match, nil)
			if err != nil {
				return nil, err
			}
			key = append(key, val)
		}
		grouping.Add(key, match)
	}

	// Select result columns.
	matches = nil
	format := Format(iql.Global)
	for _, group := range grouping.Get() {
		for _, match := range group {
			var row types.Row
			var i int
			for _, sel := range iql.Select {
				if !sel.IsPublic() {
					continue
				}
				val, err := sel.Expr.Eval(match, group)
				if err != nil {
					return nil, err
				}
				if val == types.Null {
					row = append(row, types.NullColumn{})
				} else {
					if format != nil {
						val = types.NewFormattedValue(val, format)
					}
					row = append(row, types.NewValueColumn(val))
					iql.resultColumns[i].ResolveValue(val)
				}
				i++
			}
			matches = append(matches, &Row{
				Data:  []types.Row{row},
				Order: match.Order,
			})
			// Idempotent and GROUP BY return one result per group.
			if idempotent || len(iql.GroupBy) > 0 {
				break
			}
		}
	}

	// Order results.
	var sortErr error
	sort.Slice(matches, func(i, j int) bool {
		o1 := matches[i].Order
		o2 := matches[j].Order
		l := len(o1)
		if len(o2) < l {
			l = len(o2)
		}
		for idx := 0; idx < l; idx++ {
			var desc bool
			if idx < len(iql.OrderBy) {
				desc = iql.OrderBy[idx].Desc
			}
			cmp, err := types.Compare(o1[idx], o2[idx])
			if err != nil {
				sortErr = err
				return true
			}
			if cmp == 0 {
				continue
			}
			if cmp < 0 {
				return !desc
			}
			return desc
		}
		return len(o1) < len(o2)
	})
	if sortErr != nil {
		return nil, sortErr
	}

	for _, match := range matches {
		iql.result = append(iql.result, match.Data[0])
	}

	iql.evaluated = true

	return iql.result, nil
}

func (iql *Query) eval(idx int, data []types.Row, result *[]*Row) error {

	if idx >= len(iql.From) {
		match := true
		row := &Row{
			Data: data,
		}
		if iql.Where != nil {
			val, err := iql.Where.Eval(row, nil)
			if err != nil {
				return err
			}
			match, err = val.Bool()
			if err != nil {
				return err
			}
		}
		if match {
			// ORDER BY
			for _, order := range iql.OrderBy {
				v, err := order.Expr.Eval(row, nil)
				if err != nil {
					return err
				}
				row.Order = append(row.Order, v)
			}
			row.Order = append(row.Order, types.IntValue(len(*result)))
			*result = append(*result, row)
		}
		return nil
	}

	rows, err := iql.From[idx].Source.Get()
	if err != nil {
		return err
	}

	for _, row := range rows {
		err := iql.eval(idx+1, append(data, row), result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (iql *Query) resolveName(name types.Reference) (*Reference, error) {

	if name.IsAbsolute() {
		index, ok := iql.fromColumns[name.String()]
		if !ok {
			return nil, fmt.Errorf("undefined column '%s'", name)
		}
		return &Reference{
			Reference: name,
			index:     index,
		}, nil
	}

	var match *Reference

	for _, from := range iql.From {
		key := types.Reference{
			Source: from.As,
			Column: name.Column,
		}
		index, ok := iql.fromColumns[key.String()]
		if ok {
			if match != nil {
				return nil, fmt.Errorf("ambiguous column name '%s'", name)
			}
			match = &Reference{
				Reference: key,
				index:     index,
			}
		}
	}
	if match != nil {
		return match, nil
	}

	// Check variables.
	b := iql.Global.Get(name.Column)
	if b != nil {
		return &Reference{
			Reference: types.Reference{
				Column: name.Column,
			},
			binding: b,
		}, nil
	}

	return nil, fmt.Errorf("undefined identifier '%s'", name)
}
