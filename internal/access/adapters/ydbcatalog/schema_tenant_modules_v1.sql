-- Access Universal tenant-module installation read model, additive v1.
-- This is NOT a public upsert API. Only a reviewed tenant module lifecycle
-- writer may create or change records, with independent authorization and CAS.
CREATE TABLE `m8-access/catalog_tenant_modules`
(
  tenant_id Utf8 NOT NULL,
  module_id Utf8 NOT NULL,
  release_version Utf8 NOT NULL,
  enabled Bool NOT NULL,
  revision Uint64 NOT NULL,
  PRIMARY KEY (tenant_id, module_id)
);
