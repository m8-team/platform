import {moduleRegistry} from '@/platform/modules/registry';
import {platformLayout} from './layouts/platform';

export const appSpec = moduleRegistry.buildNextAppSpec({
  metadata: {
    title: {
      default: 'M8 Platform',
      template: '%s | M8 Platform',
    },
  },
  layouts: {
    platform: platformLayout,
  },
  state: {
    actor: null,
    context: {
      organization: null,
      workspace: null,
      project: null,
    },
  },
});
