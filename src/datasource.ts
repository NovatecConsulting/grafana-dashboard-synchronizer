import { DataQuery, DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { SynchronizeOptions } from './types';

export class DataSource extends DataSourceWithBackend<DataQuery, SynchronizeOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<SynchronizeOptions>) {
    super(instanceSettings);
  }
}
