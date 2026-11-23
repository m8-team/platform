package cataloggrpc

import (
	"errors"

	"github.com/m8-team/platform/internal/access/app/catalog"
	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/module"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/resource"
	"github.com/m8-team/platform/internal/access/domain/role"
	accesspb "github.com/m8-team/platform/internal/proto/m8/platform/access/v1"
)

var errInvalidManifest = errors.New("invalid catalog manifest")

func scopeFromProto(value accesspb.CatalogScope)(authorization.ScopeKind,error){
	switch value {
	case accesspb.CatalogScope_CATALOG_SCOPE_INSTANCE:return authorization.ScopeInstance,nil
	case accesspb.CatalogScope_CATALOG_SCOPE_COLLECTION:return authorization.ScopeCollection,nil
	case accesspb.CatalogScope_CATALOG_SCOPE_ORGANIZATION:return authorization.ScopeOrganization,nil
	case accesspb.CatalogScope_CATALOG_SCOPE_SUBTREE:return authorization.ScopeSubtree,nil
	case accesspb.CatalogScope_CATALOG_SCOPE_RESOURCE_KIND:return authorization.ScopeResourceKind,nil
	case accesspb.CatalogScope_CATALOG_SCOPE_TENANT:return authorization.ScopeTenant,nil
	default:return "",errInvalidManifest
	}
}
func scopeToProto(value authorization.ScopeKind)accesspb.CatalogScope{
	switch value {
	case authorization.ScopeInstance:return accesspb.CatalogScope_CATALOG_SCOPE_INSTANCE
	case authorization.ScopeCollection:return accesspb.CatalogScope_CATALOG_SCOPE_COLLECTION
	case authorization.ScopeOrganization:return accesspb.CatalogScope_CATALOG_SCOPE_ORGANIZATION
	case authorization.ScopeSubtree:return accesspb.CatalogScope_CATALOG_SCOPE_SUBTREE
	case authorization.ScopeResourceKind:return accesspb.CatalogScope_CATALOG_SCOPE_RESOURCE_KIND
	case authorization.ScopeTenant:return accesspb.CatalogScope_CATALOG_SCOPE_TENANT
	default:return accesspb.CatalogScope_CATALOG_SCOPE_UNSPECIFIED
	}
}
func parseScopes(input []accesspb.CatalogScope)([]authorization.ScopeKind,error){
	if len(input)==0{return nil,errInvalidManifest}
	out:=make([]authorization.ScopeKind,0,len(input))
	for _,value:=range input{
		scope,err:=scopeFromProto(value);if err!=nil{return nil,err}
		out=append(out,scope)
	}
	return out,nil
}

// fromProto translates only definition data. The caller's authenticated
// workload identity is injected by app/catalog.API, never by this mapper.
func fromProto(input *accesspb.CatalogModuleManifest)(catalog.PublishInput,error){
	if input==nil{return catalog.PublishInput{},errInvalidManifest}
	result:=catalog.PublishInput{ModuleID:input.GetModuleId(),Version:input.GetVersion()}
	for _,item:=range input.GetResources(){
		if item==nil{return catalog.PublishInput{},errInvalidManifest}
		scopes,err:=parseScopes(item.GetAllowedScopes());if err!=nil{return catalog.PublishInput{},err}
		kind,err:=resource.NewTypeDefinition(result.ModuleID,item.GetName(),scopes)
		if err!=nil{return catalog.PublishInput{},err}
		result.Resources=append(result.Resources,kind)
	}
	for _,item:=range input.GetPermissions(){
		if item==nil||item.GetGraphPermission()!="use"{return catalog.PublishInput{},errInvalidManifest}
		action,err:=permission.NewPermission(item.GetName());if err!=nil{return catalog.PublishInput{},err}
		definition,err:=permission.NewDefinition(action,item.GetResourceType(),item.GetGraphPermission())
		if err!=nil{return catalog.PublishInput{},err}
		result.Permissions=append(result.Permissions,definition)
	}
	for _,item:=range input.GetRoles(){
		if item==nil{return catalog.PublishInput{},errInvalidManifest}
		scopes,err:=parseScopes(item.GetAllowedScopes());if err!=nil{return catalog.PublishInput{},err}
		var actions []permission.Permission
		for _,name:=range item.GetPermissions(){
			p,err:=permission.NewPermission(name);if err!=nil{return catalog.PublishInput{},err}
			actions=append(actions,p)
		}
		ref:=role.TemplateRef{Name:item.GetName(),Version:item.GetVersion()}
		template,err:=role.NewTemplate(result.ModuleID,ref,actions,scopes)
		if err!=nil{return catalog.PublishInput{},err}
		result.Roles=append(result.Roles,template)
	}
	return result,nil
}

func toProtoManifest(manifest module.Manifest)*accesspb.CatalogModuleManifest{
	out:=&accesspb.CatalogModuleManifest{ModuleId:manifest.ModuleID(),Version:manifest.Version()}
	for _,item:=range manifest.Resources(){
		kind:=&accesspb.CatalogResourceType{Name:item.Name()}
		for _,scope:=range item.Scopes(){kind.AllowedScopes=append(kind.AllowedScopes,scopeToProto(scope))}
		out.Resources=append(out.Resources,kind)
	}
	for _,item:=range manifest.Permissions(){
		out.Permissions=append(out.Permissions,&accesspb.CatalogPermissionDefinition{
			Name:item.Permission().Name(),ResourceType:item.ResourceType(),GraphPermission:item.GraphPermission(),
		})
	}
	for _,item:=range manifest.Roles(){
		roleItem:=&accesspb.CatalogRoleTemplate{Name:item.Ref().Name,Version:item.Ref().Version}
		for _,action:=range item.Permissions(){roleItem.Permissions=append(roleItem.Permissions,action.Name())}
		for _,scope:=range item.AllowedScopes(){roleItem.AllowedScopes=append(roleItem.AllowedScopes,scopeToProto(scope))}
		out.Roles=append(out.Roles,roleItem)
	}
	return out
}
func toProtoRecord(record catalog.PublishedRecord)*accesspb.CatalogManifestRecord{
	return &accesspb.CatalogManifestRecord{
		Manifest:toProtoManifest(record.Manifest),
		PublisherSubjectId:record.Manifest.Publisher().ID(),
		Revision:record.Revision,
	}
}
