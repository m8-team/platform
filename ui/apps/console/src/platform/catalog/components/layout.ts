import {z} from 'zod';

export const layoutComponents = {
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

  Card: {
    description: 'Generic content card.',
    slots: ['default'],
    props: z.object({
      title: z.string().optional(),
      titleLevel: z.enum(['2', '3']).default('2'),
    }),
  },
};
