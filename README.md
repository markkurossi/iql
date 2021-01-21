# IQL - Internet Query Language

IQL is an SQL-inspired query language for processing Internet
resources. The IQL uses common data formats as input tables and allows
users to run SQL-like queries over the tables. The currently supported
data formats are comma-separated values (CSV), JavaScript Object
Notation (JSON), and HTML. The data sources can be retrieved from HTTP
and HTTPS URLs, local files, and data URIs.

# Examples

The [examples](examples/) directory contains sample data files and
queries. The data files are also hosted at my [web
site](https://markkurossi.com/iql/examples/) and we use that location
for these examples. Please, check also [Appendix
B](#appendix-b-data-visualization-examples) for additional data
visualization examples.

The [store.html](https://markkurossi.com/iql/examples/store.html) file
contains 2 data sources, encoded has HTML tables. The "customers"
table contain information about store customers:


```sql
SELECT customers.'.id'      AS ID,
       customers.'.name'    AS Name,
       customers.'.address' AS Address
FROM 'https://markkurossi.com/iql/examples/store.html'
     FILTER 'table:nth-of-type(1) tr' AS customers
WHERE '.id' <> null;
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
WHERE '.id' <> null;
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
contains order information, encoded as comma-separted values
(CSV). The data file does not have CSV headers at its first line so we
use the `noheaders` filter flag:

```sql
SELECT orders.'0' AS ID,
       orders.'1' AS Customer,
       orders.'2' AS Product,
       orders.'3' AS Count
FROM 'https://markkurossi.com/iql/examples/orders.csv'
FILTER 'noheaders' AS orders;
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
        WHERE '.id' <> null
     ) AS customers,
     (
        SELECT p.'.id'    AS ID,
               p.'.name'  AS Name,
               p.'.price' AS Price
        FROM storeurl FILTER 'table:nth-of-type(2) tr' AS p
        WHERE '.id' <> null
     ) AS products,
     (
       SELECT o.'0' AS ID,
       	      o.'1' AS Customer,
       	      o.'2' AS Product,
       	      o.'3' AS Count
       FROM ordersurl FILTER 'noheaders' AS o
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

The IQL follows SQL in all constructs where possible. The full
[syntax](#appendix-a-iql-grammar-bnf) is defined in the
[iql.iso-ebnf](iql.iso-ebnf) file and it is also available as
[SVG](iql.svg) and [HTML](iql.html) versions.

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
 - `noheaders`: the first line of the CSV data is not a header
   line. You must use column indices to select columns from the data.
 - `prepend-headers`=*header*[,...]: prepend the headers to the CSV
   file's header line. This option can be used to fix malformed CSV
   files which contain an invalid header line.

For example, if your input file is as follows:

```csv
Year; Value; Delta
# lines beginning with # character are ignored
1970; 100;   0
1971; 101;   1
1972; 200;   99
```

The fields can be processed with the following IQL code:

```sql
SELECT data.'0' AS Year,
       data.'1' AS Value,
       data.'2' AS Delta
FROM 'test-options.csv'
     FILTER 'noheaders skip=1 comma=; comment=# trim-leading-space'
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
FROM 'test-options.csv'
     FILTER 'comma=; comment=# trim-leading-space';
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

### JSON

The JSON data source extracts input from JSON documents. The data
source uses the [jsonq](https://github.com/markkurossi/jsonq) package
for the JSON processing. This means that the filter and column
selectors are JSONQ selectors which emulate XPath expressions. The
input document processing is done as follows:
 - the `FILTER` selector selects input rows
 - the `SELECT` selectors select columns from input rows

For example, if your input file is as follows:

```json
{
    "colors": [
	{
	    "name": "Black",
	    "red": 0,
	    "green": 0,
	    "blue": 0
	},
	{
	    "name": "Red",
	    "red": 205,
	    "green": 0,
	    "blue": 0
	},
	"... objects omitted ...",
	{
	    "name": "Bright White",
	    "red": 255,
	    "green": 255,
	    "blue": 255
	}
    ]
}
```

The color values can be processed with the following IQL code:

```sql
SELECT src.name         AS Name,
       src.red          AS Red,
       src.green        AS Green,
       src.blue         AS Blue
from 'ansi.json' FILTER 'colors' AS src;
```

```
┏━━━━━━━━━━━━━━━━━━━━━┳━━━━━┳━━━━━━━┳━━━━━━┓
┃ Name                ┃ Red ┃ Green ┃ Blue ┃
┡━━━━━━━━━━━━━━━━━━━━━╇━━━━━╇━━━━━━━╇━━━━━━┩
│ Black               │   0 │     0 │    0 │
│ Red                 │ 205 │     0 │    0 │
│ Green               │   0 │   205 │    0 │
│ Yellow              │ 205 │   205 │    0 │
│ Blue                │   0 │     0 │  238 │
│ Magenta             │ 205 │     0 │  205 │
│ Cyan                │   0 │   205 │  205 │
│ White               │ 229 │   229 │  229 │
│ Bright Black (Gray) │ 127 │   127 │  127 │
│ Bright Red          │ 255 │     0 │    0 │
│ Bright Green        │   0 │   255 │    0 │
│ Bright Yellow       │ 255 │   255 │    0 │
│ Bright Blue         │  92 │    92 │  255 │
│ Bright Magenta      │ 255 │     0 │  255 │
│ Bright Cyan         │   0 │   255 │  255 │
│ Bright White        │ 255 │   255 │  255 │
└─────────────────────┴─────┴───────┴──────┘
```

## System Variables

 |Variable|Type   |Default| Description |
 |--------|-------|-------|-------------|
 |REALFMT |VARCHAR|`%g`|The formatting option for real numbers.|
 |TABLEFMT|VARCHAR|`uc`|The table formatting style.|
 |TERMOUT |BOOLEAN|`ON`|Controls the terminal output from the queries.|

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

### Mathematical Functions

 - FLOOR(*numeric*): rounds the *numeric* value down to the largest
   integer less than or equal to the argument value.

### String Functions

 - BASE64DEC(*expression*): decodes the Base64 encoded string and
   returns the resulting data, converted to string
 - BASE64ENC(*expression*): return the Base64 encoding of the string
   *expression*
 - CHAR(*code*): returns the Unicode character for integer value
   *code*.
 - CHARINDEX(*expression*, *search* [, *start*]): return the first
   index of the substring *search* in *expression*. The optional
   argument *start* specifies the search start location. If the
   *start* argument is omitted or smaller than zero, the search start
   from the beginning of *expression*. **Note** that the returned
   index value is 1-based. The function returns the value 0 if the
   *search* substring could not be found from *expression*.
 - CONCAT(*val1*, *val2* [, ..., *valn*]): concatenates the argument
   string expressions into a string. All NULL expressions are handles
   as empty strings.
 - CONCAT_WS(*separator*, *val1*, *val2* [, ..., *valn*]):
   concatenates the argument string expressions into a string where
   arguments are separated by the *separator* string. All NULL
   expressions are ingored and they are not separated by the
   *separator* string. If the *separator* is NULL, this works like the
   CONCAT() function.
 - LASTCHARINDEX(*expression*, *search*): return the last index of the
   substring *search* in *expression*. **Note** that the returned
   index value is 1-based. The function returns the value 0 if the
   *search* substring could not be found from *expression*.
 - LEFT(*expression*, *count*): returns the *count* leftmost
   characters from the string *expression*.
 - LEN(*expression*): returns the number of Unicode code points in the
   string representation of *expression*.
 - LOWER(*expression*): returns the lowercase representation of the
   *expression*.
 - LPAD(*expression*, *length* [, *pad*]): pads the *expression* from
   the start with *pad* characters so that the resulting string has
   *lenght* characters. If the *expression* is longer than *length*,
   the function returns *length* leftmost characters from the string
   *expression*. If the argument *pad* is omitted, the space character
   (' ') is used as padding.
 - LTRIM(*expression*): remove the leading whitespace from the
   string representation of *expression*.
 - NCHAR(*expression*): returns the Unicode character with the integer
   code *expression*
 - REPLICATE(*expression*, *count*): repeats the string value
   *expression* count times. If the *count* is negative, the function
   returns NULL.
 - REVERSE(*expression*): return the reverse order of the argument
   string *expression*.
 - RIGHT(*expression*, *count*): returns the *count* rightmost
   characters from the string *expression*.
 - RTRIM(*expression*): remove the trailing whitespace from the
   string representation of *expression*.
 - SPACE(*count*): return a string containing *count* space
   characters.
 - STUFF(*string*, *start*, *length*, *replace*): remove *length*
   characters from the index *start* from the string expression
   *string* and replace the removed characters with *replace*. If
   *start* is smaller than or equal to 0, the function returns
   NULL. If the *start* is larger than the length of *string*, the
   function returns NULL. If *length* is negative, the function
   returns NULL. If *length* is larger than the length of *string*,
   the function removes all characters starting from the index
   *start*. If the *replace* values is NULL, no replacement characters
   are inserted.
 - SUBSTRING(*expression*, *start*, *length*): returns a substring of
   the *expression*. The *start* specifies the start index of the
   substring to return. **Note** that the start index is 1-based. If
   the start index is 0 or negative, the substring will start from the
   beginning of the *expression*. The *length* specifies non-negative
   length of the returned substring. If the *length* argument is
   negative, an error will be generated. If *start* + *length* is
   larger than the length of *expression*, the substring contains
   character to the end of *expression*.
 - TRIM(*expression*): remove the leading and trailing whitespace
   from the string representation of *expression*.
 - UNICODE(*expression*): returns the integer value of the first
   Unicode character of the string *expression*
 - UPPER(*expression*): returns the uppercase representation of the
   *expression*.

### Date and Time Functions

 - DATEDIFF(*diff*, *from*, *to*): returns the time difference between
   *from* and *to*. The *diff* specifies the units in which the
   difference is computed:
   - `year`, `yy`, `yyyy`: difference between date year parts
   - `day`, `dd`, `d`: difference in calendar days
   - `hour`, `hh`: difference in hours
   - `minute`, `mi`, `n`: difference in minutes
   - `second`, `ss`, `s`: difference seconds
   - `millisecond`, `ms`: difference in milliseconds
   - `microsecond`, `mcs`: difference in microseconds
   - `nanosecond`, `ns`: difference in nanoseconds
 - DAY(*date*): returns an integer representing the day of the month
   of the argument *date*
 - GETDATE(): returns the current system timestamp
 - MONTH(*date*): returns an integer representing the month of the
   year of the argument *date*
 - YEAR(*date*): returns an integer representing the year of the
   argument *date*.

### Data Visualization Functions

 - HBAR(*value*, *max*, *width* [,*pad*]): creates a horizontal
   histogram bar that is *width* characters long. The leftmost
   *value*/*max* characters are rendered with the Unicode Box Elements
   (U+2580-U+259F) and the remaining characters are filled with the
   *pad* character. The default padding character is space (' ',
   0x20). If the *pad* is a string, it must be one rune long.

# Appendix A: IQL Grammar BNF

![IQL Grammar](docs/iql.svg)

# Appendix B: Data Visualization Examples

These data visualization examples use a dataset from [Freie
Universität
Berlin](https://www.geo.fu-berlin.de/en/v/soga/Basics-of-statistics/Continous-Random-Variables/The-Standard-Normal-Distribution/The-Standard-Normal-Distribution-An-Example/index.html). The copyright of the dataset is as follows:

> You may use this project freely under the [Creative Commons Attribution-ShareAlike 4.0 International License](https://creativecommons.org/licenses/by-sa/4.0/). Please cite as follow: Hartmann, K., Krois, J., Waske, B. (2018): E-Learning Project SOGA: Statistics and Geospatial Data Analysis. Department of Earth Sciences, Freie Universitaet Berlin.

 The data set
[students.csv](https://userpage.fu-berlin.de/soga/200/2010_data_sets/students.csv)
contains 8239 records, each having 16 attributes of a particular
student. The CSV file has header row but it does not have the column
name for the first data column which is the record sequence
number. Therefore, the examples below use the `prepend-headers=seq`
CSV filtering option. All example snippets below are taken from the
[students.iql](examples/students.iql) IQL file.

All examples below assume that the following settings have been made
in the sample preamble:

```sql
-- Print real numbers two decimal digits.
SET REALFMT = '%.2f';

-- Define the students.csv data URL.
DECLARE dataurl VARCHAR;
SET dataurl = 'https://userpage.fu-berlin.de/soga/200/2010_data_sets/students.csv';
```

## Height Histogram

The height histogram shows student heights in 5cm data ranges. We use
the HBAR() function to create horizontal histogram bars.

```sql
SELECT height,
       count,
       HBAR(count, max(count), 20) AS histogram
FROM (
       SELECT height / 5 * 5 AS height,
              COUNT(height)  AS count
       FROM dataurl FILTER 'prepend-headers=seq'
       GROUP BY height / 5
       ORDER BY height
     );
```

```
┏━━━━━━━━┳━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━┓
┃ height ┃ count ┃ histogram            ┃
┡━━━━━━━━╇━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━┩
│    135 │     3 │                      │
│    140 │    32 │ ▍                    │
│    145 │   126 │ █▊                   │
│    150 │   352 │ █████▏               │
│    155 │   732 │ ██████████▊          │
│    160 │  1138 │ ████████████████▊    │
│    165 │  1230 │ ██████████████████▏  │
│    170 │  1350 │ ████████████████████ │
│    175 │  1196 │ █████████████████▋   │
│    180 │  1037 │ ███████████████▎     │
│    185 │   644 │ █████████▌           │
│    190 │   291 │ ████▎                │
│    195 │    87 │ █▎                   │
│    200 │    19 │ ▎                    │
│    205 │     2 │                      │
└────────┴───────┴──────────────────────┘
```

## Male Height Histogram

This example draws height histogram for male students.

```sql
SELECT height,
       count,
       HBAR(count, max(count), 20) AS histogram
FROM (
       SELECT height / 5 * 5 AS height,
              COUNT(height)  AS count,
              gender         AS ",gender"
       FROM dataurl FILTER 'prepend-headers=seq'
       WHERE gender='Male'
       GROUP BY height / 5
       ORDER BY height
     );
```

```
┏━━━━━━━━┳━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━┓
┃ height ┃ count ┃ histogram            ┃
┡━━━━━━━━╇━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━┩
│    140 │     1 │                      │
│    145 │     2 │                      │
│    150 │     3 │                      │
│    155 │    25 │ ▌                    │
│    160 │   117 │ ██▍                  │
│    165 │   322 │ ██████▋              │
│    170 │   711 │ ██████████████▋      │
│    175 │   954 │ ███████████████████▋ │
│    180 │   969 │ ████████████████████ │
│    185 │   630 │ █████████████        │
│    190 │   287 │ █████▉               │
│    195 │    87 │ █▊                   │
│    200 │    19 │ ▍                    │
│    205 │     2 │                      │
└────────┴───────┴──────────────────────┘
```

## Changing Histogram Background Color

We can use VT100 codes to change the histogram background color. This
helps us to visualize some histogram bars which have very small number
of samples.

First, we create a function for rendering the histogram bars:

```sql
CREATE FUNCTION histogram(val INTEGER, max INTEGER, width INTEGER)
RETURNS VARCHAR
AS
BEGIN
    RETURN CONCAT(CHAR(0x1b), '[107m',
                  HBAR(val, max, width),
                  CHAR(0x1b), '[0m');
END
```

Then we update the height histogram program to use the new
`histogram()` function to render the data bars:

```sql
SELECT height,
       count,
       histogram(count, max(count), 20) AS histogram
FROM (
       SELECT height / 5 * 5 AS height,
              COUNT(height)  AS count
       FROM dataurl FILTER 'prepend-headers=seq'
       GROUP BY height / 5
       ORDER BY height
     );
```

# Appendix C: TODO

 - [ ] Queries:
   - [ ] Push table specific AND-relation SELECT expressions down to
         data source.
 - [ ] Aggregate:
   - [ ] Value cache
 - [ ] HTTP resource cache
 - [ ] YAML data format
 - [ ] SQL Server base year for YEAR(0) is 1900
