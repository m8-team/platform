import * as React from 'react';

import {Outlet, useLocation, useNavigate} from '@tanstack/react-router';
import {AsideHeader} from '@gravity-ui/navigation';
import {Magnifier, Shield} from '@gravity-ui/icons';
import {Button, Icon, Label, Select, Text, TextInput} from '@gravity-ui/uikit';

import {useAppUI} from '@/app/providers/app-providers';
import {buildBreadcrumbs, getActiveNavigation, navigationItems} from '@/shared/config/navigation';
import {BreadcrumbTrail} from '@/shared/ui/blocks';

const tenantOptions = [
  {value: 'tenant-demo', content: 'tenant-demo'},
  {value: 'tenant-sandbox', content: 'tenant-sandbox'},
];

const organizationOptions = [
  {value: 'org-1', content: 'org-1'},
  {value: 'org-4', content: 'org-4'},
];

const environmentOptions = [
  {value: 'prod', content: 'prod'},
  {value: 'staging', content: 'staging'},
  {value: 'dev', content: 'dev'},
];

const regionOptions = [
  {value: 'eu-central', content: 'eu-central'},
  {value: 'us-east', content: 'us-east'},
];

export function AppShell() {
  const location = useLocation();
  const navigate = useNavigate();
  const {context, globalSearch, navCompact, setContext, setGlobalSearch, setNavCompact} = useAppUI();
  const [searchInput, setSearchInput] = React.useState(globalSearch);

  React.useEffect(() => {
    setSearchInput(globalSearch);
  }, [globalSearch]);

  const activeNavigation = getActiveNavigation(location.pathname);
  const breadcrumbs = buildBreadcrumbs(location.pathname);

  return (
    <div className="app-shell">
      <aside className="app-shell__aside">
        <AsideHeader
          compact={navCompact}
          onChangeCompact={setNavCompact}
          logo={{text: 'M8 IAM', icon: Shield}}
          menuItems={navigationItems.map((item) => ({
            id: item.id,
            title: item.title,
            icon: item.icon,
            current: activeNavigation === item.id,
            onItemClick: () => navigate({to: item.to}),
          }))}
          renderFooter={() => (
            <div className="aside-footer">
              <Label theme="success" size="s">live-ready</Label>
              <Text variant="caption-2" color="secondary">
                mock/live repositories
              </Text>
            </div>
          )}
        />
      </aside>
      <div className="app-shell__main">
        <header className="app-shell__topbar">
          <div className="app-shell__search">
            <TextInput
              size="l"
              placeholder="Global search"
              value={searchInput}
              onUpdate={setSearchInput}
              onKeyDown={(event) => {
                if (event.key === 'Enter') {
                  React.startTransition(() => {
                    setGlobalSearch(searchInput);
                    navigate({to: '/search'});
                  });
                }
              }}
            />
            <Button
              view="action"
              onClick={() => {
                React.startTransition(() => {
                  setGlobalSearch(searchInput);
                  navigate({to: '/search'});
                });
              }}
            >
              <Icon data={Magnifier} />
              Search
            </Button>
          </div>
          <div className="app-shell__controls">
            <Select
              label="Tenant"
              value={[context.tenantId]}
              options={tenantOptions}
              onUpdate={(value) =>
                setContext((current) => ({...current, tenantId: value[0] ?? current.tenantId}))
              }
            />
            <Select
              label="Org"
              value={[context.organizationId]}
              options={organizationOptions}
              onUpdate={(value) =>
                setContext((current) => ({
                  ...current,
                  organizationId: value[0] ?? current.organizationId,
                }))
              }
            />
            <Select
              label="Env"
              value={[context.environment]}
              options={environmentOptions}
              onUpdate={(value) =>
                setContext((current) => ({
                  ...current,
                  environment: value[0] ?? current.environment,
                }))
              }
            />
            <Select
              label="Region"
              value={[context.region]}
              options={regionOptions}
              onUpdate={(value) =>
                setContext((current) => ({...current, region: value[0] ?? current.region}))
              }
            />
            <div className="app-shell__profile">
              <Label theme="utility" size="m">Me</Label>
            </div>
          </div>
        </header>
        <div className="app-shell__breadcrumbs">
          <BreadcrumbTrail items={breadcrumbs} />
        </div>
        <main className="app-shell__content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
