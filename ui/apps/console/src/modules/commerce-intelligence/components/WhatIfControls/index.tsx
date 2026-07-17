/* eslint-disable react-refresh/only-export-components */
import {Button, Card, Select, Text, TextInput} from '@gravity-ui/uikit'
import {Store} from '@tanstack/store'

export type ScenarioSettings = {
  markdown: number
  priceIndex: number
  stockDays: number
  risk: 'Низкая' | 'Средняя' | 'Высокая'
}

export const scenarioSettingsStore = new Store<ScenarioSettings>({
  markdown: 15,
  priceIndex: 98,
  stockDays: 47,
  risk: 'Средняя',
})

export function WhatIfControls({
  value,
  onChange,
  onRun,
}: {
  value: ScenarioSettings
  onChange: (value: ScenarioSettings) => void
  onRun: () => void
}) {
  const update = (patch: Partial<ScenarioSettings>) => {
    const next = {...value, ...patch}
    scenarioSettingsStore.setState(() => next)
    onChange(next)
  }

  return (
    <Card view="outlined" type="container" className="ci-card">
      <div className="ci-card__header">
        <div>
          <Text as="h2" variant="header-1">
            What-if контролы
          </Text>
          <Text variant="caption-2" color="secondary">
            Параметры для пересчета сценария
          </Text>
        </div>
      </div>
      <div className="ci-whatif">
        <label className="ci-filter">
          <Text variant="caption-2" color="secondary">Средняя уценка %</Text>
          <TextInput type="number" value={String(value.markdown)} onUpdate={(next) => update({markdown: Number(next)})} />
        </label>
        <label className="ci-filter">
          <Text variant="caption-2" color="secondary">Целевой индекс цены</Text>
          <TextInput type="number" value={String(value.priceIndex)} onUpdate={(next) => update({priceIndex: Number(next)})} />
        </label>
        <label className="ci-filter">
          <Text variant="caption-2" color="secondary">Цель по запасам, дней</Text>
          <TextInput type="number" value={String(value.stockDays)} onUpdate={(next) => update({stockDays: Number(next)})} />
        </label>
        <label className="ci-filter">
          <Text variant="caption-2" color="secondary">Толерантность к риску</Text>
          <Select
            value={[value.risk]}
            options={[
              {value: 'Низкая', content: 'Низкая'},
              {value: 'Средняя', content: 'Средняя'},
              {value: 'Высокая', content: 'Высокая'},
            ]}
            onUpdate={(next) => update({risk: (next[0] as ScenarioSettings['risk']) ?? value.risk})}
          />
        </label>
        <Button view="action" onClick={onRun}>
          Запустить симуляцию
        </Button>
      </div>
    </Card>
  )
}
