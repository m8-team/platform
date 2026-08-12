import {z} from 'zod';

export const routeTreeRouteSchema = z.object({
  path: z.string().min(1),
  title: z.string().min(1),
  href: z.string().min(1).optional(),
});

export type RouteTreeRoute = z.output<typeof routeTreeRouteSchema>;

export const navigationComponents = {
  ApplicationNavigation: {
    description: 'Primary navigation composed from enabled M8 modules.',
    props: z.object({
      ariaLabel: z.string().default('Application navigation'),
    }),
  },

  RouteTree: {
    description: 'Hierarchical navigation tree for application routes.',
    props: z.object({
      routes: z.array(routeTreeRouteSchema).min(1),
    }),
  },

};
