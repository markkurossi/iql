# IQL - Internet Query Language

IQL is an SQL-inspired query language for processing Internet
resources. The IQL uses common data formats as input tables and allows
users to run SQL-like queries over the tables. The currently supported
data formats are comma-separated values (CSV) and HTML. The data
sources can be retrieved from HTTP and HTTPS URLs, local files, and
data URIs.

# Examples

The [examples](examples/) directory contains sample data files and
queries. The data files are also hosted at my [web
site](https://markkurossi.com/iql/examples/) and we use that location
for these examples.

The [store.html](https://markkurossi.com/iql/examples/store.html) file
contains 2 data sources, encoded has HTML tables. The "customers"
table contain information about store customers:


```sql
SELECT customers.'.id'      AS ID,
       customers.'.name'    AS Name,
       customers.'.address' AS Address
FROM 'https://markkurossi.com/iql/examples/store.html'
     FILTER 'table:nth-of-type(1) tr' AS customers
WHERE ID <> null;
```

```
┏━━━━┳━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ ID ┃ Name             ┃ Address                                           ┃
┡━━━━╇━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┩
│  1 │ Alyssa P. Hacker │ 77 Massachusetts Ave Cambridge, MA 02139          │
│  2 │ Ben Bitdiddle    │ 2200 Mission College Blvd. Santa Clara, CA  95052 │
│  3 │ Cy D. Fect       │ 208 S. Akard St. Dallas, TX 75202                 │
│  4 │ Eva Lu Ator      │ 353 Jane Stanford Way Stanford, CA 94305          │
│  5 │ Lem E. Tweakit   │ 1 Hacker Way Menlo Park, CA 94025                 │
│  6 │ Louis Reasoner   │ Princeton NJ 08544, United States                 │
└────┴──────────────────┴───────────────────────────────────────────────────┘
```

The "products" table defines the store products:

```sql
SELECT products.'.id'    AS ID,
       products.'.name'  AS Name,
       products.'.price' AS Price
FROM 'https://markkurossi.com/iql/examples/store.html'
     FILTER 'table:nth-of-type(2) tr' AS products
WHERE products.ID <> null;
```

```
┏━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━┓
┃ ID ┃ Name                                              ┃ Price ┃
┡━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━┩
│  1 │ Structure and Interpretation of Computer Programs │ 14.95 │
│  2 │ GNU Emacs Manual, For Version 21, 15th Edition    │  9.95 │
│  3 │ ISO/IEC 9075-1:2016(en) SQL — Part 1 Framework    │       │
└────┴───────────────────────────────────────────────────┴───────┘
```

The [orders.csv](https://markkurossi.com/iql/examples/orders.csv) file
contains order information, encoded as comma-separted values (CSV):

```sql
SELECT orders.'0' AS ID,
       orders.'1' AS Customer,
       orders.'2' AS Product,
       orders.'3' AS Count
FROM 'https://markkurossi.com/iql/examples/orders.csv' AS orders;
```

```
┏━━━━┳━━━━━━━━━━┳━━━━━━━━━┳━━━━━━━┓
┃ ID ┃ Customer ┃ Product ┃ Count ┃
┡━━━━╇━━━━━━━━━━╇━━━━━━━━━╇━━━━━━━┩
│  1 │        1 │       2 │     1 │
│  2 │        4 │       1 │     2 │
│  3 │        5 │       3 │     5 │
└────┴──────────┴─────────┴───────┘
```

In addition of listing individual tables, you can join tables and
compute values over the columns:

```sql
DECLARE storeurl VARCHAR;
SET storeurl = 'https://markkurossi.com/iql/examples/store.html';

DECLARE ordersurl VARCHAR;
SET ordersurl = 'https://markkurossi.com/iql/examples/orders.csv';

SELECT customers.Name                AS Name,
       customers.Address             AS Address,
       products.Name                 AS Product,
       orders.Count                  AS Count,
       products.Price * orders.Count AS Price
FROM (
        SELECT c.'.id'      AS ID,
               c.'.name'    AS Name,
               c.'.address' AS Address
        FROM storeurl FILTER 'table:nth-of-type(1) tr' AS c
        WHERE ID <> null
     ) AS customers,
     (
        SELECT p.'.id'    AS ID,
               p.'.name'  AS Name,
               p.'.price' AS Price
        FROM storeurl FILTER 'table:nth-of-type(2) tr' AS p
        WHERE ID <> null
     ) AS products,
     (
       SELECT o.'0' AS ID,
       	      o.'1' AS Customer,
       	      o.'2' AS Product,
       	      o.'3' AS Count
       FROM ordersurl AS o
     ) AS orders
WHERE orders.Product = products.ID AND orders.Customer = customers.ID;
```

```
┏━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━┳━━━━━━━┓
┃ Name             ┃ Address                                  ┃ Product                                           ┃ Count ┃ Price ┃
┡━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━╇━━━━━━━┩
│ Alyssa P. Hacker │ 77 Massachusetts Ave Cambridge, MA 02139 │ GNU Emacs Manual, For Version 21, 15th Edition    │     1 │  9.95 │
│ Eva Lu Ator      │ 353 Jane Stanford Way Stanford, CA 94305 │ Structure and Interpretation of Computer Programs │     2 │  29.9 │
│ Lem E. Tweakit   │ 1 Hacker Way Menlo Park, CA 94025        │ ISO/IEC 9075-1:2016(en) SQL — Part 1 Framework    │     5 │       │
└──────────────────┴──────────────────────────────────────────┴───────────────────────────────────────────────────┴───────┴───────┘
```

# Query Language Documentation

The IQL follows SQL in all constructs where possible. The full syntax
is defined in the [iql.iso-ebnf](iql.iso-ebnf) file and it is also
available as [SVG](iql.svg) and [HTML](iql.html) versions.

![IQL Grammar](iql.svg)

## Data Sources

### HTML

The HTML data source extracts input from HTML documents. The data
source uses the [goquery](https://github.com/PuerkitoBio/goquery)
package for the HTML processing. This means that the filter and column
selectors are CSS selectors, implemented by the
[cascadia](https://github.com/andybalholm/cascadia) library. The input
document processing is done as follows:
 - the `FILTER` selector selects input rows
 - the `SELECT` selectors select columns from input rows

### CSV

The CSV data source extracts input from comma-separated values (CSV)
data.  The data source uses Go's CSV encoding package for decoding the
data. The `FILTER` parameter can be used to specify CSV processing
options:
 - `skip`=*count*: skip the first *count* input lines
 - `comma`=*rune*: use *rune* to separate columns
 - `comment`=*rune*: skip lines starting with *rune*
 - `trim-leading-space`: trim leading space from columns
 - `headers`: the first line of the CSV data is a header line and its
   names are used to map select columns into CSV record columns

For example, if you input file is as follows:

```csv
Year; Value; Delta
# lines beginning with # character are ignored
1970; 100;   0
1971; 101;   1
1972; 200;   99
```

The fields can be processing with the following IQL code:

```sql
SELECT data.'0' AS Year,
       data.'1' AS Value,
       data.'2' AS Delta
FROM 'test_options.csv'
     FILTER 'skip=1 comma=; comment=# trim-leading-space'
     AS data;
```

```
┏━━━━━━┳━━━━━━━┳━━━━━━━┓
┃ Year ┃ Value ┃ Delta ┃
┡━━━━━━╇━━━━━━━╇━━━━━━━┩
│ 1970 │   100 │     0 │
│ 1971 │   101 │     1 │
│ 1972 │   200 │    99 │
└──────┴───────┴───────┘
```

Since our sample CSV file did have a header row, we can also use it to
name the data columns:

```sql
SELECT Year, Value, Delta
FROM 'test_options.csv'
     FILTER 'headers comma=; comment=# trim-leading-space';
```

This query gives the same result as the previous example:

```
┏━━━━━━┳━━━━━━━┳━━━━━━━┓
┃ Year ┃ Value ┃ Delta ┃
┡━━━━━━╇━━━━━━━╇━━━━━━━┩
│ 1970 │   100 │     0 │
│ 1971 │   101 │     1 │
│ 1972 │   200 │    99 │
└──────┴───────┴───────┘
```

## System Variables

 - REALFMT VARCHAR: specifies the formatting option for real
   numbers. The default value is `%g`.

## Built-in Functions

### Aggregate Functions

 - AVG(*expression*): returns the average value of all the values. The
   NULL values are ignored.
 - COUNT(*expression*): returns the count of all the values. The NULL
   values are ignored
 - MAX(*expression*): returns the maximum value of all the values. The
   NULL values are ignored.
 - MIN(*expression*): returns the minimum value of all the values. The
   NULL values are ignored.
 - NULLIF(*expr*, *value*): returns NULL if the *expr* and *value* are
   equal and the value of *expr* otherwise.
 - SUM(Expression): returns the sum of all the values. The NULL values
   are ignored.

### String Functions

 - BASE64ENC(*expression*): return the Base64 encoding of the string
   *expression*
 - BASE64DEC(*expression*): decodes the Base64 encoded string and
   returns the resulting data, converted to string
 - LEFT(*expression*, *count*): returns the *count* leftmost
   characters from the string *expression*.
 - LEN(*expression*): returns the number of Unicode code points in the
   string representation of *expression*.
 - LOWER(*expression*): returns the lowercase representation of the
   *expression*.
 - LTRIM(*expression*): remove the leading whitespace from the
   string representation of *expression*.
 - NCHAR(*expression*): returns the Unicode character with the integer
   code *expression*
 - TRIM(*expression*): remove the leading and trailing whitespace
   from the string representation of *expression*.
 - RTRIM(*expression*): remove the trailing whitespace from the
   string representation of *expression*.
 - UNICODE(*expression*): returns the integer value of the first
   Unicode character of the string *expression*
 - UPPER(*expression*): returns the uppercase representation of the
   *expression*.

# TODO

 - [ ] Queries:
   - [ ] Push table specific AND-relation SELECT expressions down to
         data source.
 - [ ] Aggregate:
   - [ ] Value cache
 - [ ] HTTP resource cache
 - [ ] JSON data format
 - [ ] YAML data format
