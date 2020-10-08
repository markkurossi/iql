//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"os"
	"testing"

	"github.com/markkurossi/tabulate"
)

func TestHTMLCorrect(t *testing.T) {
	source, err := NewHTML("test.html", "tbody > tr", []ColumnSelector{
		{
			Name: Reference{
				Column: ".stock",
			},
			As: "Stock",
		},
		{
			Name: Reference{
				Column: ".price",
			},
			As: "Price",
		},
		{
			Name: Reference{
				Column: ".share",
			},
			As: "Share",
		},
	})
	if err != nil {
		t.Fatalf("NewHTML failed: %s", err)
	}
	rows, err := source.Get()
	if err != nil {
		t.Fatalf("html.Get() failed: %s", err)
	}
	tab := Table(source, tabulate.Unicode)
	for _, columns := range rows {
		row := tab.Row()
		for _, col := range columns {
			row.Column(col.String())
		}
	}
	tab.Print(os.Stdout)
}
