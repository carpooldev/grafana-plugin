import { DataSourceInstanceSettings, CoreApp, AnnotationSupport, AnnotationQuery } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { CarpoolQuery, CarpoolDataSourceOptions, QueryType } from './types';
import { QueryEditor } from 'components/QueryEditor';

export class DataSource extends DataSourceWithBackend<CarpoolQuery, CarpoolDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<CarpoolDataSourceOptions>) {
    super(instanceSettings);
    super.annotations = CarpoolAnnotationSupport(this);
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


export const CarpoolAnnotationSupport: (
  ds: DataSource
) => AnnotationSupport<CarpoolQuery> = (ds: DataSource) => {
  return {
    prepareAnnotation: (
      query: AnnotationQuery<CarpoolQuery>
    ): AnnotationQuery<CarpoolQuery> => {
      return query
    },
    prepareQuery: (anno: AnnotationQuery<CarpoolQuery>) => {
      if (!anno.target) {
        return undefined;
      }

      return anno.target;
    },
    QueryEditor: QueryEditor
  };
};
