//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"

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

	return nil, errors.New("json: not implemented yet")
}

// Columns implements the Source.Columns().
func (src *JSON) Columns() []types.ColumnSelector {
	return src.columns
}

// Get implements the Source.Get().
func (src *JSON) Get() ([]types.Row, error) {
	return src.rows, nil
}
