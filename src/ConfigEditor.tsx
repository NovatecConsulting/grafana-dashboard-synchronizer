import React, { ChangeEvent, PureComponent } from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions } from './types';
import { Checkbox, Field, Input, RadioButtonGroup } from '@grafana/ui';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

const options = [
  { label: 'Push', value: 'push' },
  { label: 'Pull', value: 'pull' },
  { label: 'Pull and Push', value: 'pullandpush' },
];

export class ConfigEditor extends PureComponent<Props, State> {
  onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        token: event.target.value,
      },
    });
  };

  onGitURLChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      jsonData: {
        gitURL: event.target.value,
      },
    });
  };

  render() {
    return (
      <div className="gf-form-group">
        <Field label="Grafana API Token" description="Token for grafana api authorization">
          <Input id="token" onChange={this.onAPIKeyChange} />
        </Field>
        <Field label="Tag" description="Tag to synchronize to">
          <Input id="tag" />
        </Field>
        <Field label="Sync Interval" description="Interval for synchronization">
          <Input id="syncInterval" />
        </Field>
        <Field label="Git URL" description="Git URL for synchronization">
          <Input id="gitURL" onChange={this.onGitURLChange} />
        </Field>
        <Field label="Git Branch" description="Git Branch for synchronization">
          <Input id="gitBranch" />
        </Field>
        <Field label="PKK" description="Private ssh key file path for git authorization">
          <Input id="privateKeyFilePath" />
        </Field>
        <RadioButtonGroup options={options} />
        <div>
          <Checkbox disabled={false} label={'Activate'} />
        </div>
      </div>
    );
  }
}
