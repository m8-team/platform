export type Tone = 'positive' | 'negative' | 'neutral'
export type StatusTone = 'success' | 'warning' | 'danger' | 'info' | 'neutral' | 'utility'

export type Kpi = {
  title: string
  value: string
  delta?: string
  deltaTone?: Tone
  subtitle?: string
  sparkline?: number[]
}

export type FilterOption = {
  label: string
  value: string
  options: Array<{value: string; content: string}>
}

export type HeatmapCell = {
  row: string
  values: Array<{column: string; value: number; label?: string; tone?: StatusTone}>
}

export type ForecastPoint = {
  date: string
  forecast: number
  actual: number
  upper: number
  lower: number
  promo?: number
}

export type PriceAction = {
  id: string
  sku: string
  product: string
  category: string
  currentPrice: string
  recommendedPrice: string
  deltaPct: string
  reason: string
  expectedRevenue: string
  expectedMargin: string
  confidence: string
  guardrailStatus: 'Пройдено' | 'Исключение' | 'Ошибка'
  approver: string
  status: 'Черновик' | 'На проверке' | 'Согласовано' | 'Запланировано' | 'Применено' | 'Отклонено'
  risk: 'Высокая' | 'Средняя' | 'Низкая'
}

export type Product = {
  sku: string
  product: string
  category: string
  currentPrice: string
  marketPrice: string
  priceIndex: number
  stock: string
  coverageDays: number
  sellThrough: string
  sales7d: string
  forecast30d: string
  elasticity: string
  lifecycle: string
  risk: 'Норма' | 'Риск' | 'Высокая' | 'Средняя' | 'Низкая'
  status: string
}

export type CompetitorMatch = {
  id: string
  sku: string
  competitor: string
  competitorProduct: string
  ourPrice: string
  competitorPrice: string
  delivery: string
  availability: string
  seller: string
  matchConfidence: string
  lastSeen: string
  differencePct: string
  alert: string
}

export type ForecastRiskSku = {
  sku: string
  product: string
  category: string
  forecast30d: string
  stock: string
  coverageDays: number
  overstockScore: number
  outOfStockScore: number
  leadTime: string
  confidence: string
  suggestedAction: string
}

export type MarkdownCandidate = {
  sku: string
  product: string
  currentPrice: string
  markdown: string
  recommendedPrice: string
  reason: string
  seasonEndStock: string
  sellThroughLift: string
  marginImpact: string
  confidence: string
  status: string
}

export type SimulationGuardrail = {
  rule: string
  limit: string
  scenarioA: string
  scenarioB: string
  status: 'Пройдено' | 'Предупреждение' | 'Ошибка'
}

export type SimulationPlannerRow = {
  sku: string
  currentPrice: string
  markdown: string
  sellThroughLift: string
  marginImpact: string
  seasonEndStock: string
  confidence: string
  status: string
}

export type Rule = {
  name: string
  type: string
  scope: string
  limit: string
  priority: string
  status: string
  updatedAt: string
  author: string
}

export type Approval = {
  id: string
  type: string
  subject: string
  decision: string
  expectedEffect: string
  risk: string
  requestedBy: string
  approver: string
  status: string
  dueAt: string
}

export type Integration = {
  provider: string
  status: string
  lastSync: string
  errors: number
  dataQuality: string
  actions: string[]
}
