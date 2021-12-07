import React, { PureComponent } from 'react';
import { DataQuery, QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { SynchronizeOptions } from './types';
import { InfoBox } from '@grafana/ui';

type Props = QueryEditorProps<DataSource, DataQuery, SynchronizeOptions>;

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
