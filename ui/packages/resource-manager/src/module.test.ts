import {describe, expect, it} from 'vitest';

import * as publicApi from './index';
import {resourceManagerModule} from './module';
import {getProjectQuery} from './queries/projects';

describe('resourceManagerModule', () => {
  it('is the package single public registration source', () => {
    expect(Object.keys(publicApi)).toEqual(['resourceManagerModule']);
    expect(publicApi.resourceManagerModule).toBe(resourceManagerModule);
    expect(resourceManagerModule.routes).toBeDefined();
    expect(resourceManagerModule.queries).toHaveLength(4);
    expect(resourceManagerModule.operations).toHaveLength(2);
  });

  it('keeps UI routing and delete input out of the project query DTO', async () => {
    const rawProject = await getProjectQuery.execute({
      input: {projectId: 'prj_1'},
      signal: new AbortController().signal,
      context: {},
    });
    const project = getProjectQuery.output.parse(rawProject);

    expect(project).toMatchObject({id: 'prj_1', version: '1'});
    expect(project).not.toHaveProperty('href');
    expect(project).not.toHaveProperty('deleteInput');
  });
});
