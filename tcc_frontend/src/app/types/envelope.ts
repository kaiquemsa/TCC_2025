import { ChartSpec } from './chart';

export type Envelope =
  | { type: 'text'; text: string }
  | { type: 'html'; html: string }
  | { type: 'chart'; spec: ChartSpec };
