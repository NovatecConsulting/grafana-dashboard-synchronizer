import { DataQuery, DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './configuration/ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { SynchronizeOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, DataQuery, SynchronizeOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
