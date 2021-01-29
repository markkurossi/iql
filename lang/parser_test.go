//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package lang

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

var parserTests = []struct {
	q    string
	v    [][]string
	rest [][][]string
}{
	{
		q: `SELECT 42;`,
		v: [][]string{{"42"}},
	},
	{
		q: `SELECT 3.14;`,
		v: [][]string{{"3.14"}},
	},
	{
		q: `SELECT 0b101, 0B101;`,
		v: [][]string{{"5", "5"}},
	},
	{
		q: `SELECT 0o0777, 0O0777, 0777;`,
		v: [][]string{{"511", "511", "511"}},
	},
	{
		q: `SELECT 0xdeadbeef, 0xdeadbeef;`,
		v: [][]string{{"3735928559", "3735928559"}},
	},
	{
		q: `SELECT -1;`,
		v: [][]string{{"-1"}},
	},
	{
		q: `SELECT -3.14;`,
		v: [][]string{{"-3.14"}},
	},
	{
		q: `SELECT 'Cincinnati';`,
		v: [][]string{{"Cincinnati"}},
	},
	{
		q: `SELECT 'O''Brien';`,
		v: [][]string{{"O'Brien"}},
	},
	{
		q: `SELECT 'Process X is 50% complete.';`,
		v: [][]string{{"Process X is 50% complete."}},
	},
	{
		q: `SELECT 'The level for job_id: %d should be between %d and %d.';`,
		v: [][]string{{
			"The level for job_id: %d should be between %d and %d.",
		}},
	},
	{
		q: `SELECT 1 + 0x01 + 0b10 + 077 + 0o70 AS Sum, 100-42 AS Diff;`,
		v: [][]string{{"123", "58"}},
	},
	{
		q: `SELECT 'foo bar baz' ~ '\bbar\b';`,
		v: [][]string{{"true"}},
	},
	{
		q: `SELECT 'foo bar baz' ~ '\bbAr\b';`,
		v: [][]string{{"false"}},
	},
	{
		q: `SELECT 'foo bar baz' ~ '(?i)\bbAr\b';`,
		v: [][]string{{"true"}},
	},
	{
		q: `SELECT 'foo bar baz' !~ '^bar';`,
		v: [][]string{{"true"}},
	},

	// 2008,100
	// 2009,101
	// 2010,200
	{
		q: `SELECT "0" AS Year, "1" AS Value
FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
FILTER 'noheaders';`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `SELECT "0" AS Year, "1" AS [,Value]
FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
FILTER 'noheaders';`,
		v: [][]string{
			{"2008"},
			{"2009"},
			{"2010"},
		},
	},
	{
		q: `SELECT Data.0 AS Year, Data.1 AS Value
FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
FILTER 'noheaders' AS Data;`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `SELECT Data.0 AS Year, Data.1 AS Value
FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
FILTER 'noheaders' AS Data
where Data.0 > 2009;`,
		v: [][]string{
			{"2010", "200"},
		},
	},
	{
		q: `SELECT Data.0 AS Year, Data.1 AS Value
FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
FILTER 'noheaders' AS Data
where Data.0 = 2009;`,
		v: [][]string{
			{"2009", "101"},
		},
	},
	{
		q: `SELECT Data.0 AS Year, Data.1 AS Value
FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
FILTER 'noheaders' AS Data
where Data.0 >= 2009;`,
		v: [][]string{
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `SELECT Data.0 AS Year, Data.1 AS Value
FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
FILTER 'noheaders' AS Data
where Data.0 < 2009;`,
		v: [][]string{
			{"2008", "100"},
		},
	},
	{
		q: `SELECT Data.0 AS Year, Data.1 AS Value
FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
FILTER 'noheaders' AS Data
where Data.0 <= 2009;`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
		},
	},
	{
		q: `
SELECT Year, Value
FROM (
        SELECT "0" AS Year,
               "1" AS Value
        FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
        FILTER 'noheaders'
     ) AS Data;`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `
SELECT Data.Year, Data.Value
FROM (
        SELECT "0" AS Year,
               "1" AS Value
        FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
        FILTER 'noheaders'
     ) AS Data;`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `
SELECT Year, Value
FROM (
        SELECT "0" AS Year,
               "1" AS Value
        FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
        FILTER 'noheaders'
     );`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},
	{
		q: `
SELECT Year AS Y, Value AS V
FROM (
        SELECT "0" AS Year,
               "1" AS Value
        FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
        FILTER 'noheaders'
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
SELECT Year AS Y, Value AS V
FROM (
        SELECT "0" AS Year,
               "1" AS Value
        FROM data FILTER 'noheaders'
     );`,
		v: [][]string{
			{"2008", "100"},
			{"2009", "101"},
			{"2010", "200"},
		},
	},

	{
		q: `
SELECT Year,
       Value,
       Year * Value AS Sum
into data
FROM (
        SELECT "0" AS Year,
               "1" AS Value
        FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
        FILTER 'noheaders'
     );
SELECT Year, Sum FROM data;`,
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
SELECT Year,
       Value,
       Year * Value AS Sum
into data
FROM (
        SELECT "0" AS Year,
               "1" AS Value
        FROM 'data:text/csv;base64,MjAwOCwxMDAKMjAwOSwxMDEKMjAxMCwyMDAK'
        FILTER 'noheaders'
     );

SELECT data.Year, data.Sum FROM data;`,
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
FROM 'data:text/csv;base64,UmVnaW9uLFVuaXQsQ291bnQKYSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg==';`,
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
        "0" + "1" AS Sum1,
        "1" + "0" AS Sum2
FROM 'data:text/csv;base64,MSw0LjEKMiw0LjIKMyw0LjMKNCw0LjQK'
FILTER 'noheaders';`,
		v: [][]string{
			{"1", "4.1", "5.1", "5.1"},
			{"2", "4.2", "6.2", "6.2"},
			{"3", "4.3", "7.3", "7.3"},
			{"4", "4.4", "8.4", "8.4"},
		},
	},
	{
		q: `SELECT 'Hello: ' + 1 + ', ' + 1.2 + ', ' + false AS Message;`,
		v: [][]string{{"Hello: 1, 1.2, false"}},
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
      FILTER 'noheaders'
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
      FILTER 'noheaders'
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
      FILTER 'noheaders'
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
      FILTER 'noheaders'
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
      FILTER 'noheaders'
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
      FILTER 'noheaders'
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
FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQuMixmb28KMTIsNDIuNyxiYXIKNywzLjE0MTUsemFwcGEKLDIuNzUseAo4LCx5CjEyLDEuMjM0LAo='
ORDER BY Ints;`,
		v: [][]string{
			{"NULL", "2.75", "x"},
			{"1", "4.2", "foo"},
			{"7", "3.1415", "zappa"},
			{"8", "NULL", "y"},
			{"12", "42.7", "bar"},
			{"12", "1.234", ""},
		},
	},
	{
		q: `
SELECT Ints, Floats, Strings
FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQuMixmb28KMTIsNDIuNyxiYXIKNywzLjE0MTUsemFwcGEKLDIuNzUseAo4LCx5CjEyLDEuMjM0LAo='
ORDER BY Floats;`,
		v: [][]string{
			{"8", "NULL", "y"},
			{"12", "1.234", ""},
			{"NULL", "2.75", "x"},
			{"7", "3.1415", "zappa"},
			{"1", "4.2", "foo"},
			{"12", "42.7", "bar"},
		},
	},
	{
		q: `
SELECT Ints, Floats, Strings
FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQuMixmb28KMTIsNDIuNyxiYXIKNywzLjE0MTUsemFwcGEKLDIuNzUseAo4LCx5CjEyLDEuMjM0LAo='
ORDER BY Strings;`,
		v: [][]string{
			{"12", "1.234", ""},
			{"12", "42.7", "bar"},
			{"1", "4.2", "foo"},
			{"NULL", "2.75", "x"},
			{"8", "NULL", "y"},
			{"7", "3.1415", "zappa"},
		},
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
       COUNT(Unit) AS Count,
       AVG(Count) AS Avg
FROM (
	  SELECT "0" AS Name,
	         "1" AS Unit,
	         "2" AS Count
	  FROM 'data:text/csv;base64,YSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
      FILTER 'noheaders'
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
       AVG(Count) AS Avg
FROM (
	  SELECT "0" AS Name,
	         "1" AS Unit,
	         "2" AS Count
	  FROM 'data:text/csv;base64,YSwxLDIwMAphLDIsMTAwCmEsMiw1MApiLDEsNTAKYiwyLDUwCmIsMywxMDAKYywxLDEwCmMsMSw3Cg=='
      FILTER 'noheaders'
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

	// Ints,Floats,Strings
	// 1,42.0,foo
	// 2,3.14,bar
	{
		q: `
SELECT Ints, Floats, Strings
FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQyLjAsZm9vCjIsMy4xNCxiYXIK'
ORDER BY Ints DESC;`,
		v: [][]string{
			{"2", "3.14", "bar"},
			{"1", "42", "foo"},
		},
	},
	{
		q: `
SELECT * FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQyLjAsZm9vCjIsMy4xNCxiYXIK';`,
		v: [][]string{
			{"1", "42", "foo"},
			{"2", "3.14", "bar"},
		},
	},
	// LIMIT tests:
	//
	// Ints,Floats,Strings
	// 1,4.2,foo
	// 12,42.7,bar
	// 7,3.1415,zappa
	// ,2.75,x
	// 8,,y
	// 12,1.234,
	{
		q: `
SELECT Ints
FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQuMixmb28KMTIsNDIuNyxiYXIKNywzLjE0MTUsemFwcGEKLDIuNzUseAo4LCx5CjEyLDEuMjM0LAo='
LIMIT 1;`,
		v: [][]string{
			{"1"},
		},
	},
	{
		q: `
SELECT Ints
FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQuMixmb28KMTIsNDIuNyxiYXIKNywzLjE0MTUsemFwcGEKLDIuNzUseAo4LCx5CjEyLDEuMjM0LAo='
LIMIT 1, 2;`,
		v: [][]string{
			{"12"},
			{"7"},
		},
	},
	{
		q: `
SELECT Ints
FROM 'data:text/csv;base64,SW50cyxGbG9hdHMsU3RyaW5ncwoxLDQuMixmb28KMTIsNDIuNyxiYXIKNywzLjE0MTUsemFwcGEKLDIuNzUseAo4LCx5CjEyLDEuMjM0LAo='
LIMIT 4, 100;`,
		v: [][]string{
			{"8"},
			{"12"},
		},
	},

	// Functions.
	{
		q: `
CREATE FUNCTION add(a INTEGER, b INTEGER)
RETURNS INTEGER
AS
BEGIN
    RETURN a + b;
END;

SELECT add(1, 2);
DROP FUNCTION add;`,
		v: [][]string{
			{"3"},
		},
	},
	{
		q: `
DROP FUNCTION IF EXISTS add;
CREATE FUNCTION add(a INTEGER, b INTEGER)
RETURNS INTEGER
BEGIN
    RETURN a + b;
END;

--SELECT add(1, 2);`,
		v: [][]string{
			{"3"},
		},
	},
}

func TestParser(t *testing.T) {
	for testID, input := range parserTests {
		name := fmt.Sprintf("Test %d", testID)
		global := NewScope(nil)
		parser := NewParser(global, bytes.NewReader([]byte(input.q)), name,
			os.Stdout)

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
				verifyResult(t, name, input.q, q, results[0])
				results = results[1:]
			}
		}
	}
}

func verifyResult(t *testing.T, name, source string, q types.Source,
	v [][]string) {
	rows, err := q.Get()
	if err != nil {
		t.Errorf("%s: q.Get failed: %v:\n%s\n", name, err, source)
		return
	}
	if len(rows) != len(v) {
		t.Errorf("%s: got %d rows, expected %d\n%s\n",
			name, len(rows), len(v), source)
		printResult(q, rows)
		return
	}
	for rowID, row := range rows {
		if len(row) != len(v[rowID]) {
			t.Fatalf("%s: row %d: got %d columns, expected %d\n%s\n",
				name, rowID, len(row), len(v[rowID]), source)
			printResult(q, rows)
			continue
		}
		for colID, col := range row {
			result := col.String()
			if result != v[rowID][colID] {
				t.Errorf("%s: %d.%d: got '%s', expected '%s'\n%s\n",
					name, rowID, colID, result, v[rowID][colID], source)
				printResult(q, rows)
			}
		}
	}
}

func printResult(q types.Source, rows []types.Row) {
	tab, err := types.Tabulate(q, tabulate.Unicode)
	if err != nil {
		return
	}
	tab.Print(os.Stdout)
}
