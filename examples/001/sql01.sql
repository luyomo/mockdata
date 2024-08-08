SELECT 
    market_id, market_type, comment, SUM(quantity), SUM(buy_quantity), SUM(buy_quantity_commission)
  , SUM(sell_quantity), SUM(sell_quantity_commission), SUM(total_quantity), SUM(total_commission), SUM(margin), currency, SUM(amount), order_date
FROM test.mocktable001
WHERE 
id = 4
AND order_date > '2024-07-21'
AND order_date < '2024-07-24'
AND market_id IN (2 ,3)
GROUP BY 
market_id, market_type, comment, currency, order_date;

id: 10
order_date: 10
market_id: 10
security_id: 10
  | Num of Rows | Time Cost | Result  |
  | ----------- | --------- | ------  |
  | 1600000     | 0.020     |         |
  | 12800000    | 0.239     |  151005 |
  | 32000000    | 0.583     |  128144 |
  | 160000000   | 1.220     |  639039 |
  | 400000000   | 6.682     | 1601734 |


  |328000000 | 0.923 | 1311892|




create index mocktable001_idx02 on mocktable001(id, market_id, security_id, market_type, order_date, comment, currency ,quantity, buy_quantity
	  , buy_quantity_commission, sell_quantity, sell_quantity_commission, total_quantity, total_commission, margin, amount );

 `id` bigint(20) NOT NULL,
  `order_date` date NOT NULL,
  `market_id` bigint(20) NOT NULL,
  `security_id` bigint(20) NOT NULL,
  `market_type` varchar(64) NOT NULL,
  `comment` varchar(128) NOT NULL,
  `quantity` bigint(20) DEFAULT NULL,
  `buy_quantity` bigint(20) DEFAULT NULL,
  `buy_quantity_commission` double DEFAULT NULL,
  `sell_quantity` bigint(20) DEFAULT NULL,
  `sell_quantity_commission` double DEFAULT NULL,
  `total_quantity` bigint(20) DEFAULT NULL,
  `total_commission` double DEFAULT NULL,
  `margin` bigint(20) DEFAULT NULL,
  `currency` varchar(32) NOT NULL,
  `amount` double DEFAULT NULL,

