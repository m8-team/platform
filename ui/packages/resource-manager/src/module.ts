import {defineModule} from '@m8/core';

import {createProjectOperation} from './operations/create-project';
import {deleteProjectOperation} from './operations/delete-project';
import {getOrganizationQuery, listOrganizationsQuery} from './queries/organizations';
import {getProjectQuery, listProjectsQuery} from './queries/projects';
import {organizationDetailsRoute} from './routes/organization-details';
import {organizationsRoute} from './routes/organizations';
import {overviewRoute} from './routes/overview';
import {projectDetailsRoute} from './routes/project-details';
import {projectsRoute} from './routes/projects';

export const resourceManagerModule = defineModule({
  id: 'resource-manager',
  title: 'Resource Manager',
  icon: 'FolderTree',
  order: 10,
  dependencies: {
    required: [],
    optional: ['audit'],
  },
  routes: {
    '/resource-manager': overviewRoute,
    '/resource-manager/organizations': organizationsRoute,
    '/resource-manager/organizations/[organizationId]': organizationDetailsRoute,
    '/resource-manager/projects': projectsRoute,
    '/resource-manager/projects/[projectId]': projectDetailsRoute,
  },
  queries: [
    listOrganizationsQuery,
    getOrganizationQuery,
    listProjectsQuery,
    getProjectQuery,
  ],
  operations: [createProjectOperation, deleteProjectOperation],
});
