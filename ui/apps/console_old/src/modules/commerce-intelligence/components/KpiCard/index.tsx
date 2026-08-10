import type {ReactNode} from 'react'
import {Card, Text} from '@gravity-ui/uikit'

import type {Tone} from '../../mock/types'

export type KpiCardProps = {
  title: string
  value: string
  delta?: string
  deltaTone?: Tone
  subtitle?: string
  icon?: ReactNode
  sparkline?: number[]
}

function buildSparkline(points: number[]) {
  if (points.length === 0) {
    return ''
  }

  const min = Math.min(...points)
  const max = Math.max(...points)
  const spread = max - min || 1

  return points
    .map((point, index) => {
      const x = (index / Math.max(points.length - 1, 1)) * 100
      const y = 32 - ((point - min) / spread) * 26
      return `${x},${y}`
    })
    .join(' ')
}

export function KpiCard({title, value, delta, deltaTone = 'neutral', subtitle, icon, sparkline}: KpiCardProps) {
  return (
    <Card view="outlined" type="container" className="ci-kpi">
      <div className="ci-kpi__top">
        <Text variant="caption-2" color="secondary">
          {title}
        </Text>
        {icon}
      </div>
      <div className="ci-kpi__value-row">
        <Text variant="header-2">{value}</Text>
        {delta ? <span className={`ci-delta ci-delta_${deltaTone}`}>{delta}</span> : null}
      </div>
      {subtitle ? (
        <Text variant="caption-2" color="secondary">
          {subtitle}
        </Text>
      ) : null}
      {sparkline ? (
        <svg className="ci-kpi__sparkline" viewBox="0 0 100 36" preserveAspectRatio="none" aria-hidden="true">
          <polyline points={buildSparkline(sparkline)} />
        </svg>
      ) : null}
    </Card>
  )
}
