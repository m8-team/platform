import type {NextAppSpec} from '@json-render/next';

import {resourceManagerModule} from '@m8/resource-manager-module';
import {defineModules} from '@/platform/modules/define-modules';

const moduleSpec = defineModules([resourceManagerModule]).buildNextAppSpec();

type AppRoutes = NextAppSpec['routes'];
type AppRoute = AppRoutes[string];
type AppPage = NonNullable<AppRoute['page']>;
type AppElements = AppPage['elements'];

type RouteDirectoryItem = {
  path: string;
  title: string;
  href?: string;
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

function createRouteDirectoryItems(routes: AppRoutes): RouteDirectoryItem[] {
  return Object.entries(routes)
    .sort(([left], [right]) => compareRoutePaths(left, right))
    .map(([path, route]) => ({
      path,
      title: getRouteTitle(route, path),
      ...(dynamicRoutePattern.test(path) ? {} : {href: path}),
    }));
}

function createRouteDirectoryShell(routes: RouteDirectoryItem[]): AppElements {
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
      children: ['homeRouteDirectoryHint', 'homeRouteDirectoryTree'],
    },
    homeRouteDirectoryHint: {
      type: 'Text',
      props: {
        text: 'Open any static page. Parameterized routes are listed as templates and require a concrete resource ID.',
        tone: 'secondary',
      },
      children: [],
    },
    homeRouteDirectoryTree: {
      type: 'RouteTree',
      props: {
        routes,
      },
      children: [],
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

  const routeDirectory = createRouteDirectoryItems(routes);
  if (routeDirectory.length === 0) {
    return homePage;
  }

  return {
    ...homePage,
    elements: {
      ...homePage.elements,
      [homePage.root]: {
        ...homeRoot,
        children: [...(homeRoot.children ?? []), 'homeRouteDirectory'],
      },
      ...createRouteDirectoryShell(routeDirectory),
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
