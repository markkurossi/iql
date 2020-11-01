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
	"unicode"

	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

var (
	_ types.Source = &Query{}
)

// Query implements an SQL query. It also implements data.Source so
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
	return len(runes) > 0 && unicode.IsUpper(runes[0])
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
func (sql *Query) Columns() []types.ColumnSelector {
	return sql.resultColumns
}

// Get implements the Source.Get().
func (sql *Query) Get() ([]types.Row, error) {
	if sql.evaluated {
		return sql.result, nil
	}

	// Create column info.
	for _, col := range sql.Select {
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
		sql.resultColumns = append(sql.resultColumns, types.ColumnSelector{
			Name: types.Reference{
				Column: col.Expr.String(),
			},
			As: as,
		})
	}

	// Eval all sources.
	for sourceIdx, from := range sql.From {
		_, err := from.Source.Get()
		if err != nil {
			return nil, err
		}
		if false {
			fmt.Printf("%d\t%s\n", sourceIdx, from.As)
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
			sql.fromColumns[key] = ColumnIndex{
				Source: sourceIdx,
				Column: columnIdx,
				Type:   col.Type,
			}
		}
	}

	// Bind SELECT expressions.
	var idempotent = true
	for _, sel := range sql.Select {
		if err := sel.Expr.Bind(sql); err != nil {
			return nil, err
		}
		if !sel.Expr.IsIdempotent() {
			idempotent = false
		}
	}
	// Bind WHERE expressions.
	if sql.Where != nil {
		if err := sql.Where.Bind(sql); err != nil {
			return nil, err
		}
	}
	// Bind GROUP BY expressions.
	for _, group := range sql.GroupBy {
		if err := group.Bind(sql); err != nil {
			return nil, err
		}
	}
	// Bind ORDER BY expressions.
	for _, order := range sql.OrderBy {
		if err := order.Expr.Bind(sql); err != nil {
			return nil, err
		}
	}

	var matches []*Row
	err := sql.eval(0, nil, &matches)
	if err != nil {
		return nil, err
	}

	// Group by.
	grouping := NewGrouping()
	for _, match := range matches {
		var key []types.Value
		for _, group := range sql.GroupBy {
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
	for _, group := range grouping.Get() {
		for _, match := range group {
			var row types.Row
			var i int
			for _, sel := range sql.Select {
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
					str := val.String()
					row = append(row, types.StringColumn(str))
					sql.resultColumns[i].ResolveType(str)
				}
				i++
			}
			matches = append(matches, &Row{
				Data:  []types.Row{row},
				Order: match.Order,
			})
			// Idempotent and GROUP BY return one result per group.
			if idempotent || len(sql.GroupBy) > 0 {
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
			if idx < len(sql.OrderBy) {
				desc = sql.OrderBy[idx].Desc
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
		sql.result = append(sql.result, match.Data[0])
	}

	sql.evaluated = true

	return sql.result, nil
}

func (sql *Query) eval(idx int, data []types.Row, result *[]*Row) error {

	if idx >= len(sql.From) {
		match := true
		row := &Row{
			Data: data,
		}
		if sql.Where != nil {
			val, err := sql.Where.Eval(row, nil)
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
			for _, order := range sql.OrderBy {
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

	rows, err := sql.From[idx].Source.Get()
	if err != nil {
		return err
	}

	for _, row := range rows {
		err := sql.eval(idx+1, append(data, row), result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sql *Query) resolveName(name types.Reference, public bool) (
	*Reference, error) {

	if name.IsAbsolute() {
		index, ok := sql.fromColumns[name.String()]
		if !ok {
			return nil, fmt.Errorf("undefined column '%s'", name)
		}
		return &Reference{
			Reference: name,
			index:     index,
			public:    public,
		}, nil
	}

	var match *Reference

	for _, from := range sql.From {
		key := types.Reference{
			Source: from.As,
			Column: name.Column,
		}
		index, ok := sql.fromColumns[key.String()]
		if ok {
			if match != nil {
				return nil, fmt.Errorf("ambiguous column name '%s'", name)
			}
			match = &Reference{
				Reference: key,
				index:     index,
				public:    public,
			}
		}
	}
	if match != nil {
		return match, nil
	}

	// Check variables.
	b := sql.Global.Get(name.Column)
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
