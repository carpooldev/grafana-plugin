import React, { ChangeEvent, useState } from 'react';
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
  const types = QueryTypes;
  const onQueryTypeChange = (event: SelectableValue<string>) => {
    onChange({ ...query, payload: { ...query.payload, queryType: event.value as QueryType } });
    onRunQuery();
    const selectedType = findType(event.value as QueryType);
    setType(selectedType);
  };


  const findType = (t: QueryType) => {
    return types.find((type) => type.value === t);
  };

  const { payload } = query;
  const [selectedType, setType] = useState(findType(payload?.queryType || QueryType.ProgramInvocations));

  return (
    <div className="gf-form">
      <InlineField label="Query Type" labelWidth={16} tooltip="Query type, currently Basic Invocations, Signers">
        <Select onChange={onQueryTypeChange} options={types} value={selectedType} />
      </InlineField>
      {selectedType?.fields.includes('programId') &&
        <InlineField label="Program Id" labelWidth={10}>
          <Input onChange={onProgramIdChange} value={payload?.programId} width={16} />
        </InlineField>
      }

      {selectedType?.fields.includes('instructionName') &&
        <InlineField label="Instruction Name" tooltip="Optional Name of instruction to filter by." labelWidth={20}>
          <Input onChange={onIxNameChange} value={payload?.instructionName} width={16} />
        </InlineField>
      }
    </div >
  );
}
