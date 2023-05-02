import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { CarpoolQuery, CarpoolDataSourceOptions, QueryType } from './types';

export class DataSource extends DataSourceWithBackend<CarpoolQuery, CarpoolDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<CarpoolDataSourceOptions>) {

    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<CarpoolQuery> {
    return {
      payload: {
        programId: '',
        queryType: QueryType.ProgramInvocations,
      }
    }
  }
}
