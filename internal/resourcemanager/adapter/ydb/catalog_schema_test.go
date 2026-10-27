package ydb

import (
	"context"
	"database/sql"
	"fmt"
)

// prepareResourceCatalogSchema is restricted to disposable local YDB tests;
// this is deliberately separate from production migration execution.
func prepareResourceCatalogSchema(ctx context.Context, db *sql.DB, prefix string) error {
	statements := []string{
		"CREATE TABLE `m8-resource-manager/resource_records` (kind Utf8 NOT NULL,id Utf8 NOT NULL,tenant_id Utf8 NOT NULL,owner_kind Utf8 NOT NULL,owner_id Utf8 NOT NULL,organization_id Utf8 NOT NULL,display_name Utf8 NOT NULL,description Utf8 NOT NULL,state Utf8 NOT NULL,revision Int64 NOT NULL,environment_key Utf8 NOT NULL,created_at Timestamp NOT NULL,updated_at Timestamp NOT NULL,PRIMARY KEY(kind,id));",
		"CREATE TABLE `m8-resource-manager/resource_create_requests` (scope Utf8 NOT NULL,idempotency_key Utf8 NOT NULL,request_hash String NOT NULL,resource_kind Utf8 NOT NULL,resource_id Utf8 NOT NULL,created_at Timestamp NOT NULL,PRIMARY KEY(scope,idempotency_key));",
		"CREATE TABLE `m8-resource-manager/resource_environment_keys` (organization_id Utf8 NOT NULL,environment_key Utf8 NOT NULL,environment_id Utf8 NOT NULL,PRIMARY KEY(organization_id,environment_key));",
		"CREATE TABLE `m8-resource-manager/resource_group_affiliations` (group_id Utf8 NOT NULL,organization_id Utf8 NOT NULL,tenant_id Utf8 NOT NULL,created_at Timestamp NOT NULL,PRIMARY KEY(group_id,organization_id));",
		"CREATE TABLE `m8-resource-manager/resource_environment_bindings` (id Utf8 NOT NULL,resource_kind Utf8 NOT NULL,resource_id Utf8 NOT NULL,organization_id Utf8 NOT NULL,environment_id Utf8 NOT NULL,version_ref Utf8 NOT NULL,configuration_ref Utf8 NOT NULL,revision Int64 NOT NULL,created_at Timestamp NOT NULL,updated_at Timestamp NOT NULL,PRIMARY KEY(id));",
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s", prefix, statement)); err != nil {
			return fmt.Errorf("resource catalog migration: %w", err)
		}
	}
	return nil
}
