//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"os"
	"testing"

	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

func TestCVSCorrect(t *testing.T) {
	source, err := New("test.csv", "", []types.ColumnSelector{
		{
			Name: types.Reference{
				Column: "0",
			},
			As: "Share",
		},
		{
			Name: types.Reference{
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
	tab := types.Tabulate(source, tabulate.Unicode)
	for _, columns := range rows {
		row := tab.Row()
		for _, col := range columns {
			row.Column(col.String())
		}
	}
	tab.Print(os.Stdout)
}
