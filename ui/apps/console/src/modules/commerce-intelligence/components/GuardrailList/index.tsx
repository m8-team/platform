import {Card, Text} from '@gravity-ui/uikit'

import {StatusBadge} from '../StatusBadge'
import {statusTone} from '../../utils'

export function GuardrailList({
  title = 'Проверки guardrails',
  items,
}: {
  title?: string
  items: Array<{name?: string; rule?: string; value?: string; limit?: string; status: string; scenarioA?: string; scenarioB?: string}>
}) {
  return (
    <Card view="outlined" type="container" className="ci-card">
      <div className="ci-card__header">
        <div>
          <Text as="h2" variant="header-1">
            {title}
          </Text>
          <Text variant="caption-2" color="secondary">
            Политики безопасного изменения цен
          </Text>
        </div>
      </div>
      <div className="ci-guardrails">
        {items.map((item) => (
          <div className="ci-guardrails__item" key={item.name ?? item.rule}>
            <div>
              <Text variant="body-2">{item.name ?? item.rule}</Text>
              <Text variant="caption-2" color="secondary">
                {item.value ?? item.limit}
                {item.scenarioA ? ` · Сценарий A: ${item.scenarioA}` : ''}
                {item.scenarioB ? ` · Сценарий B: ${item.scenarioB}` : ''}
              </Text>
            </div>
            <StatusBadge tone={statusTone(item.status)}>{item.status}</StatusBadge>
          </div>
        ))}
      </div>
    </Card>
  )
}
