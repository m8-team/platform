import {z} from 'zod';

const toastSchema = z.object({
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
});

export const controlComponents = {
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
        .enum(['normal', 'action', 'outlined', 'outlined-danger', 'flat', 'raised'])
        .default('normal'),
      toast: toastSchema.optional(),
    }),
  },
};
