import { Envelope } from '../types/envelope';

export type HistoryFlag = 'Y' | 'N';

export interface QueryRequest {
  uuid: string;
  question: string;
  history: HistoryFlag;
}

export interface QueryResponse {
  sql: string;
  envelope: Envelope;
}
