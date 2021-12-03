import React, { PureComponent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { MyDataSourceOptions, MyQuery } from './types';
import { InfoBox } from '@grafana/ui';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  render() {
    return (
      <InfoBox
        title="Do Not Use This Data Source!"
        severity="error"
        url="https://github.com/NovatecConsulting/grafana-dashboard-sync-plugin"
      >
        This data source is <strong>not</strong> intended to be used to query data or to be used in dashboards and
        panels.
      </InfoBox>
    );
  }
}
