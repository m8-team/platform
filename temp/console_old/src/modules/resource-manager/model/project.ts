export type ProjectStatus = 'Active' | 'Suspended' | 'Failed' | 'Provisioning' | 'Deleting'

export interface Project {
  name: string
  projectId: string
  workspace: string
  organization: string
  status: ProjectStatus
  desiredState: string
  actualState: string
  updated: string
  owner: string
  lastOperation: string
}
