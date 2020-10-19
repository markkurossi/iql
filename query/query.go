//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"
	"os"
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
	Global        *Scope
	selectColumns []types.ColumnSelector
	fromColumns   map[string]columnIndex
	evaluated     bool
	result        []types.Row
}

// NewQuery creates a new query object.
func NewQuery(global *Scope) *Query {
	return &Query{
		Global:      global,
		fromColumns: make(map[string]columnIndex),
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
	return sql.selectColumns
}

// Get implements the Source.Get().
func (sql *Query) Get() ([]types.Row, error) {
	if sql.evaluated {
		return sql.result, nil
	}

	// Create column info.
	for _, col := range sql.Select {
		// Promote expressions to aliases unless explicit aliases are
		// defined.
		var as string
		if len(col.As) > 0 {
			as = col.As
		} else {
			as = col.Expr.String()
		}
		sql.selectColumns = append(sql.selectColumns, types.ColumnSelector{
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
			var key string
			if len(from.As) > 0 {
				key = fmt.Sprintf("%s.%s", from.As, col.As)
			} else {
				key = col.As
			}
			sql.fromColumns[key] = columnIndex{
				source: sourceIdx,
				column: columnIdx,
			}
		}
	}

	// Bind select expressions.
	var idempotent = true
	for _, sel := range sql.Select {
		err := sel.Expr.Bind(sql)
		if err != nil {
			return nil, err
		}
		if !sel.Expr.IsIdempotent() {
			idempotent = false
		}
	}

	// Bind Where expressions.
	if sql.Where != nil {
		err := sql.Where.Bind(sql)
		if err != nil {
			return nil, err
		}
	}

	var matches [][]types.Row
	err := sql.eval(0, nil, nil, &matches)
	if err != nil {
		return nil, err
	}

	// Select result columns.

	var columns [][]types.ColumnSelector
	for _, sel := range sql.From {
		columns = append(columns, sel.Source.Columns())
	}

	for _, match := range matches {
		var row types.Row
		for i, sel := range sql.Select {
			if sel.IsPublic() {
				val, err := sel.Expr.Eval(match, columns, matches)
				if err != nil {
					return nil, err
				}
				if val == types.Null {
					row = append(row, types.NullColumn{})
				} else {
					str := val.String()
					row = append(row, types.StringColumn(str))
					sql.selectColumns[i].ResolveType(str)
				}
			}
		}
		sql.result = append(sql.result, row)
		if idempotent {
			break
		}
	}
	sql.evaluated = true

	return sql.result, nil
}

func (sql *Query) eval(idx int, data []types.Row,
	columns [][]types.ColumnSelector, result *[][]types.Row) error {

	if idx >= len(sql.From) {
		match := true
		if sql.Where != nil {
			val, err := sql.Where.Eval(data, columns, nil)
			if err != nil {
				return err
			}
			match, err = val.Bool()
			if err != nil {
				return err
			}
		}
		if match {
			*result = append(*result, data)
		}
		return nil
	}

	rows, err := sql.From[idx].Source.Get()
	if err != nil {
		return err
	}
	cols := sql.From[idx].Source.Columns()
	columns = append(columns, cols)

	for _, row := range rows {
		err := sql.eval(idx+1, append(data, row), columns, result)
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
