#!/bin/bash

for num in $(seq 1 400)
do
#     echo "Num: $(($num % 5))"
    mysql -h 4.246.254.131 -u root -P 4000 test -e "explain analyze SELECT security_id, market_type, comment, SUM(quantity), SUM(buy_quantity), SUM(buy_quantity_commission) , SUM(sell_quantity), SUM(sell_quantity_commission), SUM(total_quantity), SUM(total_commission), SUM(margin), currency, SUM(amount), order_date FROM test.mocktable001 WHERE id = $(($num % 10)) AND order_date > '2024-07-21' AND order_date < '2024-07-24' GROUP BY security_id, market_type, comment, currency, order_date" &

done
