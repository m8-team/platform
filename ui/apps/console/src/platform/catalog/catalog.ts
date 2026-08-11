import {defineCatalog} from '@json-render/core';
import {schema} from '@json-render/react/schema';

import {contentComponents} from './components/content';
import {controlComponents} from './components/controls';
import {layoutComponents} from './components/layout';
import {navigationComponents} from './components/navigation';
import {resourceComponents} from './components/resources';

export const catalog = defineCatalog(schema, {
  components: {
    ...layoutComponents,
    ...contentComponents,
    ...navigationComponents,
    ...controlComponents,
    ...resourceComponents,
  },

  actions: {},
});
