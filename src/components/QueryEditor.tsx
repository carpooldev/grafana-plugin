import React, { ChangeEvent } from 'react';
import { InlineField, Input, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { CarpoolDataSourceOptions, CarpoolQuery, QueryType, QueryTypes } from '../types';

type Props = QueryEditorProps<DataSource, CarpoolQuery, CarpoolDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onProgramIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, payload: { ...query.payload, programId: event.target.value } });
    onRunQuery();
  };
  const onIxNameChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, payload: { ...query.payload, instructionName: event.target.value } });
    onRunQuery();
  };
  const onQueryTypeChange = (event: SelectableValue<string>) => {
    onChange({ ...query, payload: { ...query.payload, queryType: event.value as QueryType } });
    onRunQuery();
  };

  const types = QueryTypes;
  const { payload } = query;

  return (
    <div className="gf-form">
      <InlineField label="Query Type" labelWidth={16} tooltip="Query type, currently Basic Invocations, Signers">
        <Select onChange={onQueryTypeChange} options={types} value={payload?.queryType || QueryType.ProgramInvocations} />
      </InlineField>
      <InlineField label="Program Id" labelWidth={10}>
        <Input onChange={onProgramIdChange} value={payload?.programId} width={16} />
      </InlineField>
      <InlineField label="Instruction Name" tooltip="Optional Name of instruction to filter by." labelWidth={20}>
        <Input onChange={onIxNameChange} value={payload?.instructionName} width={16} />
      </InlineField>
    </div>
  );
}
