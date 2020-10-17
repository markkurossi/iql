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
contains 3 data sources, encoded has HTML tables. The "customers"
table contain information about store customers:


```sql
SELECT customers.'.id'      AS ID,
       customers.'.name'    AS Name,
       customers.'.address' AS Address
FROM 'https://markkurossi.com/iql/examples/store.htmll'
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
│  3 │ ISO/IEC 9075-1:2016(en) SQL — Part 1 Framework    │  0.00 │
└────┴───────────────────────────────────────────────────┴───────┘
```

The "orders" table defines the orders:

```sql
SELECT orders.'.id'           AS ID,
       orders.':nth-child(2)' AS Customer,
       orders.':nth-child(3)' AS Product,
       orders.':nth-child(4)' AS Count
FROM 'https://markkurossi.com/iql/examples/store.html'
     FILTER 'table:nth-of-type(3) tr' AS orders
WHERE ID <> null;
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
SELECT customers.Name                AS Name,
       customers.Address             AS Address,
       products.Name                 AS Product,
       orders.Count                  AS Count,
       products.Price * orders.Count AS Price
FROM (
        SELECT c.'.id'      AS ID,
               c.'.name'    AS Name,
               c.'.address' AS Address
        FROM 'https://markkurossi.com/iql/examples/store.html'
	     FILTER 'table:nth-of-type(1) tr' AS c
        WHERE ID <> null
     ) AS customers,
     (
        SELECT p.'.id'    AS ID,
               p.'.name'  AS Name,
               p.'.price' AS Price
        FROM 'https://markkurossi.com/iql/examples/store.html'
	     FILTER 'table:nth-of-type(2) tr' AS p
        WHERE ID <> null
     ) AS products,
     (
        SELECT o.'.id'           AS ID,
               o.':nth-child(2)' AS Customer,
               o.':nth-child(3)' AS Product,
               o.':nth-child(4)' AS Count
        FROM 'https://markkurossi.com/iql/examples/store.html'
	     FILTER 'table:nth-of-type(3) tr' AS o
        WHERE ID <> null
     ) as orders
WHERE orders.Product = products.ID AND orders.Customer = customers.ID;
```

```
┏━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━┳━━━━━━━┓
┃ Name             ┃ Address                                  ┃ Product                                           ┃ Count ┃ Price ┃
┡━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━╇━━━━━━━┩
│ Alyssa P. Hacker │ 77 Massachusetts Ave Cambridge, MA 02139 │ GNU Emacs Manual, For Version 21, 15th Edition    │     1 │  9.95 │
│ Eva Lu Ator      │ 353 Jane Stanford Way Stanford, CA 94305 │ Structure and Interpretation of Computer Programs │     2 │ 29.90 │
│ Lem E. Tweakit   │ 1 Hacker Way Menlo Park, CA 94025        │ ISO/IEC 9075-1:2016(en) SQL — Part 1 Framework    │     5 │  0.00 │
└──────────────────┴──────────────────────────────────────────┴───────────────────────────────────────────────────┴───────┴───────┘
```

# TODO

 - [X] Statements are terminated by ';' so EOF is an error
 - [ ] Query language documentation
 - [ ] Aggregate:
   - [ ] Column selectors resulting single row
   - [ ] Value cache
 - [ ] Evaluate all queries from input stream
 - [ ] HTTP resource cache

# Query language

<img align="center" src="iql.svg">

```sql
SELECT ref.Name, ref.Price, ref.Weigth, portfolio.Weigth AS Portfolio
FROM
    (
        SELECT '.name'     AS Name,
               '.avgprice' AS Price,
               '.share'    AS Weight,
               '.link'     AS link
        FROM ',reference.html' FILTER 'tbody > tr'
        WHERE link <> ''
    ) AS ref,
    (
        SELECT '0' AS Name,
               '1' AS Weigth
        FROM ',portfolio.csv'
    ) AS portfolio,
WHERE ref.Name = portfolio.Name;
```

```sql
SELECT ind.'.name'         	      		AS Name,
       ind.':nth-child(5)' 	      		AS Price,
       ind.'.share'    	   	      		AS Weigth,
       ind.a     	   	      		AS link,
       portfolio.'0'   	   	      		AS name,
       portfolio.'1'   	   	      		AS Count,
       Count * Price	   	      		AS Invested,
       Count * Price / SUM(Count * Price) * 100 AS "My Weight"
FROM ',reference.html' FILTER 'tbody > tr' AS ind,
     ',portfolio.csv' AS portfolio
WHERE ind.link <> '' AND ind.Name = portfolio.name;
```
