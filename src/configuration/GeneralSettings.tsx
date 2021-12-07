import React, { FormEvent, MouseEvent } from 'react';
import { Button, InlineField, Input } from '@grafana/ui';
import { OptionsChange } from 'types';

export const GeneralSettings: React.FC<OptionsChange> = ({ options, onChange }) => {
  const hasToken = options.secureJsonFields?.grafanaApiToken;

  const onUrlChangeFactory = () => (event: FormEvent<HTMLInputElement>) => {
    const newJsonData = { ...options.jsonData };
    newJsonData.grafanaUrl = event.currentTarget.value;

    onChange({
      ...options,
      jsonData: newJsonData,
    });
  };

  const onTokenChangeFactory = () => (event: FormEvent<HTMLInputElement>) => {
    const newSecureJsonData = { ...options.secureJsonData };
    newSecureJsonData.grafanaApiToken = event.currentTarget.value;

    onChange({
      ...options,
      secureJsonData: newSecureJsonData,
    });
  };

  const onTokenResetFactory = () => (event: MouseEvent<HTMLButtonElement>) => {
    const newSecureJsonFields = { ...options.secureJsonFields };
    newSecureJsonFields['grafanaApiToken'] = false;

    onChange({
      ...options,
      secureJsonFields: newSecureJsonFields,
    });
  };

  return (
    <div className="gf-form-group">
      <h3 className="page-heading">General</h3>

      <InlineField label="Grafana URL" labelWidth={20} tooltip="The URL to access the local Grafana instance.">
        <Input
          className="width-20"
          placeholder="http://localhost:3000/"
          value={options.jsonData.grafanaUrl}
          onChange={onUrlChangeFactory()}
        />
      </InlineField>

      <InlineField label="Grafana API-Token" labelWidth={20} tooltip="Valid API token for the local Grafana instance.">
        {hasToken ? (
          <>
            <Input type="text" disabled value="configured" width={24} />
            <Button variant="secondary" onClick={onTokenResetFactory()} style={{ marginLeft: 4 }}>
              Reset
            </Button>
          </>
        ) : (
          <Input
            className="width-20"
            placeholder="eyJrIjoidmhCY1A1R0pSWWxtbzcycm5lVng3YTQ5bmUzdXlaVGwiLCJuIjoidGVzdCIsImlkIjoxfQ=="
            value={options.secureJsonData?.grafanaApiToken}
            onChange={onTokenChangeFactory()}
          />
        )}
      </InlineField>
    </div>
  );
};
