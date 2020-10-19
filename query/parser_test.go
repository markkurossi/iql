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
	`select 1 + 0x01 + 0b10 + 077 + 0o70 as Sum, 100-42 as Diff;`,

	// 2008,100
	// 2009,101
	// 2010,200
	`select "0" As Year, "1" as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK';`,
	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data;`,

	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year > 2009;`,
	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year = 2009;`,
	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year >= 2009;`,

	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year < 2009;`,
	`select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year <= 2009;`,

	`
select Year, Value
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     ) as Data;`,
	`
select Data.Year, Data.Value
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     ) as Data;`,
	`
select Year, Value
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     );`,
	`
select Year as Y, Value as V
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     );`,
	`
declare data varchar;
set data = 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK';
select Year as Y, Value as V
from (
        select "0" AS Year,
               "1" AS Value
        from data
     );`,
	`
select SUM(Year) as Sum
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     );`,
}

func TestParser(t *testing.T) {
	for _, input := range parserTests {
		parser := NewParser(bytes.NewReader([]byte(input)), "{data}")
		q, err := parser.Parse()
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
