import {defineCatalog} from '@json-render/core';
import {schema} from '@json-render/react/schema';
import {z} from 'zod';

export const catalog = defineCatalog(schema, {
    components: {
        Page: {
            description: 'Top-level application page content container.',
            props: z.object({
                width: z.enum(['normal', 'wide', 'full']).default('wide'),
            }),
        },

        Stack: {
            description: 'Vertical stack of child components.',
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

        Card: {
            description: 'Generic content card.',
            props: z.object({
                title: z.string().optional(),
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
                        title: z.string(),
                        content: z.string().optional(),
                    })
                    .optional(),
            }),
        },
    },

    actions: {},
});
