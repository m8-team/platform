import {Chart} from '@gravity-ui/charts'
import type {ChartData} from '@gravity-ui/charts'

export interface BusinessChartSeries {
  dataKey: string
  name?: string
  kind?: 'line' | 'bar' | 'scatter'
  dashed?: boolean
  emphasis?: boolean
}

interface BusinessChartProps {
  data: readonly unknown[]
  xKey: string
  series: readonly BusinessChartSeries[]
  ariaLabel: string
  height?: 'small' | 'medium' | 'large'
  orientation?: 'vertical' | 'horizontal'
  xAxisTitle?: string
}

export function BusinessChart({
  data,
  xKey,
  series,
  ariaLabel,
  height = 'large',
  orientation = 'vertical',
  xAxisTitle,
}: BusinessChartProps) {
  const scatterOnly = series.every(({kind = 'line'}) => kind === 'scatter')
  const horizontalBars = orientation === 'horizontal'
  const chartData: ChartData = {
    legend: {enabled: series.length > 1},
    series: {
      data: series.map((item) => {
        const name = item.name ?? item.dataKey

        if (item.kind === 'bar' && horizontalBars) {
          return {
            type: 'bar-y',
            name,
            data: data.map((row, index) => ({
              x: readChartValue(row, item.dataKey),
              y: readChartValue(row, xKey) ?? index + 1,
            })),
          }
        }

        if (item.kind === 'bar') {
          return {
            type: 'bar-x',
            name,
            data: data.map((row, index) => ({
              x: readChartValue(row, xKey) ?? index + 1,
              y: readChartValue(row, item.dataKey),
            })),
          }
        }

        if (item.kind === 'scatter') {
          return {
            type: 'scatter',
            name,
            data: data.map((row, index) => ({
              x: readChartValue(row, xKey) ?? index + 1,
              y: readChartValue(row, item.dataKey),
            })),
          }
        }

        return {
          type: 'line',
          name,
          dashStyle: item.dashed ? 'Dash' : undefined,
          lineWidth: item.emphasis ? 2 : 1,
          marker: {enabled: false},
          data: data.map((row, index) => ({
            x: readChartValue(row, xKey) ?? index + 1,
            y: readChartValue(row, item.dataKey),
          })),
        }
      }),
    },
    xAxis: {
      type: horizontalBars || scatterOnly ? 'linear' : 'category',
      title: xAxisTitle ? {text: xAxisTitle} : undefined,
    },
    yAxis: [{type: horizontalBars ? 'category' : 'linear'}],
  }

  return (
    <div
      className={`ci-chart ci-chart_height_${height}`}
      role="img"
      aria-label={ariaLabel}
    >
      <Chart data={chartData} />
    </div>
  )
}

function readChartValue(row: unknown, key: string): string | number | null {
  if (!row || typeof row !== 'object') {
    return null
  }

  const value = Reflect.get(row, key)
  return typeof value === 'string' || typeof value === 'number' ? value : null
}
