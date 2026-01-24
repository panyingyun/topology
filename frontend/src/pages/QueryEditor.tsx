import React, { useState, useEffect, useRef } from 'react';
import { useParams } from 'react-router-dom';
import { QueryResult, OptimizedSQL } from '../types';
import { queryService } from '../services/queryService';
import { DataTable } from '../components/DataTable';
import { highlightSQL } from '../utils/sqlHighlighter';

export const QueryEditor: React.FC = () => {
  const { connectionId } = useParams<{ connectionId?: string }>();
  const [sql, setSql] = useState('SELECT * FROM transactions LIMIT 1000;');
  const [result, setResult] = useState<QueryResult | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);
  const [optimizedSQL, setOptimizedSQL] = useState<OptimizedSQL | null>(null);
  const [isOptimizing, setIsOptimizing] = useState(false);
  const [showOptimized, setShowOptimized] = useState(false);
  const editorRef = useRef<HTMLTextAreaElement>(null);
  const [lineNumbers, setLineNumbers] = useState<string[]>(['1']);

  useEffect(() => {
    updateLineNumbers();
  }, [sql]);

  const updateLineNumbers = () => {
    const lines = sql.split('\n');
    setLineNumbers(lines.map((_, i) => (i + 1).toString()));
  };

  const handleExecute = async () => {
    if (!sql.trim()) return;

    setIsExecuting(true);
    try {
      const queryResult = await queryService.executeQuery(connectionId || '1', sql);
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

  const handleApplyOptimized = () => {
    if (optimizedSQL) {
      setSql(optimizedSQL.optimized);
      setShowOptimized(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      handleExecute();
    }
  };

  return (
    <div className="flex flex-col h-screen overflow-hidden">
      <header className="h-14 flex items-center justify-between px-6 border-b border-[#E2E8F0] dark:border-gray-800 bg-white/20 dark:bg-black/10 backdrop-blur-md">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-emerald-status"></div>
            <span className="text-xs font-bold uppercase tracking-wider text-[#616289]">Main Postgres</span>
          </div>
          <div className="h-4 w-px bg-gray-300 dark:bg-gray-700"></div>
          <div className="flex items-center gap-2">
            <span className="material-symbols-outlined text-primary text-[20px]">table_chart</span>
            <span className="text-sm font-semibold">transactions</span>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button
            className="flex items-center gap-2 px-4 py-1.5 rounded-lg border border-primary/30 bg-primary/5 text-primary text-xs font-bold hover:bg-primary/10 transition-all"
            onClick={handleOptimize}
            disabled={isOptimizing || !sql.trim()}
          >
            <span className="material-symbols-outlined text-[18px]">auto_awesome</span>
            {isOptimizing ? 'Optimizing...' : 'AI Optimize'}
          </button>
          <button
            className="flex items-center gap-2 px-4 py-1.5 rounded-lg bg-primary text-white text-xs font-bold shadow-lg shadow-primary/20 hover:brightness-110 transition-all"
            onClick={handleExecute}
            disabled={isExecuting || !sql.trim()}
          >
            <span className="material-symbols-outlined text-[16px]">play_arrow</span>
            {isExecuting ? 'Running...' : 'Run'}
          </button>
          <button className="flex items-center gap-2 px-3 py-1.5 rounded-lg neumorphic-card text-xs font-bold text-[#616289] hover:text-primary transition-all">
            <span className="material-symbols-outlined text-[16px]">save</span>
            Save
          </button>
        </div>
      </header>

      <div className="flex-1 flex flex-col min-w-0">
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

        {showOptimized && optimizedSQL && (
          <div className="border-t border-[#E2E8F0] dark:border-gray-800 bg-primary/5 p-6">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="font-bold text-sm mb-1">AI Optimized Version</h3>
                <p className="text-xs text-[#616289]">{optimizedSQL.suggestions.join(', ')}</p>
              </div>
              <button
                className="px-4 py-2 rounded-lg bg-primary text-white text-xs font-bold hover:brightness-110 transition-all"
                onClick={handleApplyOptimized}
              >
                Apply
              </button>
            </div>
            <div className="bg-white/50 dark:bg-black/20 rounded-lg p-4 font-mono text-xs">
              <pre className="whitespace-pre-wrap">{optimizedSQL.optimized}</pre>
            </div>
          </div>
        )}

        {result && (
          <DataTable
            data={result}
            onUpdate={(rowIndex, column, value) => {
              console.log('Update:', { rowIndex, column, value });
            }}
            onAddRow={() => {
              console.log('Add row');
            }}
            onDeleteRow={(rowIndex) => {
              console.log('Delete row:', rowIndex);
            }}
            onCommit={() => {
              console.log('Commit changes');
            }}
          />
        )}
      </div>
    </div>
  );
};
