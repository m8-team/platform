-- Convert Log tables into ReplicatedMergeTree with ORDER BY tuple().
-- Review and execute only when the previous stage succeeded.

CREATE TABLE `zendendb`.`OZON_Orders_temp` (`page` UInt32, `date_start` Date, `date_end` Date, `timestamp` DateTime) ENGINE = ReplicatedMergeTree() ORDER BY tuple();

CREATE TABLE `zendendb`.`YANDEX_warehouseIds` (`warehouseId` UInt32, `warehouseName` String) ENGINE = ReplicatedMergeTree() ORDER BY tuple() COMMENT 'Каталог складов МП';

CREATE TABLE `zendendb`.`plan_2024` (`Период` String, `Оборот План` Float32, `Контрагент` String, `ТМ` String, `Вид Номенклатуры` String) ENGINE = ReplicatedMergeTree() ORDER BY tuple() COMMENT 'Архив планов продаж МП ТМ';

CREATE TABLE `zendendb`.`prices_TM_MP_filter` (`art` String) ENGINE = ReplicatedMergeTree() ORDER BY tuple() COMMENT 'Фильтры артикулов для отчета по ценам загрузка через https://console.yandex.cloud/folders/b1gk4um23pnt6oh52r89/storage/buckets/loading?key=price_filter_TM%2F';

CREATE TABLE `zendendb`.`Склад_missing` (`Период` DateTime, `НомерСтроки` Decimal(9, 0), `Регистратор_TRef` String, `Регистратор` String, `Активность` String, `ВидДвижения` Decimal(1, 0), `СкладГуид` String, `Склад` String, `НоменклатураГуид` String, `Артикул` String, `КодДляОбмена` Decimal(7, 0), `Номенклатура` String, `ХарактеристикаНоменклатуры` String, `Качество` String, `Количество` Decimal(15, 3)) ENGINE = ReplicatedMergeTree() ORDER BY tuple();

CREATE TABLE `zendendb`.`СпособыДоставкиИМ` (`DELIVERY_CODE` String, `DELIVERY_NAME` String) ENGINE = ReplicatedMergeTree() ORDER BY tuple();

CREATE TABLE `zendendb`.`Управление_остатками` (`Номер артикула` String, `Название товара` String, `Номер SKU` Int32, `Какие признаки есть у товара` String, `В какой зоне на складе разместили товар. Зависит от категории и габаритов` String, `Кластер, где размещены остатки товара` String, `Склад, где размещены остатки товара` String, `Сколько товаров доступно к продаже со складов Ozon` Int32, `Сколько товаров без брака и с достаточным сроком годности скоро разместим на складе Ozon и начнём продавать` Int32, `Сколько маркируемых остатков вы сняли с продажи. Создайте заявку на вывоз, чтобы забрать их со склада` Int32, `Сколько товаров не можем добавить в продажу из-за проблем с документами` Int32, `Сколько товаров с истекающим сроком годности. Создайте заявку на вывоз заранее, чтобы успеть забрать до утилизации` Int32, `Сколько товаров с дефектами обнаружили во время приёмки на складе Ozon. Создайте заявку на вывоз, чтобы забрать их, иначе утилизируем` Int32, `Сколько товаров нельзя уценить из-за сильных повреждений. Создайте заявку на вывоз, чтобы забрать их, иначе утилизируем` Int32, `Сколько товаров не смогли опознать во время приёмки. Создайте заявку на вывоз, чтобы забрать их, иначе утилизируем` Int32, `Сколько товаров отправили на проверку. Это могло произойти, например, после обнаружения брака, истёкшего срока годности, пересорта` Int32, `Сколько товаров в созданных, но ещё не отгруженных заявках на поставку` Int32, `Сколько товаров в поставках, которые сейчас доставляются на склад Ozon. Товары будут в этом статусе до окончания приёмки` Int32, `Количество товаров едет на склад Ozon после отмен и возвратов от покупателей` Int32, `Сколько товаров готовим к вывозу со склада Ozon по вашей заявке` Int32) ENGINE = ReplicatedMergeTree() ORDER BY tuple();

CREATE TABLE `zendendb`.`Цены` (`Период` DateTime, `ТипЦен` String, `ТипЦенГуид` String, `КодДляОбмена` Decimal(7, 0), `ТоварГуид` String, `Товар` String, `Цена` Decimal(15, 2)) ENGINE = ReplicatedMergeTree() ORDER BY tuple();

