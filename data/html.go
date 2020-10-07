//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"io"

	"github.com/PuerkitoBio/goquery"
)

// HTML implements a data source from HTML data.
type HTML struct {
	columns []ColumnSelector
	rows    []Row
}

// NewHTML creates a new HTML data source from the argument URL.
func NewHTML(url, filter string, columns []ColumnSelector) (Source, error) {
	input, err := openInput(url)
	if err != nil {
		return nil, err
	}
	return NewHTMLFromReader(input, filter, columns)
}

// NewHTMLFromReader creates a new HTML data source from the argument
// reader.
func NewHTMLFromReader(input io.ReadCloser, filter string,
	columns []ColumnSelector) (Source, error) {

	defer input.Close()
	doc, err := goquery.NewDocumentFromReader(input)
	if err != nil {
		return nil, err
	}

	var rows []Row

	doc.Find(filter).Each(func(i int, s *goquery.Selection) {
		var row Row
		for _, col := range columns {
			sel := s.Find(col.Name.Column)
			switch sel.Length() {
			case 0:
				row = append(row, StringColumn(""))

			case 1:
				row = append(row, StringColumn(sel.Text()))

			default:
				strings := sel.Map(func(i int, s *goquery.Selection) string {
					return s.Text()
				})
				row = append(row, StringsColumn(strings))
			}
		}
		rows = append(rows, row)
	})

	return &HTML{
		columns: columns,
		rows:    rows,
	}, nil
}

// Columns implements the Source.Columns().
func (html *HTML) Columns() []ColumnSelector {
	return html.columns
}

// Get implements the Source.Get().
func (html *HTML) Get() ([]Row, error) {
	return html.rows, nil
}
