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

func TestCSVCorrect(t *testing.T) {
	name := "test.csv"
	source, err := New([]string{name}, "noheaders", []types.ColumnSelector{
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
	if len(rows) != 3 {
		t.Errorf("%s: unexpected number of rows", name)
	}
	if len(rows[0]) != 2 {
		t.Errorf("%s: unexpected number of columns", name)
	}
	tab, err := types.Tabulate(source, tabulate.Unicode)
	if err != nil {
		t.Errorf("%s: tabulate failed: %s", name, err)
	}
	tab.Print(os.Stdout)
}

func TestCSVOptions(t *testing.T) {
	source, err := New([]string{"test_options.csv"},
		"noheaders skip=1 comma=;  comment=#",
		[]types.ColumnSelector{
			{
				Name: types.Reference{
					Column: "0",
				},
				As: "Year",
			},
			{
				Name: types.Reference{
					Column: "1",
				},
				As: "Value",
			},
			{
				Name: types.Reference{
					Column: "2",
				},
				As: "Delta",
			},
		})
	if err != nil {
		t.Fatalf("NewCSV failed: %s", err)
	}
	tab, err := types.Tabulate(source, tabulate.Unicode)
	if err != nil {
		t.Fatalf("csv.Get() failed: %s", err)
	}
	tab.Print(os.Stdout)
}
