//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/jsonq"
)

// JSONFilter filters the input with the CSS selector string and
// returns the matching elements.
func JSONFilter(input io.ReadCloser, filter string) ([]interface{}, error) {
	defer input.Close()

	data, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, err
	}
	var v interface{}
	err = json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	return jsonq.Ctx(v).Select(filter).Get()
}

// JSON implements a data source from JavaScript Object Notation (JSON).
type JSON struct {
	columns []types.ColumnSelector
	rows    []types.Row
}

// NewJSON creates a new JSON data source from the input.
func NewJSON(input []io.ReadCloser, filter string,
	columns []types.ColumnSelector) (types.Source, error) {

	for _, in := range input {
		defer in.Close()
	}

	var rows []types.Row

	for idx, in := range input {
		data, err := ioutil.ReadAll(in)
		if err != nil {
			return nil, err
		}
		var v interface{}
		err = json.Unmarshal(data, &v)
		if err != nil {
			return nil, err
		}
		filtered, err := jsonq.Ctx(v).Select(filter).Get()
		if err != nil {
			return nil, err
		}
		if len(filtered) == 0 {
			continue
		}

		if idx == 0 && len(columns) == 0 {
			// SELECT *
			switch obj := filtered[0].(type) {
			case map[string]interface{}:
				for col := range obj {
					columns = append(columns, types.ColumnSelector{
						Name: types.Reference{
							Column: col,
						},
					})
				}
				sort.Slice(columns, func(i, j int) bool {
					return columns[i].Name.Column < columns[j].Name.Column
				})

			default:
				return nil, errors.New("json: 'SELECT *' not supported")
			}
		}

		rows, err = processJSON(filtered, rows, filter, columns)
		if err != nil {
			return nil, err
		}
	}

	return &JSON{
		columns: columns,
		rows:    rows,
	}, nil
}

func processJSON(filtered []interface{}, rows []types.Row, filter string,
	columns []types.ColumnSelector) ([]types.Row, error) {

	for _, f := range filtered {
		var row types.Row
		for i, col := range columns {
			sel, err := jsonq.Get(f, col.Name.Column)
			if err != nil {
				return nil, err
			}
			row = append(row,
				types.StringColumn(strings.TrimSpace(fmt.Sprintf("%v", sel))))
			columns[i].ResolveString(row[i].String())
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// Columns implements the Source.Columns().
func (src *JSON) Columns() []types.ColumnSelector {
	return src.columns
}

// Get implements the Source.Get().
func (src *JSON) Get() ([]types.Row, error) {
	return src.rows, nil
}
