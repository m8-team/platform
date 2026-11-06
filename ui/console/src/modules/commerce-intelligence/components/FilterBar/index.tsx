import {Button, Card, Select, Text, TextInput} from '@gravity-ui/uikit'

export type FilterConfig = {
  key: string
  label: string
  type?: 'select' | 'date'
  options?: Array<{value: string; content: string}>
  placeholder?: string
}

export type FilterValues = Record<string, string>

export function FilterSelect({
  label,
  value,
  options,
  onUpdate,
}: {
  label: string
  value: string
  options: Array<{value: string; content: string}>
  onUpdate: (value: string) => void
}) {
  return (
    <label className="ci-filter">
      <Text variant="caption-2" color="secondary">
        {label}
      </Text>
      <Select value={[value]} options={options} width="max" onUpdate={(next) => onUpdate(next[0] ?? value)} />
    </label>
  )
}

export function DateRangePicker({
  label,
  value,
  onUpdate,
}: {
  label: string
  value: string
  onUpdate: (value: string) => void
}) {
  return (
    <label className="ci-filter">
      <Text variant="caption-2" color="secondary">
        {label}
      </Text>
      <TextInput value={value} placeholder="Период" onUpdate={onUpdate} />
    </label>
  )
}

export function FilterBar({
  filters,
  values,
  onChange,
  onReset,
  onSaveView,
  primaryAction,
}: {
  filters: FilterConfig[]
  values: FilterValues
  onChange: (key: string, value: string) => void
  onReset: () => void
  onSaveView?: () => void
  primaryAction?: {label: string; onClick: () => void}
}) {
  return (
    <Card view="outlined" type="container" className="ci-filterbar">
      <div className="ci-filterbar__grid">
        {filters.map((filter) => {
          const value = values[filter.key] ?? filter.options?.[0]?.value ?? ''

          if (filter.type === 'date') {
            return <DateRangePicker key={filter.key} label={filter.label} value={value} onUpdate={(next) => onChange(filter.key, next)} />
          }

          return (
            <FilterSelect
              key={filter.key}
              label={filter.label}
              value={value}
              options={filter.options ?? [{value: 'all', content: filter.placeholder ?? 'Все'}]}
              onUpdate={(next) => onChange(filter.key, next)}
            />
          )
        })}
      </div>

      <div className="ci-filterbar__actions">
        <Button view="outlined" onClick={onReset}>
          Сбросить
        </Button>
        {onSaveView ? (
          <Button view="outlined" onClick={onSaveView}>
            Сохранить вид
          </Button>
        ) : null}
        {primaryAction ? (
          <Button view="action" onClick={primaryAction.onClick}>
            {primaryAction.label}
          </Button>
        ) : null}
      </div>
    </Card>
  )
}
