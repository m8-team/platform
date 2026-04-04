import * as React from 'react';

import {CircleExclamation, CircleInfo} from '@gravity-ui/icons';
import {selectionColumn, useTable, Table as GravityTable} from '@gravity-ui/table';
import type {ColumnDef, RowSelectionState} from '@gravity-ui/table/tanstack';
import {
  Alert,
  Button,
  Card,
  Flex,
  Icon,
  Label,
  Loader,
  Pagination,
  Tabs,
  Text,
} from '@gravity-ui/uikit';

import {formatCount, formatDateTime, humanizeStatus, statusToSeverity} from '@/shared/lib/format';
import type {ActivityItem, BreadcrumbItem, EntityStatus, OperationStep, Severity, SummaryMetric} from '@/shared/types/iam';

type PageHeaderProps = {
  eyebrow?: string;
  title: string;
  description?: string;
  actions?: React.ReactNode;
  compact?: boolean;
};

export function PageHeader({eyebrow, title, description, actions, compact}: PageHeaderProps) {
  return (
    <div className={`page-header${compact ? ' page-header_compact' : ''}`}>
      <div className="page-header__copy">
        {eyebrow ? <Text variant="caption-2" color="secondary">{eyebrow}</Text> : null}
        <Text variant="display-1" as="h1" className="page-header__title">
          {title}
        </Text>
        {description ? (
          <Text variant="body-2" color="secondary" className="page-header__description">
            {description}
          </Text>
        ) : null}
      </div>
      {actions ? <div className="page-header__actions">{actions}</div> : null}
    </div>
  );
}

export function BreadcrumbTrail({items}: {items: BreadcrumbItem[]}) {
  return (
    <div className="breadcrumbs-trail">
      {items.map((item, index) => (
        <React.Fragment key={`${item.href ?? item.label}-${index}`}>
          {index > 0 ? <span className="breadcrumbs-trail__separator">/</span> : null}
          {item.href ? (
            <a className="breadcrumbs-trail__link" href={item.href}>
              {item.label}
            </a>
          ) : (
            <span className="breadcrumbs-trail__current">{item.label}</span>
          )}
        </React.Fragment>
      ))}
    </div>
  );
}

export function StatusBadge({status}: {status: EntityStatus}) {
  return (
    <Label theme={statusToSeverity(status)} size="s">
      {humanizeStatus(status)}
    </Label>
  );
}

export function ToneBadge({tone, children}: {tone: Severity; children: React.ReactNode}) {
  return (
    <Label
      theme={
        tone === 'success'
          ? 'success'
          : tone === 'warning'
            ? 'warning'
            : tone === 'danger'
              ? 'danger'
              : 'info'
      }
      size="s"
    >
      {children}
    </Label>
  );
}

export function SectionCard({
  title,
  description,
  actions,
  children,
  className,
}: React.PropsWithChildren<{
  title: string;
  description?: string;
  actions?: React.ReactNode;
  className?: string;
}>) {
  return (
    <Card className={`section-card${className ? ` ${className}` : ''}`}>
      <div className="section-card__header">
        <div>
          <Text variant="subheader-2" as="h2">
            {title}
          </Text>
          {description ? (
            <Text variant="body-1" color="secondary">
              {description}
            </Text>
          ) : null}
        </div>
        {actions ? <div className="section-card__actions">{actions}</div> : null}
      </div>
      <div className="section-card__body">{children}</div>
    </Card>
  );
}

export function MetricGrid({items}: {items: SummaryMetric[]}) {
  return (
    <div className="metric-grid">
      {items.map((metric) => (
        <Card key={metric.id} className={`metric-card metric-card_tone_${metric.tone ?? 'info'}`}>
          <Text variant="caption-2" color="secondary">
            {metric.title}
          </Text>
          <Text variant="display-2" as="div" className="metric-card__value">
            {metric.value}
          </Text>
          <div className="metric-card__meta">
            {metric.delta ? <ToneBadge tone={metric.tone ?? 'info'}>{metric.delta}</ToneBadge> : null}
            {metric.description ? (
              <Text variant="body-1" color="secondary">
                {metric.description}
              </Text>
            ) : null}
          </div>
        </Card>
      ))}
    </div>
  );
}

export function ActivityFeed({items}: {items: ActivityItem[]}) {
  return (
    <div className="timeline-list">
      {items.map((item) => (
        <div key={item.id} className="timeline-list__item">
          <ToneBadge tone={item.tone ?? 'info'}>{item.time}</ToneBadge>
          <div>
            <Text variant="subheader-1" as="div">
              {item.title}
            </Text>
            <Text variant="body-1" color="secondary">
              {item.description}
            </Text>
          </div>
        </div>
      ))}
    </div>
  );
}

export function OperationTimeline({steps}: {steps: OperationStep[]}) {
  return (
    <div className="operation-timeline">
      {steps.map((step) => (
        <div key={step.id} className="operation-timeline__item">
          <StatusBadge status={step.status} />
          <Text variant="body-2">{step.title}</Text>
        </div>
      ))}
    </div>
  );
}

export function FilterBar({children}: React.PropsWithChildren) {
  return <div className="filter-bar">{children}</div>;
}

export function KeyValueGrid({
  items,
  columns = 3,
}: {
  items: Array<{label: string; value: React.ReactNode}>;
  columns?: number;
}) {
  return (
    <div
      className="key-value-grid"
      style={{gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`}}
    >
      {items.map((item) => (
        <div key={item.label} className="key-value-grid__item">
          <Text variant="caption-2" color="secondary">
            {item.label}
          </Text>
          <Text variant="body-2" as="div">
            {item.value}
          </Text>
        </div>
      ))}
    </div>
  );
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="state-card state-card_empty">
      <Icon data={CircleInfo} size={24} />
      <Text variant="subheader-2">{title}</Text>
      {description ? <Text color="secondary">{description}</Text> : null}
      {action}
    </div>
  );
}

export function LoadingState({title = 'Загрузка данных'}: {title?: string}) {
  return (
    <div className="state-card">
      <Loader size="l" />
      <Text variant="subheader-2">{title}</Text>
    </div>
  );
}

export function ErrorState({
  title = 'Не удалось загрузить экран',
  description,
  onRetry,
}: {
  title?: string;
  description?: string;
  onRetry?: () => void;
}) {
  return (
    <div className="state-card state-card_error">
      <Icon data={CircleExclamation} size={24} />
      <Text variant="subheader-2">{title}</Text>
      {description ? <Text color="secondary">{description}</Text> : null}
      {onRetry ? (
        <Button view="outlined" onClick={onRetry}>
          Повторить
        </Button>
      ) : null}
    </div>
  );
}

export function DetailTabs({
  items,
  activeTab,
  onSelectTab,
}: {
  items: Array<{id: string; title: string}>;
  activeTab: string;
  onSelectTab: (tabId: string) => void;
}) {
  return (
    <Tabs
      className="detail-tabs"
      size="l"
      items={items}
      activeTab={activeTab}
      onSelectTab={onSelectTab}
    />
  );
}

type TableAction<T> = {
  label: string;
  view?: 'flat' | 'outlined' | 'normal' | 'action' | 'flat-secondary' | 'outlined-action';
  onClick: (rows: T[]) => void;
};

export function DataTableCard<T extends {id: string}>({
  title,
  description,
  data,
  columns,
  emptyTitle,
  emptyDescription,
  selectable,
  bulkActions,
}: {
  title: string;
  description?: string;
  data: T[];
  columns: ColumnDef<T>[];
  emptyTitle?: string;
  emptyDescription?: string;
  selectable?: boolean;
  bulkActions?: TableAction<T>[];
}) {
  const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({});
  const [page, setPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(10);

  const pageStart = (page - 1) * pageSize;
  const pageItems = data.slice(pageStart, pageStart + pageSize);
  const tableColumns = selectable
    ? [selectionColumn as ColumnDef<T>, ...columns]
    : columns;

  const table = useTable({
    columns: tableColumns,
    data: pageItems,
    enableRowSelection: selectable,
    getRowId: (row) => row.id,
    onRowSelectionChange: setRowSelection,
    state: {
      rowSelection,
    },
  });

  const selectedItems = table.getSelectedRowModel().rows.map((row) => row.original);

  return (
    <SectionCard title={title} description={description}>
      {bulkActions && selectedItems.length > 0 ? (
        <div className="table-bulk-actions">
          <Text variant="body-1" color="secondary">
            Выбрано: {formatCount(selectedItems.length)}
          </Text>
          <Flex gap="2" wrap>
            {bulkActions.map((action) => (
              <Button
                key={action.label}
                size="s"
                view={action.view ?? 'outlined'}
                onClick={() => action.onClick(selectedItems)}
              >
                {action.label}
              </Button>
            ))}
          </Flex>
        </div>
      ) : null}
      {data.length > 0 ? (
        <>
          <div className="table-wrap">
            <GravityTable table={table} />
          </div>
          <div className="table-footer">
            <Text variant="caption-2" color="secondary">
              {formatCount(data.length)} записей
            </Text>
            <Pagination
              page={page}
              pageSize={pageSize}
              total={data.length}
              pageSizeOptions={[10, 20, 50]}
              onUpdate={(nextPage, nextPageSize) => {
                setPage(nextPage);
                setPageSize(nextPageSize);
              }}
            />
          </div>
        </>
      ) : (
        <EmptyState
          title={emptyTitle ?? 'Данные отсутствуют'}
          description={emptyDescription ?? 'Нет записей для текущих фильтров.'}
        />
      )}
    </SectionCard>
  );
}

export function ResourceList({items}: {items: string[]}) {
  return (
    <div className="resource-list">
      {items.map((item) => (
        <div key={item} className="resource-list__item">
          <span className="resource-list__dot" />
          <Text variant="body-2">{item}</Text>
        </div>
      ))}
    </div>
  );
}

export function PillList({items}: {items: string[]}) {
  return (
    <div className="pill-list">
      {items.map((item) => (
        <Label key={item} theme="utility" size="s">
          {item}
        </Label>
      ))}
    </div>
  );
}

export function JsonCodeBlock({value}: {value: unknown}) {
  return (
    <pre className="json-code-block">
      {JSON.stringify(value, null, 2)}
    </pre>
  );
}

export function HighlightAlert({
  title,
  message,
  theme = 'info',
}: {
  title: string;
  message: React.ReactNode;
  theme?: 'info' | 'warning' | 'danger' | 'success';
}) {
  return (
    <Alert theme={theme} title={title} message={message} />
  );
}

export function FieldHint({children}: React.PropsWithChildren) {
  return (
    <Text variant="caption-2" color="secondary">
      {children}
    </Text>
  );
}

export function DateTimeText({value}: {value?: string | null}) {
  return <Text variant="body-2">{formatDateTime(value)}</Text>;
}
