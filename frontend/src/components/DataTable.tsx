import React, { useState } from 'react';
import { QueryResult } from '../types';

interface DataTableProps {
  data: QueryResult;
  queryText?: string;
  onUpdate?: (rowIndex: number, column: string, value: any) => void;
  onAddRow?: () => void;
  onDeleteRow?: (rowIndex: number) => void;
  onCommit?: () => void;
}

export const DataTable: React.FC<DataTableProps> = ({
  data,
  queryText,
  onUpdate,
  onAddRow,
  onDeleteRow,
  onCommit,
}) => {
  const [editingCell, setEditingCell] = useState<{ row: number; col: string } | null>(null);
  const [editedData, setEditedData] = useState<Record<string, any>[]>(data.rows);

  const handleCellChange = (rowIndex: number, column: string, value: any) => {
    const newData = [...editedData];
    if (!newData[rowIndex]) {
      newData[rowIndex] = {};
    }
    newData[rowIndex][column] = value;
    setEditedData(newData);
    onUpdate?.(rowIndex, column, value);
  };

  const handleCommit = () => {
    onCommit?.();
  };

  return (
    <div className="h-full border-t border-[#E2E8F0] dark:border-gray-800 bg-white/40 dark:bg-black/30 overflow-hidden flex flex-col">
      <div className="px-6 h-12 flex items-center justify-between bg-white/60 dark:bg-white/5 border-b border-[#E2E8F0] dark:border-gray-800">
        <div className="flex items-center gap-3">
          <span className="text-[10px] font-bold uppercase tracking-widest text-[#616289]">Data View</span>
          <div className="flex items-center gap-1 bg-primary/10 px-2 py-0.5 rounded text-[10px] text-primary font-bold">
            <span>LIVE</span>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 border-r border-gray-200 dark:border-gray-700 pr-3">
            {onAddRow && (
              <button
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white dark:bg-gray-900 border border-[#E2E8F0] dark:border-gray-800 text-[11px] font-bold text-[#616289] hover:text-primary hover:border-primary/30 transition-all"
                onClick={onAddRow}
              >
                <span className="material-symbols-outlined text-[16px]">add_row_below</span>
                Add Row
              </button>
            )}
            {onDeleteRow && (
              <button
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white dark:bg-gray-900 border border-[#E2E8F0] dark:border-gray-800 text-[11px] font-bold text-red-500 hover:bg-red-50 hover:border-red-200 transition-all"
                onClick={() => editingCell && onDeleteRow(editingCell.row)}
              >
                <span className="material-symbols-outlined text-[16px]">delete</span>
                Delete Row
              </button>
            )}
          </div>
          {onCommit && (
            <button
              className="flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-[#6366F1] text-white text-[11px] font-bold shadow-lg shadow-indigo-500/30 hover:brightness-110 transition-all"
              onClick={handleCommit}
            >
              <span className="material-symbols-outlined text-[16px]">publish</span>
              Commit Changes
            </button>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-auto p-0 custom-scrollbar">
        <table className="w-full text-left text-xs border-collapse">
          <thead className="text-gray-400 bg-white/30 dark:bg-black/10 sticky top-0 backdrop-blur-sm z-20">
            <tr>
              {data.columns.map((col) => (
                <th key={col} className="py-3 px-6 font-medium border-b border-gray-100 dark:border-gray-800">
                  {col}
                </th>
              ))}
              {(onUpdate || onDeleteRow) && (
                <th className="py-3 px-6 font-medium border-b border-gray-100 dark:border-gray-800 w-20 text-center">
                  Action
                </th>
              )}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50 dark:divide-gray-800">
            {editedData.map((row, rowIndex) => (
              <tr
                key={rowIndex}
                className={`hover:bg-primary/5 transition-colors group ${
                  editingCell?.row === rowIndex ? 'editing-row' : ''
                }`}
              >
                {data.columns.map((col) => {
                  const isEditing = editingCell?.row === rowIndex && editingCell?.col === col;
                  const value = row[col] || '';

                  return (
                    <td
                      key={col}
                      className={`py-2 px-6 font-mono relative editable-cell ${
                        isEditing ? 'ring-2 ring-primary/50 bg-white dark:bg-gray-900 border-primary rounded-sm' : ''
                      }`}
                      onClick={() => setEditingCell({ row: rowIndex, col })}
                    >
                      {isEditing ? (
                        <input
                          className="w-full bg-transparent border-none p-0 focus:ring-0 text-xs font-mono"
                          type="text"
                          value={value}
                          onChange={(e) => handleCellChange(rowIndex, col, e.target.value)}
                          onBlur={() => setEditingCell(null)}
                          autoFocus
                        />
                      ) : (
                        <div className="w-full bg-transparent border-none p-0 text-xs font-mono">
                          {typeof value === 'string' && value.includes('$') ? (
                            <span className="font-bold">{value}</span>
                          ) : (
                            value
                          )}
                        </div>
                      )}
                    </td>
                  );
                })}
                {(onUpdate || onDeleteRow) && (
                  <td className="py-2 px-6 text-center">
                    <button
                      className="w-8 h-8 rounded-lg flex items-center justify-center text-gray-400 hover:text-primary hover:bg-primary/10 transition-all mx-auto"
                      onClick={() => setEditingCell({ row: rowIndex, col: data.columns[0] })}
                    >
                      <span className="material-symbols-outlined text-[18px]">edit_square</span>
                    </button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="h-10 shrink-0 flex items-center justify-between px-6 bg-white dark:bg-black/40 border-t border-[#E2E8F0] dark:border-gray-800 text-[10px] font-medium text-[#616289]">
        <div className="flex items-center gap-2 overflow-hidden mr-4">
          <span className="material-symbols-outlined text-[14px] shrink-0">terminal</span>
          <span className="truncate italic opacity-70">Query executed successfully</span>
        </div>
        <div className="flex items-center gap-4 shrink-0">
          <div className="flex items-center gap-2 text-emerald-status">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-status animate-pulse"></span>
            <span className="uppercase tracking-widest font-bold">Ready</span>
          </div>
          <div className="h-3 w-px bg-gray-200 dark:bg-gray-700"></div>
          <span className="flex items-center gap-1">
            <span className="font-bold text-primary">{data.rowCount.toLocaleString()}</span> records found
          </span>
        </div>
      </div>
    </div>
  );
};
