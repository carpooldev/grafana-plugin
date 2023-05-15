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
  ProgramDeployments = 'programDeployments',
  FailedProgramDeployments = 'failedProgramDeployments',
}

export const QueryTypes = [
  {
    label: 'Program Invocations',
    value: QueryType.ProgramInvocations,
    fields: ['programId', 'instructionName']
  },
  {
    label: 'Program Signers',
    value: QueryType.ProgramSigners,
    fields: ['programId', 'instructionName']
  },
  {
    label: 'Program Failure Rate',
    value: QueryType.ProgramFailureRate,
    fields: ['programId', 'instructionName']
  },
  {
    label: 'Program Failures',
    value: QueryType.ProgramFailues,
    fields: ['programId', 'instructionName']
  },
  {
    label: 'Program Deployments',
    value: QueryType.ProgramDeployments,
    fields: ['programId']
  },
  {
    label: 'Failed Program Deployments',
    value: QueryType.FailedProgramDeployments,
    fields: ['programId']
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
