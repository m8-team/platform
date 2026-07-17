import {Text} from '@gravity-ui/uikit'

import type {HeatmapCell} from '../../mock/types'

function cellClass(value: number) {
  if (value >= 105 || value >= 85) {
    return 'ci-heatmap__cell_hot'
  }

  if (value <= 95 || value <= 35) {
    return 'ci-heatmap__cell_cool'
  }

  return 'ci-heatmap__cell_mid'
}

export function Heatmap({rows}: {rows: HeatmapCell[]}) {
  const columns = rows[0]?.values.map((value) => value.column) ?? []

  return (
    <div className="ci-heatmap">
      <div className="ci-heatmap__row ci-heatmap__row_header" style={{gridTemplateColumns: `150px repeat(${columns.length}, minmax(92px, 1fr))`}}>
        <span />
        {columns.map((column) => (
          <Text key={column} variant="caption-2" color="secondary">
            {column}
          </Text>
        ))}
      </div>
      {rows.map((row) => (
        <div className="ci-heatmap__row" key={row.row} style={{gridTemplateColumns: `150px repeat(${columns.length}, minmax(92px, 1fr))`}}>
          <Text variant="caption-2">{row.row}</Text>
          {row.values.map((cell) => (
            <div className={`ci-heatmap__cell ${cellClass(cell.value)}`} key={`${row.row}-${cell.column}`}>
              {cell.label ?? cell.value}
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}
