-- Attach source partitions ONCE on a single replica, no ON CLUSTER. NOT retry-safe.
-- Review and execute only when the previous stage succeeded.

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2011' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2012' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2013' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2014' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2015' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2016' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2017' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2018' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2019' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2020' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2021' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2022' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2023' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2024' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2025' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks` ATTACH PARTITION ID '2026' FROM `corp_portal_import`.`b_tasks`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2013' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2014' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2015' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2016' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2017' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2018' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2019' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2020' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2021' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2022' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2023' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2024' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2025' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_elapsed_time` ATTACH PARTITION ID '2026' FROM `corp_portal_import`.`b_tasks_elapsed_time`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2013' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2014' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2015' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2016' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2017' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2018' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2019' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2020' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2021' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2022' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2023' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2024' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2025' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_tasks_log` ATTACH PARTITION ID '2026' FROM `corp_portal_import`.`b_tasks_log`;

ALTER TABLE `corp_portal`.`b_user` ATTACH PARTITION ID 'all' FROM `corp_portal_import`.`b_user`;

ALTER TABLE `corp_portal`.`employees` ATTACH PARTITION ID 'all' FROM `corp_portal_import`.`employees`;

ALTER TABLE `corp_portal`.`projects` ATTACH PARTITION ID 'all' FROM `corp_portal_import`.`projects`;

ALTER TABLE `corp_portal`.`projets_mapping` ATTACH PARTITION ID 'all' FROM `corp_portal_import`.`projets_mapping`;

ALTER TABLE `corp_portal`.`projets_mapping_old` ATTACH PARTITION ID 'all' FROM `corp_portal_import`.`projets_mapping_old`;

ALTER TABLE `corp_portal`.`time_corp_portal` ATTACH PARTITION ID 'all' FROM `corp_portal_import`.`time_corp_portal`;

ALTER TABLE `corp_portal`.`time_jira_portal` ATTACH PARTITION ID 'all' FROM `corp_portal_import`.`time_jira_portal`;

