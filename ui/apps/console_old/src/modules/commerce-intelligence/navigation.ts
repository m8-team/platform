import {
  BarsPlay,
  Briefcase,
  Check,
  Cloud,
  Database,
  Gear,
  Layers,
  ListUl,
  Rocket,
  Speedometer,
} from '@gravity-ui/icons'

export const commerceBasePath = '/commerce-intelligence'

export const commerceNavItems = [
  {title: 'Обзор', path: `${commerceBasePath}/overview`, icon: Rocket},
  {title: 'Ценовые действия', path: `${commerceBasePath}/price-actions`, icon: Speedometer},
  {title: 'Товары', path: `${commerceBasePath}/products`, icon: Database},
  {title: 'Конкуренты', path: `${commerceBasePath}/competitors`, icon: Briefcase},
  {title: 'Прогнозы', path: `${commerceBasePath}/forecasts`, icon: BarsPlay},
  {title: 'Центр разметки', path: `${commerceBasePath}/markdown`, icon: Layers},
  {title: 'Симуляции', path: `${commerceBasePath}/simulation`, icon: Rocket},
  {title: 'Правила', path: `${commerceBasePath}/rules`, icon: Gear},
  {title: 'Согласования', path: `${commerceBasePath}/approvals`, icon: Check},
  {title: 'Интеграции', path: `${commerceBasePath}/integrations`, icon: Cloud},
  {title: 'Журнал изменений', path: `${commerceBasePath}/overview`, icon: ListUl, muted: true},
]
