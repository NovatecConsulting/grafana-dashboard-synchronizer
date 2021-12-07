import React, { PureComponent } from 'react';
import { DataSourcePluginOptionsEditorProps, DataSourceSettings } from '@grafana/data';
import { SynchronizeOptions, SecureSynchronizeOptions } from '../types';
import { InfoBox } from '@grafana/ui';
import {} from '@emotion/core';
import { GeneralSettings } from './GeneralSettings';
import { GitSettings } from './GitSettings';
import { PushSettings } from './PushSettings';
import { PullSettings } from './PullSettings';

interface Props extends DataSourcePluginOptionsEditorProps<SynchronizeOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onChange = (newOptions: DataSourceSettings<SynchronizeOptions, SecureSynchronizeOptions>) => {
    const { onOptionsChange, options } = this.props;

    const newJsonData: SynchronizeOptions = {
      ...newOptions.jsonData,
      pushConfiguration: {
        ...newOptions.jsonData.pushConfiguration,
      },
      pullConfiguration: {
        ...newOptions.jsonData.pullConfiguration,
      },
    };

    onOptionsChange({
      ...options,
      jsonData: newJsonData,
      secureJsonData: {
        ...newOptions.secureJsonData,
      },
      secureJsonFields: {
        ...newOptions.secureJsonFields,
      },
    });
  };

  render() {
    const { options } = this.props;

    return (
      <>
        <div className="gf-form-group">
          <InfoBox
            title="Important"
            severity="info"
            url="https://github.com/NovatecConsulting/grafana-dashboard-sync-plugin"
          >
            This is <strong>not</strong> a normal data source and it is <strong>not intended </strong>to provide any
            kind of data but can be considered more of a generic plugin. The purpose of this data source is that it can
            be used to push dashboards to an external Git repository or to import/update the dashboards within it to the
            local Grafana instance. For more information, read the <code>README</code> in the related Github repository.
          </InfoBox>

          <GeneralSettings onChange={this.onChange} options={options} />

          <GitSettings onChange={this.onChange} options={options} />

          <PushSettings onChange={this.onChange} options={options} />

          <PullSettings onChange={this.onChange} options={options} />
        </div>
      </>
    );
  }
}
