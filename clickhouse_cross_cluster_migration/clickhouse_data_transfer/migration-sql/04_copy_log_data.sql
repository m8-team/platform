-- Copy Log data ONCE. Repeating creates duplicate records.
-- Review and execute only when the previous stage succeeded.

INSERT INTO `zendendb`.`OZON_Orders_temp` SELECT * FROM `zendendb_import`.`OZON_Orders_temp`;

INSERT INTO `zendendb`.`YANDEX_warehouseIds` SELECT * FROM `zendendb_import`.`YANDEX_warehouseIds`;

INSERT INTO `zendendb`.`plan_2024` SELECT * FROM `zendendb_import`.`plan_2024`;

INSERT INTO `zendendb`.`prices_TM_MP_filter` SELECT * FROM `zendendb_import`.`prices_TM_MP_filter`;

INSERT INTO `zendendb`.`Склад_missing` SELECT * FROM `zendendb_import`.`Склад_missing`;

INSERT INTO `zendendb`.`СпособыДоставкиИМ` SELECT * FROM `zendendb_import`.`СпособыДоставкиИМ`;

INSERT INTO `zendendb`.`Управление_остатками` SELECT * FROM `zendendb_import`.`Управление_остатками`;

INSERT INTO `zendendb`.`Цены` SELECT * FROM `zendendb_import`.`Цены`;

