import React, { FormEvent } from 'react';
import { InlineField, InlineSwitch, Input } from '@grafana/ui';
import { OptionsChange, PushOptions } from 'types';
import { CommonPushPullSettings } from './CommonPushPullSettings';

export const PushSettings: React.FC<OptionsChange> = ({ options, onChange }) => {
  const pushOptions = options.jsonData?.pushConfiguration || {};
  const isEnable = pushOptions.enable;

  const onPushSettingsChange = (newOptions: PushOptions) => {
    const newJsonData = { ...options.jsonData };
    newJsonData.pushConfiguration = newOptions;

    onChange({
      ...options,
      jsonData: newJsonData,
    });
  };

  const onPushTagsChangeFactory = () => (event: FormEvent<HTMLInputElement>) => {
    const newPushOptions = { ...pushOptions };
    newPushOptions.pushTags = event.currentTarget.checked;
    onPushSettingsChange(newPushOptions);
  };

  const onSelectorChangeFactory = () => (event: FormEvent<HTMLInputElement>) => {
    const newPushOptions = { ...pushOptions };
    newPushOptions.tagPattern = event.currentTarget.value;
    onPushSettingsChange(newPushOptions);
  };

  return (
    <div className="gf-form-group">
      <h3 className="page-heading">Push Dashboards</h3>

      <CommonPushPullSettings ppOptions={options.jsonData.pushConfiguration || {}} onChange={onPushSettingsChange} />

      {isEnable && (
        <>
          <InlineField
            label="Selector-Tag"
            labelWidth={20}
            tooltip="Defines the tag to match in order to synchronize the associated dashboard."
          >
            <Input
              className="width-20"
              placeholder="sync"
              value={pushOptions.tagPattern}
              onChange={onSelectorChangeFactory()}
            />
          </InlineField>

          <InlineField
            label="Push Tags"
            labelWidth={20}
            tooltip="Defines whether the selector tag should be pushed to the Git repository. Otherwise, the tag is removed from the dashboard before pushing."
          >
            <InlineSwitch value={pushOptions.pushTags} onChange={onPushTagsChangeFactory()} />
          </InlineField>
        </>
      )}
    </div>
  );
};
