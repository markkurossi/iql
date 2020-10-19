//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/markkurossi/iql/types"
)

// CSV implements a data source from comma-separated values (CSV).
type CSV struct {
	columns []types.ColumnSelector
	rows    []types.Row
}

// NewCSV creates a new CSV data source from the input.
func NewCSV(input io.ReadCloser, filter string,
	columns []types.ColumnSelector) (types.Source, error) {

	defer input.Close()

	reader := csv.NewReader(input)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var indices []int

	for _, col := range columns {
		i, err := strconv.Atoi(col.Name.Column)
		if err != nil {
			return nil, err
		}
		indices = append(indices, i)
	}

	var rows []types.Row

	for _, record := range records {
		var row types.Row
		for i := range columns {
			idx := indices[i]
			var val string

			if idx < 0 {
				if -idx > len(record) {
					return nil,
						fmt.Errorf("csv: index %d (%d) out of bounds", i, idx)
				}
				val = record[len(record)+idx]
			} else {
				if idx >= len(record) {
					return nil,
						fmt.Errorf("csv: index %d (%d) out of bounds", i, idx)
				}
				val = record[idx]
			}
			columns[i].ResolveType(val)
			row = append(row, types.StringColumn(val))
		}
		rows = append(rows, row)
	}

	return &CSV{
		columns: columns,
		rows:    rows,
	}, nil
}

// Columns implements the Source.Columns().
func (c *CSV) Columns() []types.ColumnSelector {
	return c.columns
}

// Get implements the Source.Get().
func (c *CSV) Get() ([]types.Row, error) {
	return c.rows, nil
}
