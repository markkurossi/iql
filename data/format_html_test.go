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

func TestHTMLCorrect(t *testing.T) {
	source, err := New("test.html", "tbody > tr", []types.ColumnSelector{
		{
			Name: types.Reference{
				Column: ".stock",
			},
			As: "Stock",
		},
		{
			Name: types.Reference{
				Column: ".price",
			},
			As: "Price",
		},
		{
			Name: types.Reference{
				Column: ".share",
			},
			As: "Share",
		},
	})
	if err != nil {
		t.Fatalf("New failed: %s", err)
	}
	tab, err := types.Tabulate(source, tabulate.Unicode)
	if err != nil {
		t.Fatalf("html.Get() failed: %s", err)
	}
	tab.Print(os.Stdout)
}
