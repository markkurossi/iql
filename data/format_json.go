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

	// XXX this could work
	if len(columns) == 0 {
		return nil, errors.New("json: 'SELECT *' not supported")
	}

	var rows []types.Row
	var err error

	for _, in := range input {
		rows, err = processJSON(in, rows, filter, columns)
		if err != nil {
			return nil, err
		}
	}

	return &JSON{
		columns: columns,
		rows:    rows,
	}, nil
}

func processJSON(in io.ReadCloser, rows []types.Row, filter string,
	columns []types.ColumnSelector) ([]types.Row, error) {

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
