//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"

	"github.com/markkurossi/htmlq/data"
)

var (
	_ data.Source = &Query{}
)

// Query implements an SQL query. It also implements data.Source so
// that the query can be used as a nested data source for other
// queries.
type Query struct {
	Select  []data.ColumnSelector
	From    []SourceSelector
	Where   Expr
	columns map[string]columnIndex
}

// SourceSelector defines an input source with an optional name alias.
type SourceSelector struct {
	Source data.Source
	As     string
}

// Columns implements the Source.Columns().
func (sql *Query) Columns() []data.ColumnSelector {
	return sql.Select
}

// Get implements the Source.Get().
func (sql *Query) Get() ([]data.Row, error) {
	// Collect column names.
	sql.columns = make(map[string]columnIndex)
	for sourceIdx, from := range sql.From {
		for columnIdx, column := range from.Source.Columns() {
			key := fmt.Sprintf("%s.%s", from.As, column.As)
			sql.columns[key] = columnIndex{
				source: sourceIdx,
				column: columnIdx,
			}
		}
	}

	// Resolve columns into references.

	var references []*Reference

	for _, sel := range sql.Select {
		ref, err := sql.resolveName(sel.Name, sel.IsPublic())
		if err != nil {
			return nil, err
		}
		references = append(references, ref)
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
	// XXX the select references should be expressions and this step
	// would evaluate expressions against each row.
	var result []data.Row
	for _, match := range matches {
		var row data.Row
		for _, ref := range references {
			if ref.public {
				row = append(row, match[ref.index.source][ref.index.column])
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
		index, ok := sql.columns[name.String()]
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
		index, ok := sql.columns[key.String()]
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
