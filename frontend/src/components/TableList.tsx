import React from 'react';
import { Table } from '../types';

interface TableListProps {
  tables: Table[];
  selectedTable?: string;
  onTableSelect: (tableName: string) => void;
}

export const TableList: React.FC<TableListProps> = ({
  tables,
  selectedTable,
  onTableSelect,
}) => {
  return (
    <div className="space-y-1">
      {tables.map((table) => (
        <div
          key={table.name}
          className={`flex items-center gap-2 px-3 py-2 ml-3 rounded-lg cursor-pointer text-sm group transition-all ${
            selectedTable === table.name
              ? 'bg-primary/5 text-primary border border-primary/10'
              : 'hover:bg-black/5 dark:hover:bg-white/5'
          }`}
          onClick={() => onTableSelect(table.name)}
        >
          <span className="material-symbols-outlined text-[18px] text-gray-400 group-hover:text-primary">
            {table.type === 'view' ? 'view_list' : 'table_chart'}
          </span>
          <span className={selectedTable === table.name ? 'font-medium' : 'text-gray-700 dark:text-gray-300'}>
            {table.name}
          </span>
          {table.rowCount !== undefined && (
            <span className="ml-auto text-[10px] text-gray-400">
              {table.rowCount.toLocaleString()}
            </span>
          )}
        </div>
      ))}
    </div>
  );
};
