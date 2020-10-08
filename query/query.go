//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"
	"unicode"

	"github.com/markkurossi/htmlq/data"
	"github.com/markkurossi/tabulate"
)

var (
	_ data.Source = &Query{}
)

// Query implements an SQL query. It also implements data.Source so
// that the query can be used as a nested data source for other
// queries.
type Query struct {
	Select        []ColumnSelector
	From          []SourceSelector
	Where         Expr
	selectColumns []data.ColumnSelector
	fromColumns   map[string]columnIndex
}

// ColumnSelector defines selected query output columns.
type ColumnSelector struct {
	Expr  Expr
	As    string
	Align tabulate.Align
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

// SourceSelector defines an input source with an optional name alias.
type SourceSelector struct {
	Source data.Source
	As     string
}

// Columns implements the Source.Columns().
func (sql *Query) Columns() []data.ColumnSelector {
	if len(sql.selectColumns) == 0 {
		for _, col := range sql.Select {
			sql.selectColumns = append(sql.selectColumns, data.ColumnSelector{
				Name: data.Reference{
					Column: col.Expr.String(),
				},
				As:    col.As,
				Align: col.Align,
			})
		}
	}
	return sql.selectColumns
}

// Get implements the Source.Get().
func (sql *Query) Get() ([]data.Row, error) {
	// Collect column names.
	sql.fromColumns = make(map[string]columnIndex)
	for sourceIdx, from := range sql.From {
		for columnIdx, column := range from.Source.Columns() {
			key := fmt.Sprintf("%s.%s", from.As, column.As)
			sql.fromColumns[key] = columnIndex{
				source: sourceIdx,
				column: columnIdx,
			}
		}
	}

	// Bind Select expressions.
	for _, sel := range sql.Select {
		err := sel.Expr.Bind(sql)
		if err != nil {
			return nil, err
		}
	}

	// Bind Where expressions.
	if sql.Where != nil {
		err := sql.Where.Bind(sql)
		if err != nil {
			return nil, err
		}
	}

	var matches [][]data.Row
	err := sql.eval(0, nil, &matches)
	if err != nil {
		return nil, err
	}

	// Select result columns.
	var result []data.Row
	for _, match := range matches {
		var row data.Row
		for _, sel := range sql.Select {
			if sel.IsPublic() {
				val, err := sel.Expr.Eval(match)
				if err != nil {
					return nil, err
				}
				row = append(row, data.StringColumn(val.String()))
			}
		}
		result = append(result, row)
	}

	return result, nil
}

func (sql *Query) eval(idx int, data []data.Row, result *[][]data.Row) error {

	if idx >= len(sql.From) {
		match := true
		if sql.Where != nil {
			val, err := sql.Where.Eval(data)
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
	for _, row := range rows {
		err := sql.eval(idx+1, append(data, row), result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sql *Query) resolveName(name data.Reference, public bool) (
	*Reference, error) {

	if name.IsAbsolute() {
		index, ok := sql.fromColumns[name.String()]
		if !ok {
			return nil, fmt.Errorf("Unknown column '%s'", name.String())
		}
		return &Reference{
			Reference: name,
			index:     index,
			public:    public,
		}, nil
	}

	var match *Reference

	for _, from := range sql.From {
		key := data.Reference{
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
	if match == nil {
		return nil, fmt.Errorf("unknown column name '%s'", name)
	}

	return match, nil
}
