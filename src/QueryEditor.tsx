import React, { PureComponent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { MyDataSourceOptions, MyQuery } from './types';
import { Label } from '@grafana/ui';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  render() {
    return (
      <Label description="This data source is not intended to provide data or for use in dashboard panels.">
        Do not use this data source!
      </Label>
    );
  }
}
