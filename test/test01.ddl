CREATE TABLE `test01` (
  `id` bigint(20) NOT NULL,
  `payer` varchar(64) DEFAULT NULL,
  `receiver` varchar(64) DEFAULT NULL,
  `amount` bigint(20) DEFAULT NULL,
  `payment_uuid` varchar(64) DEFAULT NULL,
  `payment_type` varchar(32) DEFAULT NULL,
  `payment_date` date DEFAULT NULL,
  PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin
