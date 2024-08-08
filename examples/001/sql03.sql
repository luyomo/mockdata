SELECT 
    security_id, market_type, comment, SUM(quantity), SUM(buy_quantity), SUM(buy_quantity_commission)
  , SUM(sell_quantity), SUM(sell_quantity_commission), SUM(total_quantity), SUM(total_commission), SUM(margin), currency, SUM(amount), order_date
FROM test.mocktable001
WHERE 
id = 4
AND order_date > '2024-07-21'
AND order_date < '2024-07-24'
GROUP BY 
security_id, market_type, comment, currency, order_date;

  | Num of Rows | Time Cost | Result  |
  | ----------- | --------- | ------  |
  | 1600000     | 0.056     |  31802  |
  | 12800000    | 0.440     |  255685 |
  | 32000000    | 1.933     |  640853 | 
  | 160000000   | 10.602    | 3197321 |
  | 400000000   | 44.243    | 7998318 |

* TiFlash
  | Num of Rows | Time Cost | Result  |
  | ----------- | --------- | ------- |
  | 160000000   | 1.823     | 3200149 |
  | 328000000   | 3.389     | 6563002 |


set @@tidb_mem_quota_query=4294967296;

ERROR 8175 (HY000): Your query has been cancelled due to exceeding the allowed memory limit for a single SQL query. Please try narrowing your query scope or increase the tidb_mem_quota_query limit and try again.[conn=3592441692]


// mockdata --threads 16 --loop 8 --config /tmp/mocktable001.yaml --output /tmp/mockdata --file-name=test.mocktable001 --rows 100000  --host=10.0.128.179 --user=root --pd-ip=10.0.233.4  --lightning-ver v8.0.0
// mockdata --threads 16 --loop 10 --config /tmp/mocktable001.yaml --output /tmp/mockdata --file-name=test.mocktable001 --rows 200000  --host=10.0.128.179 --user=root --pd-ip=10.0.233.4  --lightning-ver v8.0.0


//  mockdata --threads 16 --loop 50 --config /tmp/data.config.yaml --output /tmp/mockdata --file-name=test.mocktable001 --rows 200000  --host=10.0.47.29 --user=root --pd-ip=10.0.141.249  --lightning-ver v8.0.0


  | Number of rows | result
  | 800000000      | 1.181s
