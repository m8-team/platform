import type {NextAppSpec} from '@json-render/next';

import {resourceManagerModule} from '@m8/resource-manager-module';
import {defineModules} from '@/platform/modules/define-modules';

const moduleSpec = defineModules([resourceManagerModule]).buildNextAppSpec();

type AppRoutes = NextAppSpec['routes'];
type AppRoute = AppRoutes[string];
type AppPage = NonNullable<AppRoute['page']>;
type AppElements = AppPage['elements'];

type RouteDirectoryItem = {
  id: string;
  element: AppElements[string];
};

const dynamicRoutePattern = /\[[^/]+\]/;

function getRouteTitle(route: AppRoute, path: string): string {
  const title = route.metadata?.title;
  return typeof title === 'string' ? title : path;
}

function compareRoutePaths(left: string, right: string): number {
  if (left === '/') {
    return -1;
  }
  if (right === '/') {
    return 1;
  }
  return left.localeCompare(right);
}

function createRouteDirectoryItem(
  path: string,
  route: AppRoute,
  index: number,
): RouteDirectoryItem {
  const id = `homeRouteDirectoryItem${index}`;
  const label = `${getRouteTitle(route, path)} — ${path}`;

  if (dynamicRoutePattern.test(path)) {
    return {
      id,
      element: {
        type: 'Text',
        props: {
          text: `${label} (requires route parameters)`,
          tone: 'secondary',
        },
        children: [],
      },
    };
  }

  return {
    id,
    element: {
      type: 'Link',
      props: {
        label,
        href: path,
        view: 'primary',
      },
      children: [],
    },
  };
}

function createRouteDirectoryElements(routes: AppRoutes): {
  itemIds: string[];
  elements: AppElements;
} {
  const routeItems = Object.entries(routes)
    .sort(([left], [right]) => compareRoutePaths(left, right))
    .map(([path, route], index) =>
      createRouteDirectoryItem(path, route, index),
    );

  return {
    itemIds: routeItems.map(({id}) => id),
    elements: Object.fromEntries(
      routeItems.map(({id, element}) => [id, element]),
    ),
  };
}

function createRouteDirectoryShell(itemIds: string[]): AppElements {
  return {
    homeRouteDirectory: {
      type: 'Card',
      props: {
        title: 'Pages and routes',
        titleLevel: '2',
      },
      children: ['homeRouteDirectoryContent'],
    },
    homeRouteDirectoryContent: {
      type: 'Stack',
      props: {
        gap: 'm',
      },
      children: ['homeRouteDirectoryHint', 'homeRouteDirectoryList'],
    },
    homeRouteDirectoryHint: {
      type: 'Text',
      props: {
        text: 'Open any static page. Parameterized routes are listed as templates and require a concrete resource ID.',
        tone: 'secondary',
      },
      children: [],
    },
    homeRouteDirectoryList: {
      type: 'Stack',
      props: {
        gap: 's',
      },
      children: itemIds,
    },
  };
}

function addRouteDirectoryToPage(
  homePage: AppPage,
  routes: AppRoutes,
): AppPage {
  const homeRoot = homePage.elements[homePage.root];
  if (!homeRoot) {
    return homePage;
  }

  const routeDirectory = createRouteDirectoryElements(routes);

  return {
    ...homePage,
    elements: {
      ...homePage.elements,
      [homePage.root]: {
        ...homeRoot,
        children: [...(homeRoot.children ?? []), 'homeRouteDirectory'],
      },
      ...createRouteDirectoryShell(routeDirectory.itemIds),
      ...routeDirectory.elements,
    },
  };
}

function withHomeRouteDirectory(routes: AppRoutes): AppRoutes {
  const homeRoute = routes['/'];
  if (!homeRoute?.page) {
    return routes;
  }

  return {
    ...routes,
    '/': {
      ...homeRoute,
      page: addRouteDirectoryToPage(homeRoute.page, routes),
    },
  };
}

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

const mergedRoutes: AppRoutes = {
  ...platformSpec.routes,
  ...moduleSpec.routes,
};

export const appSpec: NextAppSpec = {
  ...platformSpec,
  routes: withHomeRouteDirectory(mergedRoutes),
};
