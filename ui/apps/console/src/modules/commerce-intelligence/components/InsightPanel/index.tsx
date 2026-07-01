import {Card, Text} from '@gravity-ui/uikit'

import {StatusBadge} from '../StatusBadge'
import type {StatusTone} from '../../mock/types'

export type Insight = {
  title: string
  text: string
  tone: StatusTone
}

export function InsightPanel({title = 'AI-аналитика', insights}: {title?: string; insights: Insight[]}) {
  return (
    <Card view="outlined" type="container" className="ci-card ci-insights">
      <div className="ci-card__header">
        <div>
          <Text as="h2" variant="header-1">
            {title}
          </Text>
          <Text variant="caption-2" color="secondary">
            Объяснения и сигналы для приоритизации действий
          </Text>
        </div>
      </div>
      <div className="ci-insights__list">
        {insights.map((insight) => (
          <div className="ci-insight" key={insight.title}>
            <div className="ci-insight__header">
              <Text variant="body-2">{insight.title}</Text>
              <StatusBadge tone={insight.tone}>AI</StatusBadge>
            </div>
            <Text variant="caption-2" color="secondary">
              {insight.text}
            </Text>
          </div>
        ))}
      </div>
    </Card>
  )
}
