//
// Copyright (c) 2020-2021 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"encoding/csv"
	"errors"
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
func NewCSV(input []io.ReadCloser, filter string,
	columns []types.ColumnSelector) (types.Source, error) {

	for _, in := range input {
		defer in.Close()
	}

	// Parse filter options

	var err error
	var skip int
	var comment rune

	headers := true
	var prependHeaders []string
	trimLeadingSpace := true
	comma := ','

	for _, option := range strings.Split(filter, " ") {
		if len(option) == 0 {
			continue
		}
		parts := strings.Split(option, "=")
		switch len(parts) {
		case 1:
			switch parts[0] {
			case "keep-leading-space":
				trimLeadingSpace = false

			case "noheaders":
				headers = false

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
				comma = runes[0]

			case "comment":
				runes := []rune(parts[1])
				if len(runes) != 1 {
					return nil, fmt.Errorf("csv: comment must be rune: %s",
						parts[1])
				}
				comment = runes[0]

			case "prepend-headers":
				prependHeaders = strings.Split(parts[1], ",")

			default:
				return nil, fmt.Errorf("csv: unknown option: %s", parts[0])
			}

		default:
			return nil, fmt.Errorf("csv: invalid filter option: %s", option)
		}
	}

	var rows []types.Row
	var indices []int

	for idx, in := range input {
		reader := csv.NewReader(in)
		reader.Comment = comment
		reader.TrimLeadingSpace = trimLeadingSpace
		reader.Comma = comma

		if len(prependHeaders) > 0 {
			reader.FieldsPerRecord = -1
		}

		records, err := reader.ReadAll()
		if err != nil {
			return nil, err
		}
		if skip > len(records) {
			skip = len(records)
		}
		records = records[skip:]

		if idx == 0 {
			if headers {
				// Mapping from column names to column indices.
				if len(records) == 0 {
					return nil, errors.New("csv: no records")
				}

				r0 := append(prependHeaders, records[0]...)

				// Collect all column names; unselected columns are
				// appended to the source's columns array.
				seen := make(map[string]bool)
				for _, col := range columns {
					seen[col.Name.Column] = true
				}
				names := make(map[string]int)
				for idx, col := range r0 {
					names[col] = idx

					if !seen[col] {
						seen[col] = true
						columns = append(columns, types.ColumnSelector{
							Name: types.Reference{
								Column: col,
							},
						})
					}
				}

				for _, col := range columns {
					i, ok := names[col.Name.Column]
					if !ok {
						return nil, fmt.Errorf("csv: unknown column: %s",
							col.Name.Column)
					}
					indices = append(indices, i)
				}
			} else {
				if len(columns) == 0 {
					return nil, errors.New(
						"csv: 'SELECT *' not supported without headers")
				}
				for _, col := range columns {
					i, err := strconv.Atoi(col.Name.Column)
					if err != nil {
						return nil, err
					}
					indices = append(indices, i)
				}
			}
		}
		if headers {
			records = records[1:]
		}

		rows, err = processCSV(rows, records, indices, columns)
		if err != nil {
			return nil, err
		}
	}

	return &CSV{
		columns: columns,
		rows:    rows,
	}, nil
}

func processCSV(rows []types.Row, records [][]string, indices []int,
	columns []types.ColumnSelector) ([]types.Row, error) {

	for _, record := range records {
		var row types.Row
		for i := range columns {
			idx := indices[i]
			var val string

			if idx < 0 {
				if -idx <= len(record) {
					val = record[len(record)+idx]
				}
			} else {
				if idx < len(record) {
					val = record[idx]
				}
			}
			columns[i].ResolveString(val)
			row = append(row, types.StringColumn(val))
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// Columns implements the Source.Columns().
func (c *CSV) Columns() []types.ColumnSelector {
	return c.columns
}

// Get implements the Source.Get().
func (c *CSV) Get() ([]types.Row, error) {
	return c.rows, nil
}
