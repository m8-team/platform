-- Attach source partitions ONCE on a single replica, no ON CLUSTER. NOT retry-safe.
-- Review and execute only when the previous stage succeeded.

ALTER TABLE `analytics`.`authentications` ATTACH PARTITION ID 'all' FROM `analytics_import`.`authentications`;

ALTER TABLE `analytics`.`authentications_channels` ATTACH PARTITION ID 'all' FROM `analytics_import`.`authentications_channels`;

ALTER TABLE `analytics`.`authentications_providers` ATTACH PARTITION ID 'all' FROM `analytics_import`.`authentications_providers`;

ALTER TABLE `analytics`.`bonus_logs` ATTACH PARTITION ID 'all' FROM `analytics_import`.`bonus_logs`;

ALTER TABLE `analytics`.`employees` ATTACH PARTITION ID 'all' FROM `analytics_import`.`employees`;

ALTER TABLE `analytics`.`employees_companies` ATTACH PARTITION ID 'all' FROM `analytics_import`.`employees_companies`;

ALTER TABLE `analytics`.`offers` ATTACH PARTITION ID 'all' FROM `analytics_import`.`offers`;

ALTER TABLE `analytics`.`prices` ATTACH PARTITION ID 'all' FROM `analytics_import`.`prices`;

ALTER TABLE `analytics`.`prices_costs` ATTACH PARTITION ID 'all' FROM `analytics_import`.`prices_costs`;

ALTER TABLE `analytics`.`prices_types` ATTACH PARTITION ID 'all' FROM `analytics_import`.`prices_types`;

ALTER TABLE `analytics`.`products` ATTACH PARTITION ID 'all' FROM `analytics_import`.`products`;

ALTER TABLE `analytics`.`stores` ATTACH PARTITION ID 'all' FROM `analytics_import`.`stores`;

ALTER TABLE `analytics`.`stores_traffic` ATTACH PARTITION ID 'all' FROM `analytics_import`.`stores_traffic`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '197001' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '201904' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202001' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202002' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202003' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202006' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202007' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202008' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202009' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202010' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202011' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202012' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202101' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202102' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202103' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202104' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202105' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202106' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202107' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202108' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202109' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202110' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202111' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202112' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202201' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202202' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202203' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202204' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202205' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202206' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202207' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202208' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202209' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202210' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202211' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202212' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202301' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202302' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202303' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202304' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202305' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202306' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202307' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202308' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202309' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202310' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202311' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202312' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202401' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202402' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202403' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202404' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202405' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202406' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202407' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202408' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202409' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202410' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202411' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202412' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202501' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202502' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202503' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202504' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202505' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202506' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202507' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202508' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202509' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202510' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202511' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202512' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202601' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202602' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202603' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202604' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202605' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202606' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202607' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202608' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202609' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202610' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_main` ATTACH PARTITION ID '202611' FROM `analytics_import`.`temp_onec_purchases_main`;

ALTER TABLE `analytics`.`temp_onec_purchases_payments` ATTACH PARTITION ID '197001' FROM `analytics_import`.`temp_onec_purchases_payments`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '197001' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202008' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202112' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202201' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202202' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202203' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202204' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202205' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202206' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202207' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202208' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202209' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202210' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202211' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202212' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202301' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202302' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202303' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202304' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202305' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202306' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202307' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202308' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202309' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202310' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202311' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202312' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202401' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202402' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202403' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202404' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202405' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202406' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202407' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202408' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202409' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202410' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202411' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202412' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202501' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202502' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202503' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202504' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202505' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202506' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202507' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202508' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202509' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202510' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202511' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202512' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202601' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202602' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202603' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202604' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202605' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202606' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202607' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202608' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202609' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202610' FROM `analytics_import`.`temp_onec_purchases_products`;

ALTER TABLE `analytics`.`temp_onec_purchases_products` ATTACH PARTITION ID '202611' FROM `analytics_import`.`temp_onec_purchases_products`;

