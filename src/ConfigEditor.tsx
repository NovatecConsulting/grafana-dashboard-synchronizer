import React, { FormEvent, PureComponent } from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions } from './types';
import { Checkbox, Field, Input } from '@grafana/ui';
import {} from '@emotion/core';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onSecureJsonChange = (event: FormEvent<HTMLInputElement>, key: String) => {
    const { onOptionsChange, options } = this.props;
    event.preventDefault();
    if (!this.props.options.secureJsonData) {
      this.props.options.secureJsonData = {};
    }
    this.props.options.secureJsonData[key] = event.target.value;
    onOptionsChange({
      ...options,
    });
  };

  onJsonChange = (event: FormEvent<HTMLInputElement>, key: String) => {
    const { onOptionsChange, options } = this.props;
    event.preventDefault();
    if (!this.props.options.jsonData) {
      this.props.options.jsonData = {};
    }
    this.props.options.jsonData[key] = event.target.value;
    onOptionsChange({
      ...options,
    });
  };

  onJsonCheckboxChange = (event: FormEvent<HTMLInputElement>, key: String) => {
    const { onOptionsChange, options } = this.props;
    event.preventDefault();
    if (!this.props.options.jsonData) {
      this.props.options.jsonData = {};
    }
    this.props.options.jsonData[key] = event.target.checked.toString();
    onOptionsChange({
      ...options,
    });
  };

  render() {
    return (
      <div className="gf-form-group">
        <Field label="Grafana API Token" description="Token for grafana api authorization">
          <Input id="token" onChange={event => this.onSecureJsonChange(event, 'token')} />
        </Field>
        <Field label="Grafana URL" description="URL for grafana api">
          <Input id="grafanaURL" onChange={event => this.onJsonChange(event, 'grafanaURL')} />
        </Field>
        <Field label="Tag" description="Tag to synchronize to">
          <Input id="tag" onChange={event => this.onJsonChange(event, 'tag')} />
        </Field>
        <Field label="Sync Interval" description="Interval for synchronization">
          <Input id="syncInterval" />
        </Field>
        <Field label="Git Push URL" description="Git push URL for synchronization">
          <Input id="gitPushURL" onChange={event => this.onJsonChange(event, 'gitPushURL')} />
        </Field>
        <Field label="Git Branch" description="Git Branch for synchronization">
          <Input id="gitBranch" />
        </Field>
        <Field label="Git Pull URL" description="Git pull URL for synchronization">
          <Input id="gitPullURL" onChange={event => this.onJsonChange(event, 'gitPullURL')} />
        </Field>
        <Field label="PKK" description="Private ssh key file path for git authorization">
          <Input id="privateKeyFilePath" onChange={event => this.onSecureJsonChange(event, 'privateKeyFilePath')} />
        </Field>
        <Checkbox
          label={'Pull'}
          description={'Pull from repo'}
          onChange={event => this.onJsonCheckboxChange(event, 'pull')}
        />
        <Checkbox
          label={'Push'}
          description={'Push from repo'}
          onChange={event => this.onJsonCheckboxChange(event, 'push')}
        />
        <div>
          <Checkbox disabled={false} label={'Activate'} />
        </div>
      </div>
    );
  }
}
