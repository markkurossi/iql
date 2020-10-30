//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

var parserTests = []IQLTest{
	{
		q: `select 1 + 0x01 + 0b10 + 077 + 0o70 as Sum, 100-42 as Diff;`,
		v: [][]string{{"123", "58"}},
	},

	// 2008,100
	// 2009,101
	// 2010,200
	{
		q: `select "0" As Year, "1" as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK';`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data;`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year > 2009;`,
		v: [][]string{
			{"2010", "200"},
		},
	},
	{
		q: `select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year = 2009;`,
		v: [][]string{
			{"2009", "101"},
		},
	},
	{
		q: `select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year >= 2009;`,
		v: [][]string{
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year < 2009;`,
		v: [][]string{
			{"2008", "100"},
		},
	},
	{
		q: `select Data.0 As Year, Data.1 as Value
from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK' as Data
where Data.Year <= 2009;`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
		},
	},
	{
		q: `
select Year, Value
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     ) as Data;`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `
select Data.Year, Data.Value
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     ) as Data;`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `
select Year, Value
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     );`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `
select Year as Y, Value as V
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     );`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `
declare data varchar;
set data = 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK';
select Year as Y, Value as V
from (
        select "0" AS Year,
               "1" AS Value
        from data
     );`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `
select SUM(Year) as Sum
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     );`,
		v: [][]string{{"6027"}},
	},
	{
		q: `
select COUNT(Year) as Count
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     );`,
		v: [][]string{{"3"}},
	},

	{
		q: `
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
		v: [][]string{
			{"2008", "100", "200800"},
			{"2009", "101", "202909"},
			{"2010", "200", "402000"},
		},
		rest: [][][]string{
			{
				{"2008", "200800"},
				{"2009", "202909"},
				{"2010", "402000"},
			},
		},
	},
	{
		q: `
select Year,
       Value,
       Year * Value as Sum
into data
from (
        select "0" AS Year,
               "1" AS Value
        from 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
     );

select data.Year, data.Sum from data;`,
		v: [][]string{
			{"2008", "100", "200800"},
			{"2009", "101", "202909"},
			{"2010", "200", "402000"},
		},
		rest: [][][]string{
			{
				{"2008", "200800"},
				{"2009", "202909"},
				{"2010", "402000"},
			},
		},
	},

	// Region,Unit,Count
	// a,1,200
	// a,2,100
	// a,2,50
	// b,1,50
	// b,2,50
	// b,3,100
	// c,1,10
	// c,1,7
	{
		q: `
SELECT Region, Unit
FROM 'data:text/csv;base64,UmVnaW9uLFVuaXQsQ291bnQKYSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
     FILTER 'headers';`,
		v: [][]string{
			{"a", "1"},
			{"a", "2"},
			{"a", "2"},
			{"b", "1"},
			{"b", "2"},
			{"b", "3"},
			{"c", "1"},
			{"c", "1"},
		},
	},

	// 1,4.1
	// 2,4.2
	// 3,4.3
	// 4,4.4
	{
		q: `
SELECT
        "0" AS Ints,
        "1" AS Floats,
        Ints + Floats AS Sum1,
        Floats + Ints AS Sum2
FROM 'data:text/csv;base64,MSw0LjEKMiw0LjIKMyw0LjMKNCw0LjQK';`,
		v: [][]string{
			{"1", "4.10", "5.10", "5.10"},
			{"2", "4.20", "6.20", "6.20"},
			{"3", "4.30", "7.30", "7.30"},
			{"4", "4.40", "8.40", "8.40"},
		},
	},
	{
		q: `SELECT 'Hello: ' + 1 + ', ' + 1.2 + ', ' + false AS Message;`,
		v: [][]string{{"Hello: 1, 1.20, false"}},
	},
	{
		q: `PRINT 'GROUP BY tests:';`,
	},

	// a,1,200
	// a,2,100
	// a,2,50
	// b,1,50
	// b,2,50
	// b,3,100
	// c,1,10
	// c,1,7
	{
		q: `
SELECT Name,
       COUNT(Unit) as Count,
       AVG(Count) as Avg
FROM (
	  SELECT "0" AS Name,
	         "1" AS Unit,
	         "2" AS Count
	  FROM 'data:text/csv;base64,YSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
     )
GROUP BY Name;`,
	},
	{
		q: `
SELECT Name,
       Unit,
       AVG(Count) as Avg
FROM (
	  SELECT "0" AS Name,
	         "1" AS Unit,
	         "2" AS Count
	  FROM 'data:text/csv;base64,YSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
     )
GROUP BY Name, Unit;`,
	},
}

func TestParser(t *testing.T) {
	for testID, input := range parserTests {
		name := fmt.Sprintf("Test %d", testID)
		parser := NewParser(bytes.NewReader([]byte(input.q)), name)

		var results [][][]string

		if input.v != nil {
			results = append(results, input.v)
		}
		results = append(results, input.rest...)

		for {
			q, err := parser.Parse()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("Parse failed: %v\nInput:\n%s\n", err, input)
			}

			if len(results) == 0 {
				tab, err := types.Tabulate(q, tabulate.Unicode)
				if err != nil {
					t.Fatalf("q.Get failed: %v\nInput:\n%s\n", err, input)
				}
				if true {
					tab.Print(os.Stdout)
				}
			} else {
				verifyResult(t, name, q, results[0])
				results = results[1:]
			}
		}
	}
}
