//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"os"
	"testing"

	"github.com/markkurossi/htmlq/data"
	"github.com/markkurossi/htmlq/query"
	"github.com/markkurossi/tabulate"
)

func TestJoin(t *testing.T) {
	ref, err := data.NewHTML("data/test.html", "tbody > tr",
		[]data.ColumnSelector{
			{
				Name: data.Reference{
					Column: ".stock",
				},
				As: "Stock",
			},
			{
				Name: data.Reference{
					Column: ".price",
				},
				As: "Price",
			},
			{
				Name: data.Reference{
					Column: ".share",
				},
				As: "Weight",
			},
		})
	if err != nil {
		t.Fatalf("NewHTML failed: %s", err)
	}
	portfolio, err := data.NewCSV("data/test.csv", "", []data.ColumnSelector{
		{
			Name: data.Reference{
				Column: "0",
			},
			As: "Stock",
		},
		{
			Name: data.Reference{
				Column: "1",
			},
			As: "Count",
		},
	})
	if err != nil {
		t.Fatalf("NewHTML failed: %s", err)
	}

	q := &query.Query{
		Select: []query.ColumnSelector{
			{
				Expr: &query.Reference{
					Reference: data.Reference{
						Source: "ref",
						Column: "Stock",
					},
				},
				As: "Stock",
			},
			{
				Expr: &query.Reference{
					Reference: data.Reference{
						Source: "ref",
						Column: "Price",
					},
				},
				As: "Price",
			},
			{
				Expr: &query.Reference{
					Reference: data.Reference{
						Source: "ref",
						Column: "Weight",
					},
				},
				As: "Weight",
			},
			{
				Expr: &query.Reference{
					Reference: data.Reference{
						Source: "portfolio",
						Column: "Count",
					},
				},
				As: "Count",
			},
		},
		From: []query.SourceSelector{
			{
				Source: ref,
				As:     "ref",
			},
			{
				Source: portfolio,
				As:     "portfolio",
			},
		},
		Where: &query.Binary{
			Type: query.BinAnd,
			Left: &query.Binary{
				Type: query.BinNeq,
				Left: &query.Reference{
					Reference: data.Reference{
						Source: "ref",
						Column: "Stock",
					},
				},
				Right: &query.Constant{
					Value: data.StringValue(""),
				},
			},
			Right: &query.Binary{
				Type: query.BinEq,
				Left: &query.Reference{
					Reference: data.Reference{
						Source: "ref",
						Column: "Stock",
					},
				},
				Right: &query.Reference{
					Reference: data.Reference{
						Source: "portfolio",
						Column: "Stock",
					},
				},
			},
		},
	}
	rows, err := q.Get()
	if err != nil {
		t.Fatalf("query.Get() failed: %s", err)
	}
	tab := data.Table(q, tabulate.Unicode)
	for _, columns := range rows {
		row := tab.Row()
		for _, col := range columns {
			row.Column(col.String())
		}
	}
	tab.Print(os.Stdout)
}
