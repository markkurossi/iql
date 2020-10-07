//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"encoding/csv"
	"fmt"
	"strconv"
)

// CSV implements a data source from comma-separated values (CSV).
type CSV struct {
	columns []ColumnSelector
	rows    []Row
}

// NewCSV creates a new CSV data source from the argument URL.
func NewCSV(url, filter string, columns []ColumnSelector) (Source, error) {
	input, err := openInput(url)
	if err != nil {
		return nil, err
	}
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

	var rows []Row

	for _, record := range records {
		var row Row
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
			row = append(row, StringColumn(val))
		}
		rows = append(rows, row)
	}

	return &CSV{
		columns: columns,
		rows:    rows,
	}, nil
}

// Columns implements the Source.Columns().
func (c *CSV) Columns() []ColumnSelector {
	return c.columns
}

// Get implements the Source.Get().
func (c *CSV) Get() ([]Row, error) {
	return c.rows, nil
}
