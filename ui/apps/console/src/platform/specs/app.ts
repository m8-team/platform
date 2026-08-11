import type {NextAppSpec} from '@json-render/next';

import {resourceManagerModule} from '@m8/resource-manager-module';
import {defineModules} from '@/platform/modules/define-modules';

const moduleSpec = defineModules([resourceManagerModule]).buildNextAppSpec();

const platformSpec: NextAppSpec = {
  metadata: {
    title: {
      default: 'M8 Platform',
      template: '%s | M8 Platform',
    },
    description: 'M8 Platform',
  },

  layouts: {
    platform: {
      root: 'layout',

      elements: {
        layout: {
          type: 'Page',
          props: {
            width: 'wide',
          },
          children: ['slot'],
        },

        slot: {
          type: 'Slot',
          props: {},
          children: [],
        },
      },
    },
  },

  routes: {
    '/': {
      layout: 'platform',

      metadata: {
        title: 'Overview',
      },

      page: {
        root: 'root',

        elements: {
          root: {
            type: 'Stack',
            props: {
              gap: 'l',
            },
            children: [
              'themeSwitcher',
              'title',
              'description',
              'button',
              'card',
            ],
          },

          themeSwitcher: {
            type: 'ThemeSwitcher',
            props: {
              label: 'Dark theme',
            },
            children: [],
          },

          title: {
            type: 'Heading',
            props: {
              text: 'M8 Platform',
              level: '1',
            },
            children: [],
          },

          description: {
            type: 'Text',
            props: {
              text: 'Declarative platform UI powered by json-render.',
              tone: 'secondary',
            },
            children: [],
          },

          button: {
            type: 'Button',
            props: {
              label: 'Get started',
              view: 'action',
              toast: {
                name: 'get-started',
                title: 'M8 Platform',
                content: 'Welcome to M8 Platform',
                theme: 'success',
              },
            },
            children: [],
          },

          card: {
            type: 'Card',
            props: {
              title: 'Platform status',
            titleLevel: '2',
            },
            children: ['status'],
          },

          status: {
            type: 'Text',
            props: {
              text: 'Platform is running',
              tone: 'positive',
            },
            children: [],
          },
        },
      },
    },
  },
};

export const appSpec: NextAppSpec = {
  ...platformSpec,
  routes: {
    ...platformSpec.routes,
    ...moduleSpec.routes,
  },
};
