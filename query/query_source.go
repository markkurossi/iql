//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"github.com/markkurossi/iql/types"
)

var (
	_ types.Source = &Source{}
)

// Source implements queries as data source. The queries are data
// sources already but they expose also non-exported columns. The
// Source filters the privat columns from the result columns.
type Source struct {
	columns []types.ColumnSelector
	rows    []types.Row
}

// NewSource creates a new query data source.
func NewSource(source types.Source) (*Source, error) {
	rows, err := source.Get()
	if err != nil {
		return nil, err
	}
	var columns []types.ColumnSelector
	for _, col := range source.Columns() {
		if col.IsPublic() {
			columns = append(columns, col)
		}
	}
	return &Source{
		columns: columns,
		rows:    rows,
	}, nil
}

// Columns implements the Source.Columns().
func (s *Source) Columns() []types.ColumnSelector {
	return s.columns
}

// Get implements the Source.Get().
func (s *Source) Get() ([]types.Row, error) {
	return s.rows, nil
}
