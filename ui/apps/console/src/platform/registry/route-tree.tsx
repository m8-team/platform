'use client';

import * as React from 'react';

import {FolderOpen} from '@gravity-ui/icons';
import {Button, DropdownMenu, Flex, Icon} from '@gravity-ui/uikit';
import {
  unstable_ListItemExpandIcon as ListItemExpandIcon,
  unstable_ListItemView as ListItemView,
  unstable_TreeList as TreeList,
  unstable_useList as useList,
  type unstable_ListTreeItemType as ListTreeItemType,
} from '@gravity-ui/uikit/unstable';

import type {RouteTreeRoute} from '@/platform/catalog/components/navigation';

type RouteTreeNodeKind = 'group' | 'route' | 'template';

type RouteTreeNodeData = RouteTreeRoute & {
  kind: RouteTreeNodeKind;
};

type MutableRouteTreeNode = {
  path: string;
  segment: string;
  route?: RouteTreeRoute;
  children: Map<string, MutableRouteTreeNode>;
};

type RouteTreeRendererProps = {
  routes: RouteTreeRoute[];
};

const routeLinkStyle: React.CSSProperties = {
  color: 'inherit',
  textDecoration: 'none',
};

function getSegmentRank(segment: string): number {
  if (segment.startsWith('[[...')) {
    return 3;
  }
  if (segment.startsWith('[...')) {
    return 2;
  }
  if (segment.startsWith('[')) {
    return 1;
  }
  return 0;
}

function compareRouteSegments(
  left: MutableRouteTreeNode,
  right: MutableRouteTreeNode,
): number {
  const rankDifference = getSegmentRank(left.segment) - getSegmentRank(right.segment);
  if (rankDifference !== 0) {
    return rankDifference;
  }

  if (left.segment === right.segment) {
    return 0;
  }

  return left.segment < right.segment ? -1 : 1;
}

function humanizeRouteSegment(segment: string): string {
  const name = segment
    .replace(/^\[\[?\.{3}/, '')
    .replace(/^\[/, '')
    .replace(/\]\]?$/, '')
    .replace(/[-_]+/g, ' ')
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .trim();

  return name ? `${name[0].toUpperCase()}${name.slice(1)}` : segment;
}

function createNodeData(node: MutableRouteTreeNode): RouteTreeNodeData {
  if (!node.route) {
    return {
      kind: 'group',
      path: node.path,
      title: humanizeRouteSegment(node.segment),
    };
  }

  return {
    ...node.route,
    kind: node.route.href ? 'route' : 'template',
  };
}

function materializeNode(
  node: MutableRouteTreeNode,
): ListTreeItemType<RouteTreeNodeData> {
  const children = [...node.children.values()]
    .sort(compareRouteSegments)
    .map(materializeNode);

  return {
    id: node.path,
    data: createNodeData(node),
    children: children.length > 0 ? children : undefined,
  };
}

function createRouteTreeItems(
  routes: RouteTreeRoute[],
): ListTreeItemType<RouteTreeNodeData>[] {
  const root: MutableRouteTreeNode = {
    path: '',
    segment: '',
    children: new Map(),
  };
  let overview: RouteTreeRoute | undefined;

  for (const route of routes) {
    if (route.path === '/') {
      overview = route;
      continue;
    }

    let parent = root;
    let accumulatedPath = '';

    for (const segment of route.path.split('/').filter(Boolean)) {
      accumulatedPath += `/${segment}`;

      let node = parent.children.get(segment);
      if (!node) {
        node = {
          path: accumulatedPath,
          segment,
          children: new Map(),
        };
        parent.children.set(segment, node);
      }

      parent = node;
    }

    parent.route = route;
  }

  const items = [...root.children.values()]
    .sort(compareRouteSegments)
    .map(materializeNode);

  if (overview) {
    items.unshift({
      id: overview.path,
      data: {
        ...overview,
        kind: overview.href ? 'route' : 'template',
      },
    });
  }

  return items;
}

function mapRouteToContent(item: RouteTreeNodeData) {
  const subtitle =
    item.kind === 'template'
      ? `${item.path} · Requires route parameters`
      : item.path;

  return {
    title: item.href ? (
      <a href={item.href} style={routeLinkStyle}>
        {item.title}
      </a>
    ) : (
      item.title
    ),
    subtitle,
  };
}

export function RouteTreeRenderer({routes}: RouteTreeRendererProps) {
  const items = React.useMemo(() => createRouteTreeItems(routes), [routes]);
  const list = useList({items, defaultExpandedState: 'expanded'});

  return (
    <TreeList
      aria-label="Application routes"
      list={list}
      mapItemDataToContentProps={mapRouteToContent}
      size="l"
      renderItem={({id, data, props: itemProps, context: {childrenIds}}) => {
        const hasChildren = Boolean(childrenIds?.length);

        return (
          <ListItemView
            {...itemProps}
            content={{
              ...itemProps.content,
              isGroup: false,
              endSlot: data.href ? (
                <DropdownMenu
                  onSwitcherClick={(event) => {
                    event.stopPropagation();
                    event.preventDefault();
                  }}
                  items={[
                    {
                      text: 'Open in new tab',
                      href: data.href,
                      target: '_blank',
                      rel: 'noreferrer',
                    },
                  ]}
                  defaultSwitcherProps={{
                    'aria-label': `Actions for ${data.title}`,
                  }}
                />
              ) : undefined,
              startSlot: hasChildren ? (
                <Button
                  size="m"
                  view="flat"
                  onClick={(event) => {
                    event.stopPropagation();
                    event.preventDefault();
                    list.state.setExpanded?.((previousState) => ({
                      ...previousState,
                      [id]: !previousState[id],
                    }));
                  }}
                  aria-label={
                    itemProps.content.expanded
                      ? `Collapse ${data.title}`
                      : `Expand ${data.title}`
                  }
                >
                  <Button.Icon>
                    <ListItemExpandIcon
                      expanded={itemProps.content.expanded}
                      behavior="action"
                    />
                  </Button.Icon>
                </Button>
              ) : (
                <Flex
                  width={28}
                  justifyContent="center"
                  spacing={
                    (itemProps.content.indentation || 0) > 0
                      ? {ml: 1}
                      : undefined
                  }
                >
                  <Icon data={FolderOpen} size={16} />
                </Flex>
              ),
            }}
          />
        );
      }}
    />
  );
}
