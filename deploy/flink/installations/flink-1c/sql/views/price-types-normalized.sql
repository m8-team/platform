CREATE VIEW price_types_normalized AS
SELECT
  `_IDRRef` AS id,
  `_Version` AS version,
  CASE `_Marked`
    WHEN 'AA==' THEN FALSE
    WHEN 'AQ==' THEN TRUE
    ELSE CAST(NULL AS BOOLEAN)
  END AS marked,
  `_PredefinedID` AS predefined_id,
  `_ParentIDRRef` AS parent_id,
  CASE `_Folder`
    WHEN 'AA==' THEN FALSE
    WHEN 'AQ==' THEN TRUE
    ELSE CAST(NULL AS BOOLEAN)
  END AS is_folder,
  `_Code` AS code,
  `_Description` AS name,
  `_Fld542RRef` AS currency_id,
  `_Fld543RRef` AS base_price_type_id,
  CASE `_Fld544`
    WHEN 'AA==' THEN FALSE
    WHEN 'AQ==' THEN TRUE
    ELSE CAST(NULL AS BOOLEAN)
  END AS calculated,
  `_Fld545` AS markup_discount_percent,
  CASE `_Fld546`
    WHEN 'AA==' THEN FALSE
    WHEN 'AQ==' THEN TRUE
    ELSE CAST(NULL AS BOOLEAN)
  END AS vat_included,
  `_Fld547RRef` AS rounding_rule_id,
  CASE `_Fld548`
    WHEN 'AA==' THEN FALSE
    WHEN 'AQ==' THEN TRUE
    ELSE CAST(NULL AS BOOLEAN)
  END AS round_up,
  `_Fld549` AS `comment`,
  `_Fld5445RRef` AS department_id,
  CASE `_Fld5676`
    WHEN 'AA==' THEN FALSE
    WHEN 'AQ==' THEN TRUE
    ELSE CAST(NULL AS BOOLEAN)
  END AS inactive,
  CASE `_Fld6571`
    WHEN 'AA==' THEN FALSE
    WHEN 'AQ==' THEN TRUE
    ELSE CAST(NULL AS BOOLEAN)
  END AS price_group,
  `_Fld8553RRef` AS property_gender_id
FROM price_types_source;
