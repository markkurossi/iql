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

	// GROUP BY tests:
	//
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
		v: [][]string{
			{"a", "3", "116"},
			{"b", "3", "66"},
			{"c", "2", "8"},
		},
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
		v: [][]string{
			{"a", "1", "200"},
			{"a", "2", "75"},
			{"b", "1", "50"},
			{"b", "2", "50"},
			{"b", "3", "100"},
			{"c", "1", "8"},
		},
	},
	{
		q: `
SELECT Name, Unit, Count,
       CASE
            WHEN Count >= 100 THEN 'Buy'
            WHEN Count >= 50 THEN  'Add'
            ELSE 'Hold'
       END AS Action
FROM (
	  SELECT "0" AS Name,
	         "1" AS Unit,
	         "2" AS Count
	  FROM 'data:text/csv;base64,YSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
     );`,
		v: [][]string{
			{"a", "1", "200", "Buy"},
			{"a", "2", "100", "Buy"},
			{"a", "2", "50", "Add"},
			{"b", "1", "50", "Add"},
			{"b", "2", "50", "Add"},
			{"b", "3", "100", "Buy"},
			{"c", "1", "10", "Hold"},
			{"c", "1", "7", "Hold"},
		},
	},
	{
		q: `
SELECT Name, Unit,
       CASE Unit
            WHEN 1 THEN 'R&D'
            WHEN 2 THEN 'Sales'
            ELSE 'HR'
       END AS Action
FROM (
	  SELECT "0" AS Name,
	         "1" AS Unit,
	         "2" AS Count
	  FROM 'data:text/csv;base64,YSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
     );`,
		v: [][]string{
			{"a", "1", "R&D"},
			{"a", "2", "Sales"},
			{"a", "2", "Sales"},
			{"b", "1", "R&D"},
			{"b", "2", "Sales"},
			{"b", "3", "HR"},
			{"c", "1", "R&D"},
			{"c", "1", "R&D"},
		},
	},

	// ORDER BY tests:
	//
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
SELECT Name, Unit, Count
FROM (
	  SELECT "0" AS Name,
	         "1" AS Unit,
	         "2" AS Count
	  FROM 'data:text/csv;base64,YSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
     )
ORDER BY Name;`,
		v: [][]string{
			{"a", "1", "200"},
			{"a", "2", "100"},
			{"a", "2", "50"},
			{"b", "1", "50"},
			{"b", "2", "50"},
			{"b", "3", "100"},
			{"c", "1", "10"},
			{"c", "1", "7"},
		},
	},
	{
		q: `
SELECT Name, Unit, Count
FROM (
	  SELECT "0" AS Name,
	         "1" AS Unit,
	         "2" AS Count
	  FROM 'data:text/csv;base64,YSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
     )
ORDER BY Name DESC;`,
		v: [][]string{
			{"c", "1", "10"},
			{"c", "1", "7"},
			{"b", "1", "50"},
			{"b", "2", "50"},
			{"b", "3", "100"},
			{"a", "1", "200"},
			{"a", "2", "100"},
			{"a", "2", "50"},
		},
	},
	{
		q: `
SELECT Name, Unit, Count
FROM (
	  SELECT "0" AS Name,
	         "1" AS Unit,
	         "2" AS Count
	  FROM 'data:text/csv;base64,YSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
     )
ORDER BY Name DESC, Unit DESC;`,
		v: [][]string{
			{"c", "1", "10"},
			{"c", "1", "7"},
			{"b", "3", "100"},
			{"b", "2", "50"},
			{"b", "1", "50"},
			{"a", "2", "100"},
			{"a", "2", "50"},
			{"a", "1", "200"},
		},
	},
	{
		q: `
SELECT Name, Unit, Count
FROM (
	  SELECT "0" AS Name,
	         "1" AS Unit,
	         "2" AS Count
	  FROM 'data:text/csv;base64,YSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
     )
ORDER BY Name DESC, Unit DESC, Count;`,
		v: [][]string{
			{"c", "1", "7"},
			{"c", "1", "10"},
			{"b", "3", "100"},
			{"b", "2", "50"},
			{"b", "1", "50"},
			{"a", "2", "50"},
			{"a", "2", "100"},
			{"a", "1", "200"},
		},
	},

	// Ints,Floats,Strings
	// 1,4.2,foo
	// 12,42.7,bar
	// 7,3.1415,zappa
	// ,2.75,x
	// 8,,y
	// 12,1.234,
	{
		q: `
SELECT Ints, Floats, Strings
FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQuMixmb28KMTIsNDIuNyxiYXIKNywzLjE0MTUsemFwcGEKLDIuNzUseAo4LCx5CjEyLDEuMjM0LAo=' FILTER 'headers'
ORDER BY Ints;`,
		v: [][]string{
			{"NULL", "2.75", "x"},
			{"1", "4.20", "foo"},
			{"7", "3.14", "zappa"},
			{"8", "NULL", "y"},
			{"12", "42.70", "bar"},
			{"12", "1.23", ""},
		},
	},
	{
		q: `
SELECT Ints, Floats, Strings
FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQuMixmb28KMTIsNDIuNyxiYXIKNywzLjE0MTUsemFwcGEKLDIuNzUseAo4LCx5CjEyLDEuMjM0LAo=' FILTER 'headers'
ORDER BY Floats;`,
		v: [][]string{
			{"8", "NULL", "y"},
			{"12", "1.23", ""},
			{"NULL", "2.75", "x"},
			{"7", "3.14", "zappa"},
			{"1", "4.20", "foo"},
			{"12", "42.70", "bar"},
		},
	},
	{
		q: `
SELECT Ints, Floats, Strings
FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQuMixmb28KMTIsNDIuNyxiYXIKNywzLjE0MTUsemFwcGEKLDIuNzUseAo4LCx5CjEyLDEuMjM0LAo=' FILTER 'headers'
ORDER BY Strings;`,
		v: [][]string{
			{"12", "1.23", ""},
			{"12", "42.70", "bar"},
			{"1", "4.20", "foo"},
			{"NULL", "2.75", "x"},
			{"8", "NULL", "y"},
			{"7", "3.14", "zappa"},
		},
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
				t.Fatalf("Parse failed: %v\nInput:\n%s\n", err, input.q)
			}

			if len(results) == 0 {
				tab, err := types.Tabulate(q, tabulate.Unicode)
				if err != nil {
					t.Fatalf("q.Get failed: %v\nInput:\n%s\n", err, input.q)
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
