import React from 'react';
import { Connection } from '../types';

interface HeaderProps {
  currentConnection?: Connection;
  currentTable?: string;
  onNewConnection: () => void;
}

export const Header: React.FC<HeaderProps> = ({
  currentConnection,
  currentTable,
  onNewConnection,
}) => {
  return (
    <header className="h-16 flex items-center justify-between px-6 glass-header z-30 shrink-0">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-white shadow-lg shadow-primary/20">
          <span className="material-symbols-outlined">database</span>
        </div>
        <div>
          <h1 className="font-bold text-lg leading-tight tracking-tight">DB Studio</h1>
          <p className="text-[10px] text-[#616289] font-semibold dark:text-gray-400 uppercase tracking-widest">Enterprise</p>
        </div>
      </div>

      <div className="flex items-center gap-6">
        <button
          className="flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-white text-xs font-bold shadow-lg shadow-primary/30 hover:brightness-110 transition-all"
          onClick={onNewConnection}
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
  );
};
