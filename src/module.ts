import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { CarpoolQuery, CarpoolDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, CarpoolQuery, CarpoolDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
