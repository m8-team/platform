import {Card, Text} from '@gravity-ui/uikit'

export function ApprovalQueue({items}: {items: Array<{label: string; value: string}>}) {
  return (
    <Card view="outlined" type="container" className="ci-card ci-approval">
      <div className="ci-card__header">
        <div>
          <Text as="h2" variant="header-1">
            Очередь согласований
          </Text>
          <Text variant="caption-2" color="secondary">
            Решения, ожидающие владельцев категорий
          </Text>
        </div>
      </div>
      <div className="ci-approval__grid">
        {items.map((item) => (
          <div className="ci-approval__item" key={item.label}>
            <Text variant="header-2">{item.value}</Text>
            <Text variant="caption-2" color="secondary">
              {item.label}
            </Text>
          </div>
        ))}
      </div>
    </Card>
  )
}
