import {defineModule} from '@/platform/modules/define-module';
import {overviewRoute} from './routes/overview';
import {organizationsRoute} from './routes/organizations';
import {organizationDetailsRoute} from './routes/organization-details';
import {projectsRoute} from './routes/projects';
import {projectDetailsRoute} from './routes/project-details';
import {
  getOrganizationQuery,
  listOrganizationsQuery,
} from './queries/organizations';
import {getProjectQuery, listProjectsQuery} from './queries/projects';
import {createProjectOperation} from './operations/create-project';
import {deleteProjectOperation} from './operations/delete-project';

export const resourceManagerModule = defineModule({
  id: 'resource-manager',
  title: 'Resource Manager',
  basePath: '/resource-manager',
  icon: 'FolderTree',
  order: 10,
  access: {
    permission: 'resource-manager.read',
    feature: 'resource-manager',
    editions: ['community', 'enterprise'],
  },
  dependencies: {
    required: [],
    optional: ['audit'],
  },
  routes: {
    '/': overviewRoute,
    '/organizations': organizationsRoute,
    '/organizations/[organizationId]': organizationDetailsRoute,
    '/projects': projectsRoute,
    '/projects/[projectId]': projectDetailsRoute,
  },
  queries: [
    listOrganizationsQuery,
    getOrganizationQuery,
    listProjectsQuery,
    getProjectQuery,
  ],
  operations: [createProjectOperation, deleteProjectOperation],
});
