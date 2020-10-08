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

func TestCVSCorrect(t *testing.T) {
	source, err := NewCSV("test.csv", "", []ColumnSelector{
		{
			Name: Reference{
				Column: "0",
			},
			As: "Share",
		},
		{
			Name: Reference{
				Column: "1",
			},
			As: "Count",
		},
	})
	if err != nil {
		t.Fatalf("NewCSV failed: %s", err)
	}
	rows, err := source.Get()
	if err != nil {
		t.Fatalf("csv.Get() failed: %s", err)
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
