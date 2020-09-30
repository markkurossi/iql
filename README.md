# htmlq
HTML processor.

# Query language

```sql
select ".name" as Name,
       ".avgprice" as Avg align MR,
       ".share" as Share align MR,
       ".link" as link
  from "tbody > tr"
  where length(link) > 0
```
