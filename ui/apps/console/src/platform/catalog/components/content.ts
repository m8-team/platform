import {z} from 'zod';

export const contentComponents = {
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
};
