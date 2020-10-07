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
	for i, row := range rows {
		fmt.Printf("Row %d: %v\n", i, row)
	}
}
