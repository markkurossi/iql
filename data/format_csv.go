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
	"strings"

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

	// Parse filter options

	var err error
	var skip int
	var headers bool

	for _, option := range strings.Split(filter, " ") {
		if len(option) == 0 {
			continue
		}
		parts := strings.Split(option, "=")
		switch len(parts) {
		case 1:
			switch parts[0] {
			case "trim-leading-space":
				reader.TrimLeadingSpace = true

			case "headers":
				headers = true

			default:
				return nil, fmt.Errorf("csv: invalid filter flag: %s", parts[0])
			}

		case 2:
			switch parts[0] {
			case "skip":
				skip, err = strconv.Atoi(parts[1])
				if err != nil {
					return nil, fmt.Errorf("csv: invalid skip count: %s",
						parts[1])
				}

			case "comma":
				runes := []rune(parts[1])
				if len(runes) != 1 {
					return nil, fmt.Errorf("csv: comma must be rune: %s",
						parts[1])
				}
				reader.Comma = runes[0]

			case "comment":
				runes := []rune(parts[1])
				if len(runes) != 1 {
					return nil, fmt.Errorf("csv: comment must be rune: %s",
						parts[1])
				}
				reader.Comment = runes[0]

			default:
				return nil, fmt.Errorf("csv: unknown option: %s", parts[0])
			}

		default:
			return nil, fmt.Errorf("csv: invalid filter option: %s", option)
		}
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if skip > len(records) {
		skip = len(records)
	}
	records = records[skip:]

	var indices []int

	if headers {
		// Mapping from column names to column indices.
		names := make(map[string]int)
		for idx, col := range records[0] {
			names[col] = idx
		}
		records = records[1:]

		for _, col := range columns {
			i, ok := names[col.Name.Column]
			if !ok {
				return nil, fmt.Errorf("cvs: unknown column: %s",
					col.Name.Column)
			}
			indices = append(indices, i)
		}
	} else {
		for _, col := range columns {
			i, err := strconv.Atoi(col.Name.Column)
			if err != nil {
				return nil, err
			}
			indices = append(indices, i)
		}
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
			columns[i].ResolveString(val)
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
