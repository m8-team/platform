'use client';

import {RouteTreeRenderer} from './route-tree';

import {defineRegistry} from '@json-render/react';
import {
  Box,
  Button,
  Card,
  Flex,
  Link,
  Switch,
  Text,
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

    Link: ({props}) => (
      <Link href={props.href} view={props.view}>
        {props.label}
      </Link>
    ),

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

    Button: ({props}) => (
      <Button
        view={props.view}
        size="l"
        onClick={() => {
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
  },
});
