'use client';

import {RouteTreeRenderer} from './route-tree';

import {defineRegistry, useBoundProp} from '@json-render/react';
import NextLink from 'next/link';
import {
  Box,
  Button,
  Card,
  Flex,
  Link as GravityLink,
  Switch,
  Text,
  TextInput as GravityTextInput,
  spacing,
} from '@gravity-ui/uikit';
import {toaster} from '@gravity-ui/uikit/toaster-singleton';

import {catalog} from '@/platform/catalog/catalog';
import {useTheme} from '@/platform/runtime/theme-context';

const gaps = {
  xs: 1,
  s: 2,
  m: 4,
  l: 6,
  xl: 8,
} as const;

const pageMaxWidths = {
  normal: 1200,
  wide: 1600,
  full: undefined,
} as const;

const headingTags = {
  '1': 'h1',
  '2': 'h2',
  '3': 'h3',
} as const;

const headingVariants = {
  '1': 'display-1',
  '2': 'header-2',
  '3': 'subheader-3',
} as const;

export const {registry} = defineRegistry(catalog, {
  actions: {
    // RuntimeProvider supplies this infrastructure handler to NextAppProvider.
    executeOperation: async () => undefined,
  },
  components: {
    Page: ({props, children}) => (
      <Box
        as="main"
        width="100%"
        maxWidth={pageMaxWidths[props.width]}
        spacing={{p: 6}}
        style={{margin: '0 auto'}}
      >
        {children}
      </Box>
    ),

    Stack: ({props, children}) => (
      <Flex direction="column" gap={gaps[props.gap]}>
        {children}
      </Flex>
    ),

    Heading: ({props}) => (
      <Text
        as={headingTags[props.level]}
        variant={headingVariants[props.level]}
      >
        {props.text}
      </Text>
    ),

    Text: ({props}) => <Text color={props.tone}>{props.text}</Text>,
    RouteTree: ({props}) => <RouteTreeRenderer routes={props.routes} />,

    Card: ({props, children}) => (
      <Card type="container" view="outlined" size="l" spacing={{p: 5}}>
        {props.title ? (
          <Text
            as={headingTags[props.titleLevel]}
            variant="subheader-3"
            className={spacing({mb: 4})}
          >
            {props.title}
          </Text>
        ) : null}

        {children}
      </Card>
    ),

    ThemeSwitcher: ({props}) => {
      const {theme, setTheme} = useTheme();

      return (
        <Switch
          checked={theme === 'dark'}
          size="l"
          onUpdate={(checked) => {
            setTheme(checked ? 'dark' : 'light');
          }}
        >
          {props.label}
        </Switch>
      );
    },

    Button: ({props, emit}) => (
      <Button
        view={props.view}
        size="l"
        onClick={() => {
          emit('press');
          if (props.toast) {
            toaster.add({
              name: props.toast.name,
              title: props.toast.title,
              content: props.toast.content,
              theme: props.toast.theme,
            });
          }
        }}
      >
        {props.label}
      </Button>
    ),

    PageHeader: ({props, children}) => <Flex direction="column" gap={2}><Text variant="display-1">{props.title}</Text>{props.description ? <Text color="secondary">{props.description}</Text> : null}{children}</Flex>,
    Grid: ({props, children}) => <div style={{display: 'grid', gridTemplateColumns: `repeat(${props.columns}, minmax(0, 1fr))`, gap: props.gap === 'l' ? 24 : props.gap === 'm' ? 16 : 8}}>{children}</div>,
    NavigationCard: ({props}) => <Card type="container" view="outlined" size="l" spacing={{p: 5}}><GravityLink href={props.href}>{props.title}</GravityLink><Text color="secondary">{props.description}</Text></Card>,
    FilterBar: ({props, bindings}) => {
      const [search, setSearch] = useBoundProp(props.search, bindings?.search);
      const [status, setStatus] = useBoundProp(props.status, bindings?.status);
      const [organizationId, setOrganizationId] = useBoundProp(props.organizationId, bindings?.organizationId);
      return <Flex gap={2} wrap="wrap">
        <GravityTextInput value={search ?? ''} placeholder={props.searchPlaceholder ?? 'Search'} onUpdate={setSearch} />
        {bindings?.status ? <GravityTextInput value={status ?? ''} placeholder="Status" onUpdate={setStatus} /> : null}
        {bindings?.organizationId ? <GravityTextInput value={organizationId ?? ''} placeholder="Organization ID" onUpdate={setOrganizationId} /> : null}
      </Flex>;
    },
    TextInput: ({props, bindings}) => {
      const [value, setValue] = useBoundProp(props.value, bindings?.value);
      return <GravityTextInput value={value ?? ''} label={props.label} placeholder={props.placeholder} onUpdate={setValue} />;
    },
    ResourceTable: ({props}) => props.loading ? <Text>Loading…</Text> : <div>{(props.rows ?? []).map((row, index) => {
      const content = <Card type="container" view="outlined" spacing={{p: 3}}>{props.columns.map(column => <Text key={column.field}>{column.title}: {String(row[column.field] ?? '')}</Text>)}</Card>;
      return typeof row.href === 'string'
        ? <NextLink key={String(row.id ?? index)} href={row.href}>{content}</NextLink>
        : <div key={String(row.id ?? index)}>{content}</div>;
    })}</div>,
    ResourceHeader: ({props}) => <Text variant="header-1">{String(props.resource?.name ?? props.resource?.id ?? props.resourceType)}</Text>,
    PropertyList: ({props}) => <Flex direction="column">{props.fields.map(field => <Text key={field.field}>{field.title}: {String(props.value?.[field.field] ?? '')}</Text>)}</Flex>,
    DangerZone: ({props, children}) => <Card type="container" view="outlined" spacing={{p: 4}}><Text color="danger">{props.title}</Text>{props.description ? <Text>{props.description}</Text> : null}{children}</Card>,
  },
});
