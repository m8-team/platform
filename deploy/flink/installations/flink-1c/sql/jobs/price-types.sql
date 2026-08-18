INSERT INTO price_types_sink
SELECT
  id,
  version,
  marked,
  predefined_id,
  parent_id,
  is_folder,
  code,
  name,
  currency_id,
  base_price_type_id,
  calculated,
  markup_discount_percent,
  vat_included,
  rounding_rule_id,
  round_up,
  `comment`,
  department_id,
  inactive,
  price_group,
  property_gender_id
FROM price_types_normalized;
