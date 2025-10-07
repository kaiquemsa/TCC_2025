export type ChartKind = 'bar' | 'line' | 'pie';

export interface ChartSeries {
  name: string;
  values: (number | null)[];
}

export interface ChartSpec {
  kind: ChartKind;
  title?: string;
  x: string[];
  series: ChartSeries[];
  yLabel?: string;
  stacked?: boolean;
  colors?: string[];
}
