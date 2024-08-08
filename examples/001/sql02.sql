SELECT 
    security_id, market_type, comment, SUM(quantity), SUM(buy_quantity), SUM(buy_quantity_commission)
  , SUM(sell_quantity), SUM(sell_quantity_commission), SUM(total_quantity), SUM(total_commission), SUM(margin), currency, SUM(amount), order_date
FROM test.mocktable001
WHERE 
id = 4
AND order_date > '2024-07-21'
AND order_date < '2024-07-24'
AND security_id IN ( 2, 3 )
GROUP BY 
security_id, market_type, comment, currency, order_date;

tikv
  | Num of Rows | Time Cost |  Result |
  | ----------- | --------- | ------- |
  | 1600000     | 0.050     |         |
  | 12800000    | 0.250     |   50929 |
  | 32000000    | 0.570     |  128358 |
  | 160000000   | 1.056     |  639319 |
  | 400000000   | 6.088     | 1599600 |

-----------

tiflash
  |
  | 328000000 | 0.797    | 1312728 |
