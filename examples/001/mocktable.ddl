
create table mocktable002(
  id bigint not null, 
  order_date date not null, 
  market_id bigint not null, 
  security_id bigint not null, 
  market_type varchar(64) not null,
  comment varchar(128) not null,
  quantity bigint,
  buy_quantity bigint, 
  buy_quantity_commission double,
  sell_quantity bigint,
  sell_quantity_commission double,
  total_quantity bigint, 
  total_commission double, 
  margin bigint, 
  currency varchar(32) not null,
  amount double,
  primary key(id, order_date, market_id, security_id, market_type, comment, currency)
);

