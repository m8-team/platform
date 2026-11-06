/* eslint-disable react-refresh/only-export-components */
import {ApprovalsPage} from './pages/ApprovalsPage'
import {CompetitorsPage} from './pages/CompetitorsPage'
import {ForecastsPage} from './pages/ForecastsPage'
import {IntegrationsPage} from './pages/IntegrationsPage'
import {MarkdownCenterPage} from './pages/MarkdownCenterPage'
import {OverviewPage} from './pages/OverviewPage'
import {PriceActionsPage} from './pages/PriceActionsPage'
import {ProductsPage} from './pages/ProductsPage'
import {RulesPage} from './pages/RulesPage'
import {SimulationPage} from './pages/SimulationPage'

export const commerceIntelligenceRoutes = [
  {path: '/commerce-intelligence/overview', component: OverviewPage},
  {path: '/commerce-intelligence/price-actions', component: PriceActionsPage},
  {path: '/commerce-intelligence/products', component: ProductsPage},
  {path: '/commerce-intelligence/competitors', component: CompetitorsPage},
  {path: '/commerce-intelligence/forecasts', component: ForecastsPage},
  {path: '/commerce-intelligence/markdown', component: MarkdownCenterPage},
  {path: '/commerce-intelligence/simulation', component: SimulationPage},
  {path: '/commerce-intelligence/rules', component: RulesPage},
  {path: '/commerce-intelligence/approvals', component: ApprovalsPage},
  {path: '/commerce-intelligence/integrations', component: IntegrationsPage},
] as const

export {
  ApprovalsPage,
  CompetitorsPage,
  ForecastsPage,
  IntegrationsPage,
  MarkdownCenterPage,
  OverviewPage,
  PriceActionsPage,
  ProductsPage,
  RulesPage,
  SimulationPage,
}
