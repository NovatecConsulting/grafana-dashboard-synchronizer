import React, { FormEvent, MouseEvent } from 'react';
import { Button, InlineField, Input, TextArea } from '@grafana/ui';
import { OptionsChange } from 'types';

export const GitSettings: React.FC<OptionsChange> = ({ options, onChange }) => {
  const hasSshKey = !!options.secureJsonFields?.privateSshKey;

  const onUrlChangeFactory = () => (event: FormEvent<HTMLInputElement>) => {
    const newJsonData = { ...options.jsonData };
    newJsonData.gitUrl = event.currentTarget.value;

    onChange({
      ...options,
      jsonData: newJsonData,
    });
  };

  const onKeyChangeFactory = () => (event: FormEvent<HTMLTextAreaElement>) => {
    const newSecureJsonData = { ...options.secureJsonData };
    newSecureJsonData.privateSshKey = event.currentTarget.value;

    onChange({
      ...options,
      secureJsonData: newSecureJsonData,
    });
  };

  const onKeyResetFactory = () => (event: MouseEvent<HTMLButtonElement>) => {
    const newSecureJsonFields = { ...options.secureJsonFields };
    newSecureJsonFields['privateSshKey'] = false;

    onChange({
      ...options,
      secureJsonFields: newSecureJsonFields,
    });
  };

  return (
    <div className="gf-form-group">
      <h3 className="page-heading">Git</h3>

      <InlineField
        label="Repository URL"
        labelWidth={20}
        tooltip="SSH URL of the Git repository that will be used to store and retrieve the dashboards."
      >
        <Input
          className="width-20"
          placeholder="http://"
          value={options.jsonData.gitUrl}
          onChange={onUrlChangeFactory()}
        />
      </InlineField>

      <InlineField label="SSH Key" labelWidth={20} tooltip="Private SSH key to access the Git repository.">
        {hasSshKey ? (
          <>
            <Input type="text" disabled value="configured" width={24} />
            <Button variant="secondary" onClick={onKeyResetFactory()} style={{ marginLeft: 4 }}>
              Reset
            </Button>
          </>
        ) : (
          <TextArea
            rows={7}
            onChange={onKeyChangeFactory()}
            placeholder="Begins with -----BEGIN RSA PRIVATE KEY-----"
            required
          />
        )}
      </InlineField>
    </div>
  );
};
