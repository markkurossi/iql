//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/markkurossi/iql/types"
)

// HTML implements a data source from HTML data.
type HTML struct {
	columns []types.ColumnSelector
	rows    []types.Row
}

// NewHTML creates a new HTML data source from the input.
func NewHTML(input io.ReadCloser, filter string,
	columns []types.ColumnSelector) (types.Source, error) {

	defer input.Close()

	doc, err := goquery.NewDocumentFromReader(input)
	if err != nil {
		return nil, err
	}

	var rows []types.Row

	doc.Find(filter).Each(func(i int, s *goquery.Selection) {
		var row types.Row
		for i, col := range columns {
			sel := s.Find(col.Name.Column)
			switch sel.Length() {
			case 0:
				row = append(row, types.StringColumn(""))

			case 1:
				row = append(row,
					types.StringColumn(strings.TrimSpace(sel.Text())))

			default:
				strings := sel.Map(func(i int, s *goquery.Selection) string {
					return s.Text()
				})
				row = append(row, types.StringsColumn(strings))
			}
			columns[i].ResolveString(row[i].String())
		}
		rows = append(rows, row)
	})

	return &HTML{
		columns: columns,
		rows:    rows,
	}, nil
}

// Columns implements the Source.Columns().
func (html *HTML) Columns() []types.ColumnSelector {
	return html.columns
}

// Get implements the Source.Get().
func (html *HTML) Get() ([]types.Row, error) {
	return html.rows, nil
}
