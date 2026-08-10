import {lazyRouteComponent} from '@tanstack/react-router'

export const commerceIntelligenceRoutes = [
  {
    path: '/commerce-intelligence/overview',
    component: lazyRouteComponent(() => import('./pages/OverviewPage'), 'OverviewPage'),
  },
  {
    path: '/commerce-intelligence/price-actions',
    component: lazyRouteComponent(() => import('./pages/PriceActionsPage'), 'PriceActionsPage'),
  },
  {
    path: '/commerce-intelligence/products',
    component: lazyRouteComponent(() => import('./pages/ProductsPage'), 'ProductsPage'),
  },
  {
    path: '/commerce-intelligence/competitors',
    component: lazyRouteComponent(() => import('./pages/CompetitorsPage'), 'CompetitorsPage'),
  },
  {
    path: '/commerce-intelligence/forecasts',
    component: lazyRouteComponent(() => import('./pages/ForecastsPage'), 'ForecastsPage'),
  },
  {
    path: '/commerce-intelligence/markdown',
    component: lazyRouteComponent(() => import('./pages/MarkdownCenterPage'), 'MarkdownCenterPage'),
  },
  {
    path: '/commerce-intelligence/simulation',
    component: lazyRouteComponent(() => import('./pages/SimulationPage'), 'SimulationPage'),
  },
  {
    path: '/commerce-intelligence/rules',
    component: lazyRouteComponent(() => import('./pages/RulesPage'), 'RulesPage'),
  },
  {
    path: '/commerce-intelligence/approvals',
    component: lazyRouteComponent(() => import('./pages/ApprovalsPage'), 'ApprovalsPage'),
  },
  {
    path: '/commerce-intelligence/integrations',
    component: lazyRouteComponent(() => import('./pages/IntegrationsPage'), 'IntegrationsPage'),
  },
] as const
