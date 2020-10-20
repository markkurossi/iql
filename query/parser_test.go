//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/markkurossi/iql/types"
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
	`
select COUNT(Year) as Count
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     );`,

	`
select Year,
       Value,
       Year * Value as Sum
into data
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     );
select Year, Sum from data;`,

	// 1,4.1
	// 2,4.2
	// 3,4.3
	// 4,4.4
	`
SELECT
        "0" AS Ints,
        "1" AS Floats,
        Ints + Floats AS Sum1,
        Floats + Ints AS Sum2
FROM 'data:text/csv;base64,MSw0LjEKMiw0LjIKMyw0LjMKNCw0LjQK';`,
}

func TestParser(t *testing.T) {
	for _, input := range parserTests {
		parser := NewParser(bytes.NewReader([]byte(input)), "{data}")
		for {
			q, err := parser.Parse()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("Parse failed: %v", err)
			}

			rows, err := q.Get()
			if err != nil {
				t.Fatalf("q.Get failed: %v", err)
			}
			tab := types.Tabulate(q, tabulate.Unicode)
			for _, columns := range rows {
				row := tab.Row()
				for _, col := range columns {
					row.Column(col.String())
				}
			}

			tab.Print(os.Stdout)
		}
	}
}
