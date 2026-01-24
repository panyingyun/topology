import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { TableSchema, Column, Index, ForeignKey } from '../types';
import { tableService } from '../services/tableService';

export const TableDesigner: React.FC = () => {
  const { connectionId, tableName } = useParams<{ connectionId?: string; tableName?: string }>();
  const [schema, setSchema] = useState<TableSchema>({
    name: tableName || 'new_table',
    columns: [],
    indexes: [],
    foreignKeys: [],
  });
  const [isOpen, setIsOpen] = useState(true);

  useEffect(() => {
    if (connectionId && tableName) {
      loadSchema();
    }
  }, [connectionId, tableName]);

  const loadSchema = async () => {
    if (!connectionId || !tableName) return;
    const loadedSchema = await tableService.getTableSchema(connectionId, tableName);
    setSchema(loadedSchema);
  };

  const handleAddColumn = () => {
    setSchema({
      ...schema,
      columns: [
        ...schema.columns,
        {
          name: `column_${schema.columns.length + 1}`,
          type: 'VARCHAR(255)',
          isPrimaryKey: false,
          isNotNull: false,
          isUnique: false,
        },
      ],
    });
  };

  const handleDeleteColumn = (index: number) => {
    setSchema({
      ...schema,
      columns: schema.columns.filter((_, i) => i !== index),
    });
  };

  const handleUpdateColumn = (index: number, field: keyof Column, value: any) => {
    const newColumns = [...schema.columns];
    newColumns[index] = { ...newColumns[index], [field]: value };
    setSchema({ ...schema, columns: newColumns });
  };

  const handleSave = async () => {
    if (!connectionId) return;
    await tableService.updateTableSchema(connectionId, schema.name, schema);
    alert('Schema saved successfully!');
  };

  if (!isOpen) return null;

  return (
    <div className="flex h-screen overflow-hidden">
      <aside className="glass-sidebar w-72 flex flex-col z-10">
        <div className="p-6 flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-white shadow-lg shadow-primary/20">
            <span className="material-symbols-outlined">database</span>
          </div>
          <div>
            <h1 className="font-bold text-lg leading-tight">Topology</h1>
            <p className="text-xs text-[#616289] dark:text-gray-400">Pro Workspace</p>
          </div>
        </div>

        <div className="px-6 py-4">
          <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl flex items-center px-4 h-11">
            <span className="material-symbols-outlined text-gray-400 text-[20px]">search</span>
            <input
              className="bg-transparent border-none focus:ring-0 text-sm w-full placeholder:text-gray-400"
              placeholder="Search tables..."
              type="text"
            />
          </div>
        </div>

        <nav className="flex-1 overflow-y-auto px-4 custom-scrollbar">
          <div className="space-y-6">
            <div>
              <div className="flex items-center justify-between px-2 mb-2">
                <span className="text-[11px] font-bold uppercase tracking-wider text-[#616289]">ecommerce_prod</span>
                <span className="material-symbols-outlined text-[16px] cursor-pointer">settings</span>
              </div>
              <div className="space-y-1">
                <div className="flex items-center gap-3 px-3 py-2 rounded-xl bg-primary/10 text-primary border border-primary/20">
                  <span className="material-symbols-outlined text-[20px]">table_rows</span>
                  <span className="text-sm font-medium">users</span>
                </div>
                <div className="flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer group">
                  <span className="material-symbols-outlined text-[20px] text-[#616289]">table_rows</span>
                  <span className="text-sm font-medium">orders</span>
                </div>
                <div className="flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer group">
                  <span className="material-symbols-outlined text-[20px] text-[#616289]">table_rows</span>
                  <span className="text-sm font-medium">products</span>
                </div>
              </div>
            </div>
          </div>
        </nav>

        <div className="p-4 border-t border-[#E2E8F0] dark:border-gray-800">
          <button className="w-full bg-primary text-white font-semibold py-3 rounded-xl shadow-lg shadow-primary/30 flex items-center justify-center gap-2 hover:brightness-110 transition-all">
            <span className="material-symbols-outlined text-[20px]">add_box</span>
            New Table
          </button>
        </div>
      </aside>

      <main className="flex-1 overflow-y-auto bg-background-light dark:bg-background-dark p-10">
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center gap-2 text-sm">
            <a className="text-[#616289] hover:text-primary transition-colors" href="#">ecommerce_prod</a>
            <span className="text-[#616289]">/</span>
            <span className="font-semibold text-[#111118] dark:text-white">Tables</span>
          </div>
        </div>

        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 lg:p-8 modal-overlay bg-black/40 backdrop-blur-sm">
          <div className="bg-[#FFFFFF] dark:bg-[#101122] w-full max-w-5xl max-h-[90vh] flex flex-col rounded-[24px] border border-[#E2E8F0] dark:border-[#1e293b] shadow-2xl overflow-hidden">
            <div className="px-8 py-6 border-b border-[#E2E8F0] dark:border-white/5 flex items-center justify-between bg-white/50 dark:bg-white/5 backdrop-blur-sm">
              <div>
                <div className="flex items-center gap-3 mb-1">
                  <span className="material-symbols-outlined text-primary text-2xl font-bold">table_rows</span>
                  <h2 className="text-2xl font-bold text-[#111118] dark:text-white tracking-tight">Table Designer</h2>
                </div>
                <p className="text-sm text-[#616289] dark:text-gray-400">
                  Editing <span className="font-mono bg-primary/10 text-primary px-2 py-0.5 rounded">{schema.name}</span> schema
                </p>
              </div>
              <button
                className="w-10 h-10 rounded-full hover:bg-black/5 dark:hover:bg-white/5 flex items-center justify-center transition-colors"
                onClick={() => setIsOpen(false)}
              >
                <span className="material-symbols-outlined text-[#616289]">close</span>
              </button>
            </div>

            <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
              <div className="border border-[#E2E8F0] dark:border-white/5 rounded-2xl overflow-hidden mb-8">
                <div className="overflow-x-auto">
                  <table className="w-full text-left">
                    <thead className="bg-gray-50 dark:bg-white/5">
                      <tr>
                        <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-wider text-[#616289]">Name</th>
                        <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-wider text-[#616289]">Type</th>
                        <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-wider text-[#616289]">PK</th>
                        <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-wider text-[#616289]">Not Null</th>
                        <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-wider text-[#616289]">Unique</th>
                        <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-wider text-[#616289]">Default</th>
                        <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-wider text-[#616289]"></th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-black/5 dark:divide-white/5">
                      {schema.columns.map((column, index) => (
                        <tr key={index} className="hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors">
                          <td className="px-6 py-3">
                            <input
                              className="bg-transparent border-none focus:ring-1 focus:ring-primary rounded text-sm font-medium w-full"
                              type="text"
                              value={column.name}
                              onChange={(e) => handleUpdateColumn(index, 'name', e.target.value)}
                            />
                          </td>
                          <td className="px-6 py-3">
                            <select
                              className="bg-transparent border-none focus:ring-1 focus:ring-primary rounded text-sm w-full"
                              value={column.type}
                              onChange={(e) => handleUpdateColumn(index, 'type', e.target.value)}
                            >
                              <option>BIGINT (AUTO_INC)</option>
                              <option>INT</option>
                              <option>VARCHAR(255)</option>
                              <option>TEXT</option>
                              <option>DECIMAL(10,2)</option>
                              <option>TIMESTAMP</option>
                              <option>BOOLEAN</option>
                            </select>
                          </td>
                          <td className="px-6 py-3">
                            <input
                              className="rounded text-primary focus:ring-primary h-4 w-4"
                              type="checkbox"
                              checked={column.isPrimaryKey}
                              onChange={(e) => handleUpdateColumn(index, 'isPrimaryKey', e.target.checked)}
                            />
                          </td>
                          <td className="px-6 py-3">
                            <input
                              className="rounded text-primary focus:ring-primary h-4 w-4"
                              type="checkbox"
                              checked={column.isNotNull}
                              onChange={(e) => handleUpdateColumn(index, 'isNotNull', e.target.checked)}
                            />
                          </td>
                          <td className="px-6 py-3">
                            <input
                              className="rounded text-primary focus:ring-primary h-4 w-4"
                              type="checkbox"
                              checked={column.isUnique}
                              onChange={(e) => handleUpdateColumn(index, 'isUnique', e.target.checked)}
                            />
                          </td>
                          <td className="px-6 py-3">
                            <input
                              className="bg-transparent border-none focus:ring-1 focus:ring-primary rounded text-xs w-full"
                              type="text"
                              value={column.defaultValue || ''}
                              onChange={(e) => handleUpdateColumn(index, 'defaultValue', e.target.value)}
                              placeholder="NULL"
                            />
                          </td>
                          <td className="px-6 py-3 text-right">
                            <button
                              className="text-[#616289] hover:text-red-500"
                              onClick={() => handleDeleteColumn(index)}
                            >
                              <span className="material-symbols-outlined text-sm">delete</span>
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <div className="p-3 bg-gray-50/50 dark:bg-white/5 border-t border-black/5 dark:border-white/5 flex justify-center">
                  <button
                    className="flex items-center gap-2 text-primary font-bold text-sm hover:underline"
                    onClick={handleAddColumn}
                  >
                    <span className="material-symbols-outlined">add</span>
                    Add New Column
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="p-6 rounded-2xl border border-[#E2E8F0] dark:border-white/5 bg-gray-50/30 dark:bg-white/2">
                  <h3 className="font-bold mb-4 flex items-center gap-2">
                    <span className="material-symbols-outlined text-primary">key</span>
                    Indexes & Keys
                  </h3>
                  <div className="space-y-3">
                    {schema.indexes.map((index, idx) => (
                      <div key={idx} className="flex items-center justify-between p-3 neumorphic-inset rounded-xl bg-background-light dark:bg-background-dark">
                        <div className="flex items-center gap-3">
                          <span className="material-symbols-outlined text-emerald-status text-lg">check_circle</span>
                          <div>
                            <p className="text-sm font-semibold">{index.name}</p>
                            <p className="text-[10px] text-[#616289]">Columns: {index.columns.join(', ')}</p>
                          </div>
                        </div>
                        <span className="material-symbols-outlined text-sm text-[#616289] cursor-pointer hover:text-primary">edit</span>
                      </div>
                    ))}
                    {schema.indexes.length === 0 && (
                      <p className="text-xs text-[#616289] text-center py-4">No indexes defined</p>
                    )}
                  </div>
                </div>

                <div className="p-6 rounded-2xl border border-[#E2E8F0] dark:border-white/5 bg-gray-50/30 dark:bg-white/2">
                  <h3 className="font-bold mb-4 flex items-center gap-2">
                    <span className="material-symbols-outlined text-primary">link</span>
                    Foreign Keys
                  </h3>
                  <div className="flex flex-col items-center justify-center h-28 border-2 border-dashed border-gray-200 dark:border-gray-800 rounded-xl">
                    <span className="material-symbols-outlined text-gray-400 mb-2">add_link</span>
                    <p className="text-xs text-[#616289]">No relations defined</p>
                    <button className="mt-2 text-[10px] font-bold text-primary uppercase tracking-wider">Add Foreign Key</button>
                  </div>
                </div>
              </div>
            </div>

            <div className="px-8 py-6 border-t border-[#E2E8F0] dark:border-white/5 flex items-center justify-between bg-gray-50 dark:bg-white/2">
              <div className="flex gap-3">
                <button
                  className="px-6 py-2.5 rounded-xl border border-[#E2E8F0] dark:border-white/10 font-semibold text-[#616289] hover:bg-white dark:hover:bg-white/5 transition-all flex items-center gap-2"
                  onClick={() => setIsOpen(false)}
                >
                  Cancel
                </button>
              </div>
              <div className="flex gap-3">
                <button className="px-6 py-2.5 rounded-xl neumorphic-card font-semibold text-[#616289] flex items-center gap-2 hover:brightness-105">
                  <span className="material-symbols-outlined text-sm">code</span>
                  SQL Preview
                </button>
                <button
                  className="px-8 py-2.5 rounded-xl bg-primary text-white font-bold shadow-xl shadow-primary/25 hover:shadow-primary/40 transition-all flex items-center gap-2"
                  onClick={handleSave}
                >
                  <span className="material-symbols-outlined text-sm">save</span>
                  Apply Changes
                </button>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};
