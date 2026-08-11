import {defineCatalog} from '@json-render/core';
import {schema} from '@json-render/react/schema';
import {z} from 'zod';

export const catalog = defineCatalog(schema, {
  components: {
    Page: {
      description: 'Top-level application page content container.',
      slots: ['default'],
      props: z.object({
        width: z.enum(['normal', 'wide', 'full']).default('wide'),
      }),
    },

    Stack: {
      description: 'Vertical stack of child components.',
      slots: ['default'],
      props: z.object({
        gap: z.enum(['xs', 's', 'm', 'l', 'xl']).default('m'),
      }),
    },

    Heading: {
      description: 'Page or section heading.',
      props: z.object({
        text: z.string(),
        level: z.enum(['1', '2', '3']).default('2'),
      }),
    },

    Text: {
      description: 'Body text.',
      props: z.object({
        text: z.string(),
        tone: z
          .enum([
            'primary',
            'secondary',
            'positive',
            'warning',
            'danger',
          ])
          .default('primary'),
      }),
    },

    RouteTree: {
      description: 'Hierarchical navigation tree for application routes.',
      props: z.object({
        routes: z
          .array(
            z.object({
              path: z.string().min(1),
              title: z.string().min(1),
              href: z.string().min(1).optional(),
            }),
          )
          .min(1),
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
    Card: {
      description: 'Generic content card.',
      slots: ['default'],
      props: z.object({
        title: z.string().optional(),
        titleLevel: z.enum(['2', '3']).default('2'),
      }),
    },

    ThemeSwitcher: {
      description: 'Switch between light and dark application themes.',
      props: z.object({
        label: z.string().default('Dark theme'),
      }),
    },

    Button: {
      description: 'User action button.',
      props: z.object({
        label: z.string(),
        view: z
          .enum(['normal', 'action', 'outlined', 'flat', 'raised'])
          .default('normal'),
        toast: z
          .object({
            name: z.string().min(1),
            title: z.string(),
            content: z.string().optional(),
            theme: z
              .enum([
                'normal',
                'info',
                'success',
                'warning',
                'danger',
                'utility',
              ])
              .default('normal'),
          })
          .optional(),
      }),
    },
  },

  actions: {},
});
