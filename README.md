# htmlq
HTML processor.

# Query language

```sql
SELECT ref.Name, ref.Price, ref.Weigth, portfolio.Weigth AS Portfolio
FROM
    (
        SELECT ".name"     AS Name,
               ".avgprice" AS Price,
               ".share"    AS Weight,
               ".link"     AS link
        FROM HTML ",reference.html" FILTER "tbody > tr"
        WHERE link <> ''
    ) AS ref,
    (
        SELECT 0 AS Name,
               1 AS Weigth
        FROM CSV ",portfolio.csv"
    ) AS portfolio,
WHERE ref.Name = portfolio.Name
```

```sql
SELECT ref.".name"         	      		AS Name,
       ref.":nth-child(5)" 	      		AS Price,
       ref.".share"    	   	      		AS Weigth,
       ref.a     	   	      		AS link,
       portfolio.0     	   	      		AS name
       portfolio.1     	   	      		AS Count
       Count * Price	   	      		AS Invested
       Count * Price / SUM(Count * Price) * 100 AS 'My Weight'
FROM HTML ",reference.html" FILTER "tbody > tr" AS ref,
     CSV  ",portfolio.csv" AS portfolio
WHERE ref.link <> '' AND ref.Name = portfolio.name
```
