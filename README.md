# IQL - Internet Query Language

IQL is SQL-inspired query language for processing Internet resources.

# Query language

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
WHERE ref.Name = portfolio.Name
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
WHERE ind.link <> '' AND ind.Name = portfolio.name
```
