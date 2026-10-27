import {loggedFetch} from '../../../platform/http/loggedFetch'

// Group is a non-owning business association inside exactly one Tenant.
// Group links are not legal parent/child relationships and grant no access.
export interface OrganizationGroup {
  id: string
  tenantId: string
  name: string
  description: string
  state: string
  revision: number
  createdAt?: string
  updatedAt?: string
}

export interface GroupLink {
  groupId: string
  organizationId: string
  createdAt?: string
}

const root = (import.meta.env.VITE_RESOURCE_MANAGER_API_URL ?? '').replace(/\/$/, '')
const prefix = '/resource-manager/v1'
const uuid = /^[a-f\d]{8}-[a-f\d]{4}-[1-8][a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$/i

function object(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

function string(value: unknown) { return typeof value === 'string' ? value : '' }

function parseGroup(raw: unknown, tenantId: string): OrganizationGroup {
  if (!object(raw) || !object(raw.group) || !object(raw.group.metadata)) {
    throw new Error('Invalid Group resource')
  }
  const group = raw.group
  const metadata = group.metadata as Record<string, unknown>
  const id = string(metadata.id)
  if (!uuid.test(id) || string(group.tenantId) !== tenantId) {
    throw new Error('Group resource is outside the selected Tenant or has an invalid ID')
  }
  return {
    id, tenantId, name: string(metadata.name) || id,
    description: string(metadata.description),
    state: string(metadata.state).replace(/^RESOURCE_STATE_/, '') || 'UNSPECIFIED',
    revision: typeof metadata.revision === 'number' ? metadata.revision : Number(metadata.revision) || 0,
    createdAt: string(metadata.createTime) || undefined,
    updatedAt: string(metadata.updateTime) || undefined,
  }
}

async function request(path: string, options: RequestInit = {}): Promise<unknown> {
  const response = await loggedFetch('resource-manager', root + prefix + path, {
    credentials: 'same-origin',
    headers: {Accept: 'application/json', ...(options.body ? {'Content-Type': 'application/json'} : {})},
    ...options,
  })
  if (!response.ok) {
    const reason = response.status === 401 || response.status === 403 ? 'Access denied'
      : response.status === 409 || response.status === 412 ? 'Resource conflict'
        : response.status === 404 ? 'Resource or endpoint not found'
          : response.status === 501 ? 'API operation is unavailable' : 'Resource Manager request failed'
    throw new Error(`${reason} (HTTP ${response.status})`)
  }
  if (response.status === 204) return {}
  const responseText = await response.text()
  return responseText ? JSON.parse(responseText) as unknown : {}
}

export async function listGroups(tenantId: string, signal?: AbortSignal): Promise<OrganizationGroup[]> {
  if (!uuid.test(tenantId)) throw new Error('Select a valid Tenant UUID')
  // v1 ListResources explicitly permits group + tenantId. Never use unscoped kind=group.
  const query = new URLSearchParams({kind: 'group', tenantId})
  const raw = await request('/resources?' + query, {signal})
  if (!object(raw) || !Array.isArray(raw.resources)) throw new Error('Invalid Group list response')
  if (string(raw.nextPageToken)) throw new Error('Group paging is not yet supported by this view')
  const groups = raw.resources.map((record) => parseGroup(record, tenantId))
  const ids = new Set(groups.map((g) => g.id))
  if (ids.size !== groups.length) throw new Error('Duplicate groups in Resource Manager response')
  return groups.sort((a, b) => a.name.localeCompare(b.name))
}

export async function getGroup(groupId: string, tenantId: string, signal?: AbortSignal): Promise<OrganizationGroup> {
  if (!uuid.test(groupId)) throw new Error('Invalid Group UUID')
  return parseGroup(await request('/resources/group/' + encodeURIComponent(groupId), {signal}), tenantId)
}

export async function listGroupLinks(groupId: string, signal?: AbortSignal): Promise<GroupLink[]> {
  if (!uuid.test(groupId)) throw new Error('Invalid Group UUID')
  const raw = await request('/groups/' + encodeURIComponent(groupId) + '/organizations', {signal})
  if (!object(raw) || !Array.isArray(raw.links)) throw new Error('Invalid Group affiliations response')
  if (string(raw.nextPageToken)) throw new Error('Group affiliation paging is not yet supported by this view')
  const links = raw.links.map((link): GroupLink => {
    if (!object(link) || link.groupId !== groupId || !uuid.test(string(link.organizationId))) {
      throw new Error('Invalid Group affiliation')
    }
    return {groupId, organizationId: string(link.organizationId), createdAt: string(link.createTime) || undefined}
  })
  if (new Set(links.map((l) => l.organizationId)).size !== links.length) {
    throw new Error('Duplicate affiliations')
  }
  return links
}

export async function createGroup(tenantId: string, name: string, description: string, idempotencyKey: string) {
  const raw = await request('/resources', {
    method: 'POST', body: JSON.stringify({
      resource: {group: {tenantId, metadata: {name: name.trim(), description: description.trim()}}},
      idempotencyKey,
    }),
  })
  return parseGroup(raw, tenantId)
}

export async function addGroupOrganization(groupId: string, organizationId: string, idempotencyKey: string) {
  const raw = await request('/group-affiliations', {
    method: 'POST', body: JSON.stringify({groupId, organizationId, idempotencyKey}),
  })
  if (!object(raw) || raw.groupId !== groupId || raw.organizationId !== organizationId) {
    throw new Error('Invalid affiliation confirmation from Resource Manager')
  }
}

export async function removeGroupOrganization(groupId: string, organizationId: string, idempotencyKey: string) {
  const params = new URLSearchParams({idempotencyKey})
  await request('/groups/' + encodeURIComponent(groupId) + '/organizations/' +
    encodeURIComponent(organizationId) + '?' + params, {method: 'DELETE'})
}

export async function verifyOrganizationInTenant(organizationId: string, tenantId: string): Promise<void> {
  if (!uuid.test(organizationId)) throw new Error('Invalid Organization UUID')
  const raw = await request('/resources/organization/' + encodeURIComponent(organizationId))
  if (!object(raw) || !object(raw.organization) ||
      raw.organization.tenantId !== tenantId || !object(raw.organization.metadata) ||
      raw.organization.metadata.id !== organizationId) {
    throw new Error('Organization does not belong to the selected Tenant')
  }
}
