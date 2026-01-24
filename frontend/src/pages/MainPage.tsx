import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { QueryResult, OptimizedSQL } from '../types';
import { queryService } from '../services/queryService';
import { DataTable } from '../components/DataTable';
import { highlightSQL } from '../utils/sqlHighlighter';

export const MainPage: React.FC = () => {
  const navigate = useNavigate();
  const [sql, setSql] = useState('SELECT * FROM transactions LIMIT 1000;');
  const [result, setResult] = useState<QueryResult | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);
  const [optimizedSQL, setOptimizedSQL] = useState<OptimizedSQL | null>(null);
  const [isOptimizing, setIsOptimizing] = useState(false);
  const [showOptimized, setShowOptimized] = useState(false);
  const [selectedTable, setSelectedTable] = useState('transactions');
  const [isExpanded, setIsExpanded] = useState(true);
  const editorRef = useRef<HTMLTextAreaElement>(null);

  // Load initial data
  useEffect(() => {
    loadInitialData();
  }, []);

  const loadInitialData = async () => {
    try {
      const queryResult = await queryService.executeQuery('1', sql);
      setResult(queryResult);
    } catch (error) {
      console.error('Failed to load initial data:', error);
    }
  };

  const handleExecute = async () => {
    if (!sql.trim()) return;

    setIsExecuting(true);
    try {
      const queryResult = await queryService.executeQuery('1', sql);
      setResult(queryResult);
      setShowOptimized(false);
    } catch (error) {
      setResult({
        columns: [],
        rows: [],
        rowCount: 0,
        error: error instanceof Error ? error.message : 'Unknown error',
      });
    } finally {
      setIsExecuting(false);
    }
  };

  const handleOptimize = async () => {
    if (!sql.trim()) return;

    setIsOptimizing(true);
    try {
      const optimized = await queryService.optimizeSQL(sql);
      setOptimizedSQL(optimized);
      setShowOptimized(true);
    } catch (error) {
      console.error('Optimization error:', error);
    } finally {
      setIsOptimizing(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      handleExecute();
    }
  };

  const handleNewConnection = () => {
    navigate('/connections');
  };

  const handleSettings = () => {
    console.log('Settings clicked');
  };

  const handleRefresh = () => {
    loadInitialData();
  };

  // Default mock data if no result
  const displayData = result || {
    columns: ['id', 'user_id', 'amount', 'status', 'created_at'],
    rows: [
      {
        id: 'TXN_8841',
        user_id: 'USR_2291',
        amount: '$1,250.00',
        status: 'COMPLETED',
        created_at: '2023-10-24 14:22:10',
      },
      {
        id: 'TXN_8842',
        user_id: 'USR_4102',
        amount: '$745.50',
        status: 'COMPLETED',
        created_at: '2023-10-24 14:15:05',
      },
      {
        id: 'TXN_8843',
        user_id: 'USR_8812',
        amount: '$920.00',
        status: 'PENDING',
        created_at: '2023-10-24 14:12:44',
      },
      {
        id: 'TXN_8844',
        user_id: 'USR_3301',
        amount: '$2,100.00',
        status: 'COMPLETED',
        created_at: '2023-10-24 14:10:02',
      },
    ],
    rowCount: 1000,
  };

  return (
    <div className="flex flex-col h-screen overflow-hidden">
      {/* Top Header */}
      <header className="h-16 flex items-center justify-between px-6 glass-header z-30 shrink-0">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-white shadow-lg shadow-primary/20">
            <span className="material-symbols-outlined">database</span>
          </div>
          <div>
            <h1 className="font-bold text-lg leading-tight tracking-tight">Topology</h1>
            <p className="text-[10px] text-[#616289] font-semibold dark:text-gray-400 uppercase tracking-widest">Enterprise</p>
          </div>
        </div>
        <div className="flex items-center gap-6">
          <button
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-white text-xs font-bold shadow-lg shadow-primary/30 hover:brightness-110 transition-all"
            onClick={handleNewConnection}
          >
            <span className="material-symbols-outlined text-[18px]">add</span>
            New Connection
          </button>
          <div className="flex items-center gap-4 border-r border-gray-200 dark:border-gray-800 pr-6">
            <button className="w-10 h-10 rounded-full neumorphic-card flex items-center justify-center text-[#616289] hover:text-primary transition-colors relative group">
              <span className="material-symbols-outlined text-[22px]">notifications</span>
              <span className="absolute top-2.5 right-2.5 w-2 h-2 bg-primary rounded-full border-2 border-[#f6f6f8] dark:border-[#101122]"></span>
            </button>
          </div>
          <div className="flex items-center gap-3 p-1 rounded-full bg-white/40 dark:bg-white/5 border border-white/20 hover:border-primary/30 transition-all cursor-pointer group pr-4">
            <div className="w-9 h-9 rounded-full bg-primary/20 flex items-center justify-center">
              <span className="material-symbols-outlined text-primary">person</span>
            </div>
            <div className="hidden sm:block">
              <p className="text-xs font-bold leading-none">Admin</p>
            </div>
          </div>
        </div>
      </header>

      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar - docs/01 style */}
        <aside className="glass-sidebar w-72 flex flex-col z-20">
          <nav className="flex-1 overflow-y-auto px-4 custom-scrollbar">
            <div className="space-y-6 pt-6">
              <div>
                <div className="flex items-center justify-between px-2 mb-4">
                  <span className="text-[11px] font-bold uppercase tracking-wider text-[#616289]">Connected DBs</span>
                  <span
                    className="material-symbols-outlined text-[16px] cursor-pointer text-[#616289] hover:text-primary transition-colors"
                    onClick={handleRefresh}
                  >
                    refresh
                  </span>
                </div>
                <div className="space-y-4">
                  <div>
                    <div
                      className="flex items-center gap-3 px-3 py-2.5 rounded-xl bg-primary/10 text-primary border border-primary/20 cursor-pointer group"
                      onClick={() => setIsExpanded(!isExpanded)}
                    >
                      <span className="material-symbols-outlined text-[20px]">storage</span>
                      <span className="text-sm font-semibold">Main Postgres</span>
                      <span
                        className={`material-symbols-outlined text-[18px] ml-auto transition-transform ${
                          isExpanded ? '' : '-rotate-90'
                        }`}
                      >
                        expand_more
                      </span>
                    </div>
                    {isExpanded && (
                      <div className="mt-2 ml-4 border-l border-gray-200 dark:border-gray-800 space-y-1">
                        <div
                          className={`flex items-center gap-2 px-3 py-2 ml-3 rounded-lg cursor-pointer text-sm group transition-all ${
                            selectedTable === 'users'
                              ? 'bg-primary/5 text-primary border border-primary/10'
                              : 'hover:bg-black/5 dark:hover:bg-white/5'
                          }`}
                          onClick={() => setSelectedTable('users')}
                        >
                          <span className="material-symbols-outlined text-[18px] text-gray-400 group-hover:text-primary">table_chart</span>
                          <span className={selectedTable === 'users' ? 'font-medium' : 'text-gray-700 dark:text-gray-300'}>users</span>
                        </div>
                        <div
                          className={`flex items-center gap-2 px-3 py-2 ml-3 rounded-lg cursor-pointer text-sm group transition-all ${
                            selectedTable === 'orders'
                              ? 'bg-primary/5 text-primary border border-primary/10'
                              : 'hover:bg-black/5 dark:hover:bg-white/5'
                          }`}
                          onClick={() => setSelectedTable('orders')}
                        >
                          <span className="material-symbols-outlined text-[18px] text-gray-400 group-hover:text-primary">table_chart</span>
                          <span className={selectedTable === 'orders' ? 'font-medium' : 'text-gray-700 dark:text-gray-300'}>orders</span>
                        </div>
                        <div
                          className={`flex items-center gap-2 px-3 py-2 ml-3 rounded-lg cursor-pointer text-sm transition-all ${
                            selectedTable === 'transactions'
                              ? 'bg-primary/5 text-primary border border-primary/10'
                              : 'hover:bg-black/5 dark:hover:bg-white/5'
                          }`}
                          onClick={() => setSelectedTable('transactions')}
                        >
                          <span className="material-symbols-outlined text-[18px]">table_chart</span>
                          <span className={selectedTable === 'transactions' ? 'font-medium' : 'text-gray-700 dark:text-gray-300'}>transactions</span>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </nav>
          <div className="p-6 border-t border-[#E2E8F0] dark:border-gray-800">
            <button
              className="w-full flex items-center justify-center gap-2 p-3 rounded-xl neumorphic-card text-xs font-bold text-[#616289] hover:text-primary transition-all"
              onClick={handleSettings}
            >
              <span className="material-symbols-outlined text-[18px]">settings</span>
              Settings
            </button>
          </div>
        </aside>

        {/* Main Content */}
        <main className="flex-1 flex flex-col overflow-hidden bg-background-light dark:bg-background-dark">
          {/* Sub-header */}
          <header className="h-14 flex items-center justify-between px-6 border-b border-[#E2E8F0] dark:border-gray-800 bg-white/20 dark:bg-black/10 backdrop-blur-md">
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-emerald-status"></div>
                <span className="text-xs font-bold uppercase tracking-wider text-[#616289]">Main Postgres</span>
              </div>
              <div className="h-4 w-px bg-gray-300 dark:bg-gray-700"></div>
              <div className="flex items-center gap-2">
                <span className="material-symbols-outlined text-primary text-[20px]">table_chart</span>
                <span className="text-sm font-semibold">{selectedTable}</span>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <button
                className="flex items-center gap-2 px-4 py-1.5 rounded-lg border border-primary/30 bg-primary/5 text-primary text-xs font-bold hover:bg-primary/10 transition-all"
                onClick={handleOptimize}
                disabled={isOptimizing || !sql.trim()}
              >
                <span className="material-symbols-outlined text-[18px]">auto_awesome</span>
                AI Optimize
              </button>
              <button
                className="flex items-center gap-2 px-4 py-1.5 rounded-lg bg-primary text-white text-xs font-bold shadow-lg shadow-primary/20 hover:brightness-110 transition-all"
                onClick={handleExecute}
                disabled={isExecuting || !sql.trim()}
              >
                <span className="material-symbols-outlined text-[16px]">play_arrow</span>
                Run
              </button>
              <button className="flex items-center gap-2 px-3 py-1.5 rounded-lg neumorphic-card text-xs font-bold text-[#616289] hover:text-primary transition-all">
                <span className="material-symbols-outlined text-[16px]">save</span>
                Save
              </button>
            </div>
          </header>

          <div className="flex-1 flex flex-col min-w-0">
            {/* SQL Editor */}
            <div className="flex-1 relative border-t border-[#E2E8F0] dark:border-gray-800">
              <div className="absolute inset-0 p-6 font-mono text-sm overflow-y-auto custom-scrollbar bg-white/50 dark:bg-black/20 pointer-events-none">
                <div className="space-y-1">
                  {sql.split('\n').map((line, index) => (
                    <div key={index} className="flex">
                      <div className="w-10 text-gray-400 text-right pr-4 select-none opacity-50">
                        {index + 1}
                      </div>
                      <div className="flex-1 text-[#111118] dark:text-blue-100">
                        {highlightSQL(line).map((token, tokenIndex) => {
                          const className =
                            token.type === 'keyword' ? 'sql-keyword' :
                            token.type === 'string' ? 'sql-string' :
                            token.type === 'number' ? 'sql-number' :
                            token.type === 'comment' ? 'sql-comment' : '';
                          return (
                            <span key={tokenIndex} className={className}>
                              {token.text}
                            </span>
                          );
                        })}
                      </div>
                    </div>
                  ))}
                  {showOptimized && optimizedSQL && (
                    <>
                      <div className="flex">
                        <div className="w-10 text-gray-400 text-right pr-4 select-none opacity-50"></div>
                        <div className="flex-1 sql-comment">-- AI Optimized Version</div>
                      </div>
                      <div className="flex">
                        <div className="w-10 text-gray-400 text-right pr-4 select-none opacity-50"></div>
                        <div className="flex-1 sql-comment">-- Optimization notes: {optimizedSQL.suggestions.join(', ')}</div>
                      </div>
                      {optimizedSQL.optimized.split('\n').map((line, index) => (
                        <div key={`opt-${index}`} className="flex">
                          <div className="w-10 text-gray-400 text-right pr-4 select-none opacity-50">
                            {sql.split('\n').length + index + 1}
                          </div>
                          <div className="flex-1 text-[#111118] dark:text-blue-300">
                            {highlightSQL(line).map((token, tokenIndex) => {
                              const className =
                                token.type === 'keyword' ? 'sql-keyword' :
                                token.type === 'string' ? 'sql-string' :
                                token.type === 'number' ? 'sql-number' :
                                token.type === 'comment' ? 'sql-comment' : '';
                              return (
                                <span key={tokenIndex} className={className}>
                                  {token.text}
                                </span>
                              );
                            })}
                          </div>
                        </div>
                      ))}
                    </>
                  )}
                </div>
              </div>
              <textarea
                ref={editorRef}
                className="absolute inset-0 w-full h-full p-6 font-mono text-sm bg-transparent text-transparent caret-primary resize-none outline-none"
                style={{ color: 'transparent', caretColor: '#6366F1' }}
                value={sql}
                onChange={(e) => setSql(e.target.value)}
                onKeyDown={handleKeyDown}
                spellCheck={false}
                placeholder="Enter your SQL query here... (Ctrl+Enter to execute)"
              />
            </div>

            {/* Data View - 60% height */}
            <div className="h-[60%] border-t border-[#E2E8F0] dark:border-gray-800 bg-white/40 dark:bg-black/30 overflow-hidden flex flex-col">
              <div className="px-6 h-12 flex items-center justify-between bg-white/60 dark:bg-white/5 border-b border-[#E2E8F0] dark:border-gray-800">
                <div className="flex items-center gap-3">
                  <span className="text-[10px] font-bold uppercase tracking-widest text-[#616289]">Data View</span>
                  <div className="flex items-center gap-1 bg-primary/10 px-2 py-0.5 rounded text-[10px] text-primary font-bold">
                    <span>LIVE</span>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <div className="flex items-center gap-2 border-r border-gray-200 dark:border-gray-700 pr-3">
                    <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white dark:bg-gray-900 border border-[#E2E8F0] dark:border-gray-800 text-[11px] font-bold text-[#616289] hover:text-primary hover:border-primary/30 transition-all">
                      <span className="material-symbols-outlined text-[16px]">add_row_below</span>
                      Add Row
                    </button>
                    <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white dark:bg-gray-900 border border-[#E2E8F0] dark:border-gray-800 text-[11px] font-bold text-red-500 hover:bg-red-50 hover:border-red-200 transition-all">
                      <span className="material-symbols-outlined text-[16px]">delete</span>
                      Delete Row
                    </button>
                  </div>
                  <button className="flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-[#6366F1] text-white text-[11px] font-bold shadow-lg shadow-indigo-500/30 hover:brightness-110 transition-all">
                    <span className="material-symbols-outlined text-[16px]">publish</span>
                    Commit Changes
                  </button>
                </div>
              </div>
              <div className="flex-1 overflow-auto p-0 custom-scrollbar">
                <table className="w-full text-left text-xs border-collapse">
                  <thead className="text-gray-400 bg-white/30 dark:bg-black/10 sticky top-0 backdrop-blur-sm z-20">
                    <tr>
                      {displayData.columns.map((col) => (
                        <th key={col} className="py-3 px-6 font-medium border-b border-gray-100 dark:border-gray-800">
                          {col}
                        </th>
                      ))}
                      <th className="py-3 px-6 font-medium border-b border-gray-100 dark:border-gray-800 w-20 text-center">Action</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-50 dark:divide-gray-800">
                    {displayData.rows.map((row, rowIndex) => (
                      <tr key={rowIndex} className="hover:bg-primary/5 transition-colors group">
                        {displayData.columns.map((col) => {
                          const value = row[col] || '';
                          return (
                            <td key={col} className="py-2 px-6 font-mono relative editable-cell">
                              {col === 'status' ? (
                                <select
                                  className="w-full bg-transparent border-none p-0 focus:ring-0 text-[10px] font-bold appearance-none"
                                  defaultValue={value as string}
                                >
                                  <option value="COMPLETED">COMPLETED</option>
                                  <option value="PENDING">PENDING</option>
                                  <option value="FAILED">FAILED</option>
                                </select>
                              ) : (
                                <input
                                  className="w-full bg-transparent border-none p-0 focus:ring-0 text-xs font-mono"
                                  type="text"
                                  defaultValue={value as string}
                                />
                              )}
                            </td>
                          );
                        })}
                        <td className="py-2 px-6 text-center">
                          <button className="w-8 h-8 rounded-lg flex items-center justify-center text-gray-400 hover:text-primary hover:bg-primary/10 transition-all mx-auto">
                            <span className="material-symbols-outlined text-[18px]">edit_square</span>
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {/* Status Bar */}
              <div className="h-10 shrink-0 flex items-center justify-between px-6 bg-white dark:bg-black/40 border-t border-[#E2E8F0] dark:border-gray-800 text-[10px] font-medium text-[#616289]">
                <div className="flex items-center gap-2 overflow-hidden mr-4">
                  <span className="material-symbols-outlined text-[14px] shrink-0">terminal</span>
                  <span className="truncate italic opacity-70">{sql}</span>
                </div>
                <div className="flex items-center gap-4 shrink-0">
                  <div className="flex items-center gap-2 text-emerald-status">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-status animate-pulse"></span>
                    <span className="uppercase tracking-widest font-bold">Editor Ready</span>
                  </div>
                  <div className="h-3 w-px bg-gray-200 dark:bg-gray-700"></div>
                  <span className="flex items-center gap-1">
                    <span className="font-bold text-primary">{displayData.rowCount.toLocaleString()}</span> records found
                  </span>
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
};
