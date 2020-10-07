//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"fmt"
	"testing"
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
	for i, row := range rows {
		fmt.Printf("Row %d: %v\n", i, row)
	}
}
