import {loggedFetch} from '../../../platform/http/loggedFetch'

const apiBase = (import.meta.env.VITE_RESOURCE_MANAGER_API_URL ?? '').replace(/\/$/, '')

type RecordValue = Record<string, unknown>

function isRecord(value: unknown): value is RecordValue {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function requiredString(value: unknown, name: string): string {
  if (typeof value !== 'string' || !value) throw new Error(`Invalid ${name} from Resource Manager`)
  return value
}

function optionalString(value: unknown): string | undefined {
  return typeof value === 'string' && value.length > 0 ? value : undefined
}

function metadata(value: unknown): RecordValue {
  if (!isRecord(value)) throw new Error('Resource Manager returned missing resource metadata')
  return value
}

async function loadJson(path: string, signal?: AbortSignal): Promise<unknown> {
  const result = await loggedFetch('resource-manager', `${apiBase}${path}`, {
    signal,
    credentials: 'same-origin',
    headers: {Accept: 'application/json'},
  })
  if (!result.ok) throw new Error(`Resource Manager returned HTTP ${result.status}`)
  return result.json()
}

export interface OrganizationDetail {
  id: string
  tenantId: string
  name: string
  description?: string
  state: string
  revision?: number
  createdAt?: string
  updatedAt?: string
  companyProfileRef?: string
}

export async function fetchOrganizationDetail(id: string, signal?: AbortSignal): Promise<OrganizationDetail> {
  const value = await loadJson(`/resource-manager/v1/resources/organization/${encodeURIComponent(id)}`, signal)
  if (!isRecord(value) || !isRecord(value.organization)) {
    throw new Error('Resource Manager returned an invalid Organization resource')
  }
  const organization = value.organization
  const meta = metadata(organization.metadata)
  const responseId = requiredString(meta.id, 'organization ID')
  if (responseId !== id) throw new Error('Resource Manager returned a different Organization')
  return {
    id: responseId,
    name: requiredString(meta.name, 'organization name'),
    tenantId: requiredString(organization.tenantId, 'tenant ID'),
    description: optionalString(meta.description),
    state: optionalString(meta.state)?.replace(/^RESOURCE_STATE_/, '') ?? 'UNSPECIFIED',
    revision: typeof meta.revision === 'number' ? meta.revision : undefined,
    createdAt: optionalString(meta.createTime),
    updatedAt: optionalString(meta.updateTime),
    companyProfileRef: optionalString(organization.companyProfileRef),
  }
}

export interface OrgUnit {
  id: string
  organizationId: string
  parentId: string
  name: string
  code?: string
  kind?: string
  state: string
  managementMode: string
  authoritativeSourceId?: string
  structureRevision: number
}

export interface OrganizationStructure {
  revision: number
  units: OrgUnit[]
}

function parseOrgUnit(value: unknown, organizationId: string, revision: number): OrgUnit {
  if (!isRecord(value)) throw new Error('Invalid OrgUnit resource')
  const meta = metadata(value.metadata)
  if (requiredString(value.organizationId, 'OrgUnit organization ID') !== organizationId) {
    throw new Error('OrgUnit belongs to another organization')
  }
  return {
    id: requiredString(meta.id, 'OrgUnit ID'),
    organizationId,
    parentId: optionalString(value.parentOrgUnitId) ?? '',
    name: requiredString(meta.name, 'OrgUnit name'),
    code: optionalString(value.code),
    kind: optionalString(value.kind),
    state: optionalString(value.lifecycleState)?.replace(/^ORG_UNIT_STATE_/, '') ?? 'UNSPECIFIED',
    managementMode:
      optionalString(value.managementMode)?.replace(/^ORG_UNIT_MANAGEMENT_MODE_/, '') ?? 'UNSPECIFIED',
    authoritativeSourceId: optionalString(value.authoritativeSourceId),
    structureRevision:
      typeof value.structureRevision === 'number' ? value.structureRevision : revision,
  }
}

// The server offers revision-bound cursor pages, not a single full-tree RPC.
// Stop on broken cursors or a revision change rather than render a mixed tree.
export async function fetchOrganizationStructure(
  organizationId: string,
  signal?: AbortSignal,
): Promise<OrganizationStructure> {
  const units: OrgUnit[] = []
  const seenTokens = new Set<string>()
  const seenUnits = new Set<string>()
  let pageToken = ''
  let revision: number | undefined

  for (let page = 0; page < 100; page += 1) {
    const parameters = new URLSearchParams({pageSize: '200'})
    if (pageToken) parameters.set('pageToken', pageToken)
    const result = await loadJson(
      `/resource-manager/v1/organizations/${encodeURIComponent(organizationId)}/org-units?${parameters}`,
      signal,
    )
    if (!isRecord(result) || !Array.isArray(result.orgUnits)) {
      throw new Error('Invalid OrgUnit list response')
    }
    if (typeof result.structureRevision !== 'number' || result.structureRevision < 1) {
      throw new Error('OrgUnit revision is missing')
    }
    if (revision !== undefined && revision !== result.structureRevision) {
      throw new Error('Organization structure changed during pagination. Refresh to reload.')
    }
    revision = result.structureRevision
    for (const rawUnit of result.orgUnits) {
      const unit = parseOrgUnit(rawUnit, organizationId, revision)
      if (seenUnits.has(unit.id)) throw new Error('Duplicate OrgUnit in API response')
      if (unit.structureRevision !== revision) throw new Error('OrgUnit revision mismatch')
      seenUnits.add(unit.id)
      units.push(unit)
    }
    const next = optionalString(result.nextPageToken)
    if (!next) return {revision, units}
    if (seenTokens.has(next)) throw new Error('Repeated OrgUnit page token')
    seenTokens.add(next)
    pageToken = next
  }
  throw new Error('OrgUnit hierarchy is too large for the initial UI. Use a scoped subtree view.')
}

export interface OrganizationResource {
  id: string
  kind: string
  name: string
  state: string
  revision?: number
}

// This endpoint currently returns an unpaginated organization-scoped list.
// Do not imply that ownership, resources outside the scope or access grants
// are included in this result.
export async function fetchOrganizationResources(
  organizationId: string,
  signal?: AbortSignal,
): Promise<OrganizationResource[]> {
  const params = new URLSearchParams({organizationId})
  const result = await loadJson(`/resource-manager/v1/resources?${params}`, signal)
  if (!isRecord(result) || !Array.isArray(result.resources)) {
    throw new Error('Invalid resource list response')
  }
  return result.resources.map((record) => {
    if (!isRecord(record)) throw new Error('Invalid resource record')
    const kind = Object.keys(record).find((key) => {
      const candidate = record[key]
      return isRecord(candidate) && isRecord(candidate.metadata)
    })
    if (!kind) throw new Error('Unknown resource record')
    const value = record[kind] as RecordValue
    const meta = metadata(value.metadata)
    return {
      id: requiredString(meta.id, 'resource ID'),
      kind,
      name: requiredString(meta.name, 'resource name'),
      state: optionalString(meta.state)?.replace(/^RESOURCE_STATE_/, '') ?? 'UNSPECIFIED',
      revision: typeof meta.revision === 'number' ? meta.revision : undefined,
    }
  })
}
