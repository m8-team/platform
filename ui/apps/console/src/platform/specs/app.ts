import type {NextAppSpec} from '@json-render/next';
import {buildNextAppSpec} from '@m8/runtime';
import {moduleRegistry} from '@/platform/modules/registry';
import type {RouteTreeRoute} from '@/platform/catalog/components/navigation';

const moduleSpec = buildNextAppSpec(moduleRegistry, {
  defaultLayout: 'platform',
});

type SpecRoutes = NextAppSpec['routes'];
type SpecRoute = SpecRoutes[string];
type RoutePage = NonNullable<SpecRoute['page']>;
type RouteElements = RoutePage['elements'];

const dynamicRoutePattern = /\[[^/]+\]/;

function getRouteTitle(route: SpecRoute, path: string): string {
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

function createRouteDirectoryItems(routes: SpecRoutes): RouteTreeRoute[] {
  return Object.entries(routes)
    .sort(([left], [right]) => compareRoutePaths(left, right))
    .map(([path, route]) => ({
      path,
      title: getRouteTitle(route, path),
      ...(dynamicRoutePattern.test(path) ? {} : {href: path}),
    }));
}

function createRouteDirectoryShell(routes: RouteTreeRoute[]): RouteElements {
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
  homePage: RoutePage,
  routes: SpecRoutes,
): RoutePage {
  const homeRoot = homePage.elements[homePage.root];
  if (!homeRoot) {
    return homePage;
  }

  const routeDirectory = createRouteDirectoryItems(routes);
  if (routeDirectory.length === 0) {
    return homePage;
  }

  const contentRootId = homeRoot.type === 'Page' && homeRoot.children?.length === 1
    ? homeRoot.children[0]
    : homePage.root;
  const contentRoot = homePage.elements[contentRootId];
  if (!contentRoot) return homePage;

  return {
    ...homePage,
    elements: {
      ...homePage.elements,
      [contentRootId]: {
        ...contentRoot,
        children: [...(contentRoot.children ?? []), 'homeRouteDirectory'],
      },
      ...createRouteDirectoryShell(routeDirectory),
    },
  };
}

function withHomeRouteDirectory(routes: SpecRoutes): SpecRoutes {
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

const baseSpec: NextAppSpec = {
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
          type: 'Stack',
          props: {gap: 's'},
          children: ['navigation', 'slot'],
        },

        navigation: {
          type: 'ApplicationNavigation',
          props: {
            ariaLabel: 'M8 Platform modules',
          },
          children: [],
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
        root: 'homePage',

        elements: {
          homePage: {
            type: 'Page',
            props: {width: 'wide'},
            children: ['root'],
          },

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

const mergedRoutes: SpecRoutes = {
  ...baseSpec.routes,
  ...moduleSpec.routes,
};

export const appSpec: NextAppSpec = {
  ...baseSpec,
  routes: withHomeRouteDirectory(mergedRoutes),
};
