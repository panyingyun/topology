import React from 'react';
import { Connection } from '../types';

interface ConnectionCardProps {
  connection: Connection;
  isSelected: boolean;
  onClick: () => void;
  onDelete?: () => void;
}

export const ConnectionCard: React.FC<ConnectionCardProps> = ({
  connection,
  isSelected,
  onClick,
  onDelete,
}) => {
  return (
    <div
      className={`flex items-center gap-3 px-3 py-2 rounded-xl cursor-pointer transition-all ${
        isSelected
          ? 'bg-primary/10 text-primary border border-primary/20'
          : 'hover:bg-black/5 dark:hover:bg-white/5'
      }`}
      onClick={onClick}
    >
      <span className="material-symbols-outlined text-[20px]">
        {connection.type === 'postgresql' ? 'storage' : 
         connection.type === 'mysql' ? 'table_chart' : 
         connection.type === 'redis' ? 'bolt' : 'schema'}
      </span>
      <span className="text-sm font-medium flex-1">{connection.name}</span>
      <div className={`w-2 h-2 rounded-full ${
        connection.status === 'connected' 
          ? 'bg-emerald-status shadow-[0_0_8px_#10B981]' 
          : 'bg-gray-400 opacity-50'
      }`}></div>
      {onDelete && (
        <button
          className="ml-2 p-1 rounded hover:bg-red-100 dark:hover:bg-red-900/20 text-red-500 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
        >
          <span className="material-symbols-outlined text-[16px]">delete</span>
        </button>
      )}
    </div>
  );
};
