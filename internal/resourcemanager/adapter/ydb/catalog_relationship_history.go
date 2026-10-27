package ydb

import (
	"context"
	"database/sql"

	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

func (c *Catalog) ListRelationshipVersions(ctx context.Context,tenant,id string) ([]v1engine.OrganizationRelationship,error) {
	if tenant=="" || id=="" {return nil,v1engine.ErrInvalid}
	current,err:=c.GetRelationship(ctx,tenant,id)
	if err!=nil{return nil,err}
	rows,err:=c.store.db.QueryContext(ctx,c.stmt(
		"DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;"+
		"SELECT revision,from_organization_id,to_organization_id,relationship_type,ownership_basis_points,"+
		"valid_from_unix,valid_to_unix,evidence_ref,state,updated_at "+
		"FROM "+relationshipHistoryTable+" WHERE tenant_id=$tenant AND relationship_id=$id ORDER BY revision ASC;"),
		sql.Named("tenant",tenant),sql.Named("id",id))
	if err!=nil{return nil,err}
	defer rows.Close()
	var versions []v1engine.OrganizationRelationship
	for rows.Next() {
		r:=current
		var typ string
		if err:=rows.Scan(&r.Revision,&r.FromOrganizationID,&r.ToOrganizationID,&typ,&r.OwnershipBasisPoints,
			&r.ValidFrom,&r.ValidTo,&r.EvidenceRef,&r.State,&r.UpdatedAt);err!=nil{return nil,err}
		r.Type=v1engine.RelationshipType(typ)
		versions=append(versions,r)
	}
	if err:=rows.Err();err!=nil{return nil,err}
	return versions,nil
}
