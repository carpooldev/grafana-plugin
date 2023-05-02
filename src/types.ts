import { DataQuery, DataSourceJsonData } from '@grafana/data';

export const DEFAULT_CARPOOL_HOST = 'https://api.carpool.dev';


export interface SolanaInvocationQuery {
  programId: string;
  instructionName?: string;
  queryType: QueryType;
}

export interface CarpoolQuery extends DataQuery {
  payload: SolanaInvocationQuery;
}

export enum QueryType {
  ProgramInvocations = 'invocations',
  ProgramSigners = 'uniqueSigners',
  ProgramFailureRate = 'failureRate',
  ProgramFailues = 'failures',
}

export const QueryTypes = [
  {
    label: 'Program Invocations',
    value: QueryType.ProgramInvocations,

  },
  {
    label: 'Program Signers',
    value: QueryType.ProgramSigners,
  },
  {
    label: 'Program Failure Rate',
    value: QueryType.ProgramFailureRate,
  },
  {
    label: 'Program Failures',
    value: QueryType.ProgramFailues,
  },
];
/**
 * These are options configured for each DataSource instance
 */
export interface CarpoolDataSourceOptions extends DataSourceJsonData {
  url: string;
  maxBuckets: number;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface CarpoolSecureJsonData {
  apiKey?: string;
}
