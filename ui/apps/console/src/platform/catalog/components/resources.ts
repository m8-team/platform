import {z} from 'zod';

const field = z.object({field: z.string(), title: z.string()});

export const resourceComponents = {
  PageHeader: {description: 'Page title and actions.', slots: ['default'], props: z.object({title: z.string(), description: z.string().optional()})},
  Grid: {description: 'Responsive content grid.', slots: ['default'], props: z.object({columns: z.number().int().positive(), gap: z.enum(['s', 'm', 'l'])})},
  NavigationCard: {description: 'Card linking to a platform route.', props: z.object({title: z.string(), description: z.string(), icon: z.string().optional(), href: z.string()})},
  FilterBar: {description: 'Resource list filters.', props: z.object({search: z.string().optional().nullable(), searchPlaceholder: z.string().optional(), status: z.string().optional().nullable(), organizationId: z.string().optional().nullable()})},
  TextInput: {description: 'Two-way bound text field.', props: z.object({value: z.string().optional().nullable(), label: z.string(), placeholder: z.string().optional()})},
  ResourceTable: {description: 'Tabular resource list. Rows may provide a native href.', props: z.object({resourceType: z.string().optional(), rows: z.array(z.record(z.string(), z.unknown())).optional(), loading: z.boolean().optional(), columns: z.array(field)})},
  ResourceHeader: {description: 'Resource detail header.', props: z.object({resourceType: z.string(), resource: z.record(z.string(), z.unknown()).optional()})},
  PropertyList: {description: 'Resource properties.', props: z.object({value: z.record(z.string(), z.unknown()).optional(), fields: z.array(field)})},
  DangerZone: {description: 'Destructive action section.', slots: ['default'], props: z.object({title: z.string(), description: z.string().optional()})},
};
