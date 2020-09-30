//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"github.com/PuerkitoBio/goquery"
)

// Query implements a document query.
type Query struct {
	Columns  []ColumnSelector
	Document string
	Where    Expr
}

// Execute executes the query against the argument document.
func (q *Query) Execute(s *goquery.Document,
	r func(columns []string) error) error {

	var lastError error

	columns := make(Columns)

	s.Find(q.Document).Each(func(i int, s *goquery.Selection) {
		for _, col := range q.Columns {
			columns[col.As] = s.Find(col.Find)
		}

		// Filter by condition.
		if q.Where != nil {
			v, err := q.Where.Eval(columns)
			if err != nil {
				lastError = err
				return
			}
			ok, err := v.Bool()
			if err != nil {
				lastError = err
				return
			}
			if !ok {
				return
			}
		}

		var result []string
		for _, col := range q.Columns {
			if col.IsPublic() {
				result = append(result, columns[col.As].Text())
			}
		}
		err := r(result)
		if err != nil {
			lastError = err
		}
	})

	return lastError
}

// Columns implements result columns.
type Columns map[string]*goquery.Selection
