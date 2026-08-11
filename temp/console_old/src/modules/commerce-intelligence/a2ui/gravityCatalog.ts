import {ActionToolbar} from '../components/ActionToolbar'
import {AppShell} from '../components/AppShell'
import {ApprovalQueue} from '../components/ApprovalQueue'
import {ChartCard} from '../components/ChartCard'
import {DataTable} from '../components/DataTable'
import {DetailDrawer} from '../components/DetailDrawer'
import {FilterBar, FilterSelect, DateRangePicker} from '../components/FilterBar'
import {GuardrailList} from '../components/GuardrailList'
import {Heatmap} from '../components/Heatmap'
import {InsightPanel} from '../components/InsightPanel'
import {KpiCard} from '../components/KpiCard'
import {PageHeader} from '../components/PageHeader'
import {ScenarioBuilder} from '../components/ScenarioBuilder'
import {SidebarNav} from '../components/SidebarNav'
import {StatusBadge} from '../components/StatusBadge'
import {TopBar} from '../components/TopBar'
import {WhatIfControls} from '../components/WhatIfControls'

function EmptyState() {
  return null
}

export const gravityCatalog = {
  components: {
    AppShell,
    TopBar,
    SidebarNav,
    PageHeader,
    FilterBar,
    FilterSelect,
    DateRangePicker,
    KpiCard,
    ChartCard,
    DataTable,
    Heatmap,
    InsightPanel,
    DetailDrawer,
    StatusBadge,
    ActionToolbar,
    ApprovalQueue,
    EmptyState,
    GuardrailList,
    ScenarioBuilder,
    WhatIfControls,
  },
}
