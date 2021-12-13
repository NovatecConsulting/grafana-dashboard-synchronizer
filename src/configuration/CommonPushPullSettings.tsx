import React, { FormEvent } from 'react';
import { InlineField, InlineSwitch, Input } from '@grafana/ui';
import { PushPullOptionsChange } from 'types';

export const CommonPushPullSettings: React.FC<PushPullOptionsChange> = ({ ppOptions, onChange }) => {
  const onEnableChangeFactory = () => (event: FormEvent<HTMLInputElement>) => {
    const newOptions = { ...ppOptions };
    newOptions.enable = event.currentTarget.checked;
    onChange(newOptions);
  };

  const onBranchChangeFactory = () => (event: FormEvent<HTMLInputElement>) => {
    const newOptions = { ...ppOptions };
    newOptions.gitBranch = event.currentTarget.value;
    onChange(newOptions);
  };

  const onIntervalChangeFactory = () => (event: FormEvent<HTMLInputElement>) => {
    const newOptions = { ...ppOptions };
    newOptions.syncInterval = parseInt(event.currentTarget.value, 10);
    onChange(newOptions);
  };

  const onFilterChangeFactory = () => (event: FormEvent<HTMLInputElement>) => {
    const newOptions = { ...ppOptions };
    newOptions.filter = event.currentTarget.value;
    onChange(newOptions);
  };

  return (
    <>
      <InlineField label="Enable" labelWidth={20}>
        <InlineSwitch value={ppOptions.enable} onChange={onEnableChangeFactory()} />
      </InlineField>

      {ppOptions.enable && (
        <>
          <InlineField label="Branch Name" labelWidth={20} tooltip="Name of the target branch.">
            <Input
              className="width-20"
              placeholder="main"
              value={ppOptions.gitBranch}
              onChange={onBranchChangeFactory()}
            />
          </InlineField>

          <InlineField
            label="Interval"
            labelWidth={20}
            tooltip="The interval (seconds) in which this operation should be triggered."
          >
            <Input
              className="width-20"
              placeholder="60"
              type="number"
              value={ppOptions.syncInterval}
              onChange={onIntervalChangeFactory()}
            />
          </InlineField>

          <InlineField
            hidden={true}
            label="Filter"
            labelWidth={20}
            tooltip="A filter (regular expression) to narrow down the dashboards considered. This filter represents a whitelisting that takes into account the dashboard name and folder."
          >
            <Input className="width-20" placeholder=".*" value={ppOptions.filter} onChange={onFilterChangeFactory()} />
          </InlineField>
        </>
      )}
    </>
  );
};
