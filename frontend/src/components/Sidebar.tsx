import React from 'react';
import { Connection, Table } from '../types';

interface SidebarProps {
  connections: Connection[];
  selectedConnectionId?: string;
  selectedTable?: string;
  tables: Table[];
  onConnectionSelect: (connectionId: string) => void;
  onTableSelect: (tableName: string) => void;
  onNewConnection: () => void;
  onNewTable: () => void;
  onSettings: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
  connections,
  selectedConnectionId,
  selectedTable,
  tables,
  onConnectionSelect,
  onTableSelect,
  onNewConnection,
  onSettings,
}) => {
  const groupedConnections = connections.reduce((acc, conn) => {
    const group = conn.group || 'Other';
    if (!acc[group]) {
      acc[group] = [];
    }
    acc[group].push(conn);
    return acc;
  }, {} as Record<string, Connection[]>);

  return (
    <aside className="glass-sidebar w-72 flex flex-col z-20">
      <div className="p-6 flex items-center gap-3">
        <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-white shadow-lg shadow-primary/20">
          <span className="material-symbols-outlined">database</span>
        </div>
        <div>
          <h1 className="font-bold text-lg leading-tight tracking-tight">DB Studio</h1>
          <p className="text-[10px] text-[#616289] font-semibold dark:text-gray-400 uppercase tracking-widest">Enterprise</p>
        </div>
      </div>

      <div className="px-6 py-4">
        <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl flex items-center px-4 h-11">
          <span className="material-symbols-outlined text-gray-400 text-[20px]">search</span>
          <input
            className="bg-transparent border-none focus:ring-0 text-sm w-full placeholder:text-gray-400"
            placeholder="Search databases..."
            type="text"
          />
        </div>
      </div>

      <nav className="flex-1 overflow-y-auto px-4 custom-scrollbar">
        <div className="space-y-6 pt-6">
          {Object.entries(groupedConnections).map(([group, conns]) => (
            <div key={group}>
              <div className="flex items-center justify-between px-2 mb-2">
                <span className="text-[11px] font-bold uppercase tracking-wider text-[#616289]">{group}</span>
                <span className="material-symbols-outlined text-[16px] cursor-pointer text-[#616289] hover:text-primary transition-colors">expand_more</span>
              </div>
              <div className="space-y-1">
                {conns.map((conn) => (
                  <div key={conn.id}>
                    <div
                      className={`flex items-center gap-3 px-3 py-2.5 rounded-xl cursor-pointer group transition-all ${
                        selectedConnectionId === conn.id
                          ? 'bg-primary/10 text-primary border border-primary/20'
                          : 'hover:bg-black/5 dark:hover:bg-white/5'
                      }`}
                      onClick={() => onConnectionSelect(conn.id)}
                    >
                      <span className="material-symbols-outlined text-[20px]">
                        {conn.type === 'postgresql' ? 'storage' : conn.type === 'mysql' ? 'table_chart' : 'database'}
                      </span>
                      <span className="text-sm font-semibold flex-1">{conn.name}</span>
                      <div className={`w-2 h-2 rounded-full ${
                        conn.status === 'connected' 
                          ? 'bg-emerald-status shadow-[0_0_8px_#10B981]' 
                          : 'bg-gray-400 opacity-50'
                      }`}></div>
                    </div>
                    {selectedConnectionId === conn.id && tables.length > 0 && (
                      <div className="mt-2 ml-4 border-l border-gray-200 dark:border-gray-800 space-y-1">
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
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </nav>

      <div className="p-6 border-t border-[#E2E8F0] dark:border-gray-800">
        <button
          className="w-full flex items-center justify-center gap-2 p-3 rounded-xl neumorphic-card text-xs font-bold text-[#616289] hover:text-primary transition-all"
          onClick={onSettings}
        >
          <span className="material-symbols-outlined text-[18px]">settings</span>
          Settings
        </button>
      </div>
    </aside>
  );
};
