import {Card, Text} from '@gravity-ui/uikit'

export function ScenarioBuilder({
  items,
}: {
  items: Array<{metric: string; base: string; scenarioA: string; scenarioB: string}>
}) {
  return (
    <Card view="outlined" type="container" className="ci-card">
      <div className="ci-card__header">
        <div>
          <Text as="h2" variant="header-1">
            Конструктор сценариев
          </Text>
          <Text variant="caption-2" color="secondary">
            Сравнение базового сценария, максимума маржи и очистки запасов
          </Text>
        </div>
      </div>
      <div className="ci-scenario">
        <div className="ci-scenario__head">
          <Text variant="caption-2" color="secondary">Метрика</Text>
          <Text variant="caption-2" color="secondary">Базовый</Text>
          <Text variant="caption-2" color="secondary">Сценарий A: Максимум маржи</Text>
          <Text variant="caption-2" color="secondary">Сценарий B: Очистка запасов</Text>
        </div>
        {items.map((item) => (
          <div className="ci-scenario__row" key={item.metric}>
            <Text variant="body-2">{item.metric}</Text>
            <Text variant="body-2">{item.base}</Text>
            <Text variant="body-2">{item.scenarioA}</Text>
            <Text variant="body-2">{item.scenarioB}</Text>
          </div>
        ))}
      </div>
    </Card>
  )
}
