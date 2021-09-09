import React, { PureComponent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { MyDataSourceOptions, MyQuery } from './types';
import { Label } from '@grafana/ui';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  render() {
    return (
      <div className="gf-form">
        <Label description="Only for Backend">Grafana Git Workflow</Label>
      </div>
    );
  }
}
