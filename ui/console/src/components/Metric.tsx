import {Card, Text} from '@gravity-ui/uikit'

interface MetricProps {
  label: string
  value: string
  description: string
  tone?: 'normal' | 'warning' | 'danger'
}

export function Metric({label, value, description, tone = 'normal'}: MetricProps) {
  return (
    <Card view="outlined" type="container" className={`m8-metric m8-metric_${tone}`}>
      <Text variant="caption-2" color="secondary">
        {label}
      </Text>
      <Text variant="header-2">{value}</Text>
      <Text variant="caption-2" color="secondary">
        {description}
      </Text>
    </Card>
  )
}
