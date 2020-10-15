//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"bytes"
	"os"
	"testing"

	"github.com/markkurossi/iql/data"
	"github.com/markkurossi/tabulate"
)

var parserTests = []string{
	`select 1 + 0x01 + 0b10 + 077 + 0o70 as Sum, 100-42 as Diff`,

	// 2008,100
	// 2009,101
	// 2010,200
	`select "0" As Year, "1" as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'`,
	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data`,

	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year > 2009`,
	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year = 2009`,
	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year >= 2009`,

	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year < 2009`,
	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year <= 2009`,
}

func TestParser(t *testing.T) {
	for _, input := range parserTests {
		q, err := Parse(bytes.NewReader([]byte(input)), "{data}")
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		rows, err := q.Get()
		if err != nil {
			t.Fatalf("q.Get failed: %v", err)
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
}
