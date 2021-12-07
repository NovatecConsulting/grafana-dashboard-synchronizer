import React from 'react';
import { OptionsChange, PullOptions } from 'types';
import { CommonPushPullSettings } from './CommonPushPullSettings';

export const PullSettings: React.FC<OptionsChange> = ({ options, onChange }) => {
  const onPullSettingsChange = (newOptions: PullOptions) => {
    const newJsonData = { ...options.jsonData };
    newJsonData.pullConfiguration = newOptions;

    onChange({
      ...options,
      jsonData: newJsonData,
    });
  };

  return (
    <div className="gf-form-group">
      <h3 className="page-heading">Pull Dashboards</h3>

      <CommonPushPullSettings ppOptions={options.jsonData.pullConfiguration || {}} onChange={onPullSettingsChange} />
    </div>
  );
};
