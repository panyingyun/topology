import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { TableData } from '../types';
import { tableService } from '../services/tableService';
import { DataTable } from '../components/DataTable';

export const DataViewer: React.FC = () => {
  const { connectionId, tableName } = useParams<{ connectionId: string; tableName: string }>();
  const [data, setData] = useState<TableData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (connectionId && tableName) {
      loadTableData();
    }
  }, [connectionId, tableName]);

  const loadTableData = async () => {
    if (!connectionId || !tableName) return;
    setLoading(true);
    try {
      const tableData = await tableService.getTableData(connectionId, tableName, 100, 0);
      setData(tableData);
    } catch (error) {
      console.error('Failed to load table data:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-[#616289]">Loading...</div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-[#616289]">No data available</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <DataTable
        data={{
          columns: data.columns,
          rows: data.rows,
          rowCount: data.totalRows,
        }}
        onUpdate={(rowIndex, column, value) => {
          console.log('Update:', { rowIndex, column, value });
        }}
        onAddRow={() => {
          console.log('Add row');
        }}
        onDeleteRow={(rowIndex) => {
          console.log('Delete row:', rowIndex);
        }}
        onCommit={async () => {
          console.log('Commit changes');
          await loadTableData();
        }}
      />
    </div>
  );
};
