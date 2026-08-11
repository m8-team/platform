import {z} from 'zod';

export const routeTreeRouteSchema = z.object({
  path: z.string().min(1),
  title: z.string().min(1),
  href: z.string().min(1).optional(),
});

export type RouteTreeRoute = z.output<typeof routeTreeRouteSchema>;

export const navigationComponents = {
  RouteTree: {
    description: 'Hierarchical navigation tree for application routes.',
    props: z.object({
      routes: z.array(routeTreeRouteSchema).min(1),
    }),
  },

  Link: {
    description: 'Navigation link to an application route.',
    props: z.object({
      label: z.string(),
      href: z.string().min(1),
      view: z.enum(['normal', 'primary', 'secondary']).default('normal'),
    }),
  },
};
