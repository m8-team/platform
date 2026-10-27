import {loggedFetch} from '../../../platform/http/loggedFetch'
import type {OrganizationDetail, OrganizationStructure, OrgUnit} from './organizationDetail'

const root = (import.meta.env.VITE_RESOURCE_MANAGER_API_URL ?? '').replace(/\/$/, '')

async function send<T>(path: string, body: unknown, method: 'POST' | 'PATCH' = 'POST'): Promise<T> {
  const response = await loggedFetch('resource-manager', `${root}${path}`, {
    method,
    credentials: 'same-origin',
    headers: {'Content-Type': 'application/json', Accept: 'application/json'},
    body: JSON.stringify(body),
  })
  if (!response.ok) {
    const reason = response.status === 401 ? 'Authentication required'
      : response.status === 403 ? 'Permission denied by M8 Access'
        : response.status === 409 || response.status === 412 ? 'Resource revision conflict. Reload before retry.'
          : response.status === 501 ? 'Operation is not implemented by the server' : 'Request rejected'
    throw new Error(`${reason} (HTTP ${response.status})`)
  }
  return await response.json() as T
}
type ResourceEnvelope = {organization?: {metadata?: {id?: string; revision?: number}}}

export interface OrganizationDraft {
  tenantId: string
  name: string
  description: string
}

export async function createOrganization(draft: OrganizationDraft, key: string): Promise<string> {
  const result = await send<ResourceEnvelope>('/resource-manager/v1/resources', {
    resource: {organization: {tenantId: draft.tenantId.trim(), metadata: {
      name: draft.name.trim(), description: draft.description.trim(),
    }}},
    idempotencyKey: key,
  })
  const id = result.organization?.metadata?.id
  if (!id) throw new Error('Organization was created but the response did not contain its ID; verify before retrying with the same key.')
  return id
}

export async function updateOrganization(
  current: OrganizationDetail,
  draft: Pick<OrganizationDraft, 'name' | 'description'>,
  key: string,
): Promise<void> {
  if (!current.revision || current.revision < 1) throw new Error('Missing organization revision')
  await send<ResourceEnvelope>(`/resource-manager/v1/resources/${encodeURIComponent(current.id)}`, {
    id: current.id,
    resource: {organization: {tenantId: current.tenantId, metadata: {
      id: current.id, name: draft.name.trim(), description: draft.description.trim(),
    }}},
    updateMask: 'metadata.name,metadata.description',
    expectedRevision: current.revision,
    idempotencyKey: key,
  }, 'PATCH')
}

export async function changeOrganizationState(
  current: OrganizationDetail,
  target: 'ACTIVE' | 'SUSPENDED',
  key: string,
): Promise<void> {
  if (!current.revision || current.revision < 1) throw new Error('Missing organization revision')
  await send<ResourceEnvelope>(`/resource-manager/v1/resources/${encodeURIComponent(current.id)}:changeState`, {
    id: current.id, kind: 'organization',
    targetState: `RESOURCE_STATE_${target}`,
    expectedRevision: current.revision, idempotencyKey: key,
  })
}

export type MutationKind = 'CREATE' | 'RENAME' | 'MOVE' | 'CLOSE'
export interface UnitMutation {
  kind: MutationKind
  orgUnitId: string
  parentOrgUnitId: string
  displayName: string
  authoritativeSourceId: string
}
export interface ChangeSetDraft {
  id: string
  organizationId: string
  expectedStructureRevision: number
  mutations: UnitMutation[]
}
export interface DiffRow {
  kind: MutationKind
  unitId: string
  beforeName: string
  afterName: string
  beforeParent: string
  afterParent: string
  beforeState: string
  afterState: string
}

function toPayload(draft: ChangeSetDraft) {
  return {
    changeSet: {
      id: draft.id,
      organizationId: draft.organizationId,
      expectedStructureRevision: draft.expectedStructureRevision,
      // Do not send effectiveAt: backend only supports immediate changes.
      mutations: draft.mutations.map((m) => ({
        kind: `ORG_UNIT_MUTATION_KIND_${m.kind}`,
        orgUnitId: m.orgUnitId,
        parentOrgUnitId: m.parentOrgUnitId,
        displayName: m.displayName,
        authoritativeSourceId: m.authoritativeSourceId,
      })),
    },
  }
}
export function createDraft(organizationId: string, revision: number, mutations: UnitMutation[]): ChangeSetDraft {
  return {id: crypto.randomUUID(), organizationId, expectedStructureRevision: revision, mutations}
}
export function diffStructure(structure: OrganizationStructure, mutations: UnitMutation[]): DiffRow[] {
  const units = new Map(structure.units.map((u: OrgUnit) => [u.id, u]))
  return mutations.map((m) => {
    const before = units.get(m.orgUnitId)
    return {
      kind: m.kind,
      unitId: m.orgUnitId,
      beforeName: before?.name ?? '—',
      afterName: m.kind === 'RENAME' || m.kind === 'CREATE' ? m.displayName : before?.name ?? '—',
      beforeParent: before?.parentId || '—',
      afterParent: m.kind === 'MOVE' || m.kind === 'CREATE' ? m.parentOrgUnitId || '—' : before?.parentId || '—',
      beforeState: before?.state ?? '—',
      afterState: m.kind === 'CLOSE' ? 'CLOSED' : m.kind === 'CREATE' ? 'ACTIVE' : before?.state ?? '—',
    }
  })
}

export interface ChangeSetValidation {
  valid: boolean
  blockers: string[]
  proposedStructureRevision: number
}
export async function validateChangeSet(draft: ChangeSetDraft): Promise<ChangeSetValidation> {
  const result = await send<ChangeSetValidation>('/resource-manager/v1/org-unit-change-sets:validate', toPayload(draft))
  if (typeof result.valid !== 'boolean') throw new Error('Invalid ChangeSet validation response')
  return {valid: result.valid, blockers: result.blockers ?? [], proposedStructureRevision: result.proposedStructureRevision}
}
export async function applyChangeSet(draft: ChangeSetDraft): Promise<{structureRevision: number; replayed: boolean}> {
  return send('/resource-manager/v1/org-unit-change-sets:apply', {
    ...toPayload(draft), idempotencyKey: draft.id,
  })
}
