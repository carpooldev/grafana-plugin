import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { CarpoolDataSourceOptions, CarpoolSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<CarpoolDataSourceOptions> { }

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const onPathChange = (event: ChangeEvent<HTMLInputElement>) => {
    const jsonData = {
      ...options.jsonData,
      url: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  const onBucketsChange = (event: ChangeEvent<HTMLInputElement>) => {
    const jsonData = {
      ...options.jsonData,
      maxBuckets: parseInt(event.target.value, 10),
    };
    onOptionsChange({ ...options, jsonData });
  };


  // Secure field (only sent to the backend)
  const onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        apiKey: event.target.value,
      },
    });
  };

  const onResetAPIKey = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: '',
      },
    });
  };

  const { jsonData, secureJsonFields } = options;
  const secureJsonData = (options.secureJsonData || {}) as CarpoolSecureJsonData;

  return (
    <div className="gf-form-group">
      <InlineField label="Carpool URL" labelWidth={12}>
        <Input
          onChange={onPathChange}
          value={jsonData.url}
          placeholder="Carpool URL"
          width={40}
        />
      </InlineField>
      <InlineField label="Carpool MaxData Points Per Query" labelWidth={12}>
        <Input
          onChange={onBucketsChange}
          value={jsonData.maxBuckets}
          placeholder="Max Buckets"
          width={40}
        />
      </InlineField>
      <InlineField label="API Key" labelWidth={12}>
        <SecretInput
          isConfigured={(secureJsonFields && secureJsonFields.apiKey) as boolean}
          value={secureJsonData.apiKey || ''}
          placeholder="Carpool.dev api key"
          width={40}
          onReset={onResetAPIKey}
          onChange={onAPIKeyChange}
        />
      </InlineField>
    </div>
  );
}
