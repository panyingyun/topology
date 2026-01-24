import React, { useState } from 'react';
import { ImportFormat, ExportFormat } from '../types';
import { importExportService } from '../services/importExportService';

export const ImportExport: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'import' | 'export'>('import');
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importFormat, setImportFormat] = useState<ImportFormat>('csv');
  const [exportFormat, setExportFormat] = useState<ExportFormat>('csv');
  const [isImporting, setIsImporting] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [showImportDialog, setShowImportDialog] = useState(false);
  const [showExportDialog, setShowExportDialog] = useState(false);

  const handleImport = async () => {
    if (!importFile) return;
    setIsImporting(true);
    try {
      const result = await importExportService.importData('1', {
        format: importFormat,
        encoding: 'UTF-8',
        tableName: 'imported_data',
      }, importFile);
      alert(result.message);
      setShowImportDialog(false);
    } catch (error) {
      alert('Import failed');
    } finally {
      setIsImporting(false);
    }
  };

  const handleExport = async () => {
    setIsExporting(true);
    try {
      const result = await importExportService.exportData('1', {
        format: exportFormat,
        tableName: 'transactions',
        rowLimit: 1000,
      });
      if (result.success && result.data) {
        const url = URL.createObjectURL(result.data);
        const a = document.createElement('a');
        a.href = url;
        a.download = result.filename || 'export';
        a.click();
        URL.revokeObjectURL(url);
      }
      setShowExportDialog(false);
    } catch (error) {
      alert('Export failed');
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="flex h-screen overflow-hidden">
      <aside className="glass-sidebar w-72 flex flex-col z-10">
        <div className="p-6 flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-white">
            <span className="material-symbols-outlined">database</span>
          </div>
          <div>
            <h1 className="font-bold text-lg leading-tight">Topology</h1>
            <p className="text-xs text-[#616289] dark:text-gray-400">Pro Workspace</p>
          </div>
        </div>

        <nav className="flex-1 overflow-y-auto px-4 custom-scrollbar">
          <div className="space-y-6 pt-6">
            <div>
              <div className="flex items-center justify-between px-2 mb-2">
                <span className="text-[11px] font-bold uppercase tracking-wider text-[#616289]">Data Operations</span>
              </div>
              <div className="space-y-1">
                <div className="flex items-center gap-3 px-3 py-2 rounded-xl bg-primary/10 text-primary border border-primary/20">
                  <span className="material-symbols-outlined text-[20px]">swap_vertical_circle</span>
                  <span className="text-sm font-medium">Import & Export</span>
                </div>
              </div>
            </div>
          </div>
        </nav>
      </aside>

      <main className="flex-1 overflow-y-auto bg-background-light dark:bg-background-dark">
        <div className="px-10 pt-8 flex items-center justify-between">
          <div className="flex items-center gap-2 text-sm">
            <a className="text-[#616289] hover:text-primary transition-colors" href="#">Data Ops</a>
            <span className="text-[#616289]">/</span>
            <span className="font-semibold text-[#111118] dark:text-white">Import & Export Wizard</span>
          </div>
        </div>

        <div className="max-w-6xl mx-auto px-10 py-10">
          <div className="mb-10 text-center">
            <h2 className="text-4xl font-extrabold text-[#111118] dark:text-white tracking-tight mb-3">Import & Export Wizard</h2>
            <p className="text-[#616289] dark:text-gray-400">Migrate your data effortlessly between formats and clusters.</p>
          </div>

          <div className="grid grid-cols-2 gap-8 mb-12">
            <div
              className={`neumorphic-card p-8 rounded-3xl cursor-pointer relative overflow-hidden group transition-all ${
                activeTab === 'import' ? 'border-2 border-primary ring-4 ring-primary/5' : 'border border-white/40 dark:border-white/5'
              }`}
              onClick={() => {
                setActiveTab('import');
                setShowImportDialog(true);
              }}
            >
              <div className="flex items-start justify-between">
                <div>
                  <div className="w-14 h-14 rounded-2xl bg-primary/10 flex items-center justify-center text-primary mb-6">
                    <span className="material-symbols-outlined text-4xl">upload_file</span>
                  </div>
                  <h3 className="text-2xl font-bold mb-2">Data Import</h3>
                  <p className="text-sm text-[#616289] leading-relaxed">Import CSV, JSON, or SQL dump files directly into your selected database tables.</p>
                </div>
                <div className="absolute -right-4 -bottom-4 opacity-10 group-hover:opacity-20 transition-opacity">
                  <span className="material-symbols-outlined text-9xl">download</span>
                </div>
              </div>
            </div>

            <div
              className={`neumorphic-card p-8 rounded-3xl cursor-pointer relative overflow-hidden group transition-all ${
                activeTab === 'export' ? 'border-2 border-primary ring-4 ring-primary/5' : 'border border-white/40 dark:border-white/5'
              }`}
              onClick={() => {
                setActiveTab('export');
                setShowExportDialog(true);
              }}
            >
              <div className="flex items-start justify-between">
                <div>
                  <div className="w-14 h-14 rounded-2xl bg-emerald-500/10 flex items-center justify-center text-emerald-500 mb-6">
                    <span className="material-symbols-outlined text-4xl">download_for_offline</span>
                  </div>
                  <h3 className="text-2xl font-bold mb-2">Data Export</h3>
                  <p className="text-sm text-[#616289] leading-relaxed">Extract tables or query results into structured formats including Excel, Parquet, and XML.</p>
                </div>
                <div className="absolute -right-4 -bottom-4 opacity-10 group-hover:opacity-20 transition-opacity">
                  <span className="material-symbols-outlined text-9xl">file_export</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Import Dialog */}
      {showImportDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/40 backdrop-blur-sm" onClick={() => setShowImportDialog(false)}></div>
          <div className="relative glass-modal w-full max-w-2xl bg-[#FFFFFF] rounded-[24px] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-300">
            <div className="p-8">
              <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-2xl bg-indigo-500/10 flex items-center justify-center text-primary">
                    <span className="material-symbols-outlined text-3xl">upload_file</span>
                  </div>
                  <div>
                    <h2 className="text-2xl font-bold">Data Import</h2>
                    <p className="text-sm text-[#616289]">Configure your data source settings</p>
                  </div>
                </div>
                <button
                  className="w-10 h-10 rounded-full hover:bg-gray-100 dark:hover:bg-white/10 flex items-center justify-center transition-colors"
                  onClick={() => setShowImportDialog(false)}
                >
                  <span className="material-symbols-outlined text-[#616289]">close</span>
                </button>
              </div>

              <div className="space-y-8">
                <div>
                  <label className="block text-sm font-semibold text-[#616289] mb-3 px-1">Source File</label>
                  <div
                    className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-2xl border-2 border-dashed border-[#d1d1d6] dark:border-gray-800 p-12 text-center cursor-pointer hover:bg-indigo-50/30 transition-colors"
                    onDrop={(e) => {
                      e.preventDefault();
                      const file = e.dataTransfer.files[0];
                      if (file) setImportFile(file);
                    }}
                    onDragOver={(e) => e.preventDefault()}
                  >
                    <span className="material-symbols-outlined text-5xl text-indigo-500/40 mb-3">cloud_upload</span>
                    <p className="text-base font-medium text-gray-700 dark:text-gray-300">
                      {importFile ? importFile.name : 'Drag and drop file here, or browse files'}
                    </p>
                    <p className="text-xs text-gray-400 mt-2">Supports CSV, JSON, XLSX, SQL (Max 2GB)</p>
                    <input
                      type="file"
                      className="hidden"
                      id="file-input"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file) setImportFile(file);
                      }}
                    />
                    <label htmlFor="file-input" className="cursor-pointer text-primary font-bold mt-2 inline-block">
                      Browse
                    </label>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-semibold text-[#616289] mb-2 px-1">Encoding</label>
                    <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl h-12 flex items-center px-4">
                      <select className="bg-transparent border-none focus:ring-0 w-full text-sm font-medium appearance-none">
                        <option>UTF-8 (Recommended)</option>
                        <option>ISO-8859-1</option>
                        <option>Windows-1252</option>
                      </select>
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-semibold text-[#616289] mb-2 px-1">Format</label>
                    <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl h-12 flex items-center px-4">
                      <select
                        className="bg-transparent border-none focus:ring-0 w-full text-sm font-medium appearance-none"
                        value={importFormat}
                        onChange={(e) => setImportFormat(e.target.value as ImportFormat)}
                      >
                        <option value="csv">CSV</option>
                        <option value="json">JSON</option>
                        <option value="xlsx">Excel</option>
                        <option value="sql">SQL</option>
                      </select>
                    </div>
                  </div>
                </div>
              </div>

              <div className="flex gap-4 pt-4 mt-8">
                <button
                  className="flex-1 py-4 rounded-2xl bg-primary text-white font-bold shadow-xl shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all transform active:scale-[0.98]"
                  onClick={handleImport}
                  disabled={!importFile || isImporting}
                >
                  {isImporting ? 'Importing...' : 'Next: Field Mapping'}
                </button>
                <button
                  className="px-8 py-4 rounded-2xl border border-gray-200 dark:border-gray-700 font-bold text-[#616289] hover:bg-gray-50 dark:hover:bg-white/5 transition-colors"
                  onClick={() => setShowImportDialog(false)}
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Export Dialog */}
      {showExportDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-[2px]">
          <div className="glass-modal w-[600px] rounded-[24px] shadow-2xl p-8 transform scale-100 opacity-100">
            <div className="flex justify-between items-start mb-8">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-2xl bg-emerald-status/10 flex items-center justify-center text-emerald-status">
                  <span className="material-symbols-outlined text-3xl">file_export</span>
                </div>
                <div>
                  <h2 className="text-2xl font-bold text-[#111118] dark:text-white">Data Export Wizard</h2>
                  <p className="text-sm text-[#616289] dark:text-gray-400">Configure your data extraction parameters</p>
                </div>
              </div>
              <button
                className="w-10 h-10 rounded-full hover:bg-black/5 dark:hover:bg-white/5 flex items-center justify-center text-[#616289] transition-colors"
                onClick={() => setShowExportDialog(false)}
              >
                <span className="material-symbols-outlined">close</span>
              </button>
            </div>

            <div className="space-y-8">
              <section>
                <label className="block text-sm font-bold text-[#616289] mb-4 uppercase tracking-wider">Export Format</label>
                <div className="grid grid-cols-3 gap-4">
                  {(['csv', 'excel', 'sql'] as ExportFormat[]).map((format) => (
                    <button
                      key={format}
                      className={`flex flex-col items-center gap-2 p-4 rounded-2xl transition-all ${
                        exportFormat === format
                          ? 'neumorphic-inset bg-background-light dark:bg-background-dark border-2 border-emerald-status/50 text-emerald-status ring-2 ring-emerald-status/10'
                          : 'neumorphic-card hover:bg-white/50 text-[#616289]'
                      }`}
                      onClick={() => setExportFormat(format)}
                    >
                      <span className="material-symbols-outlined text-2xl">
                        {format === 'csv' ? 'csv' : format === 'excel' ? 'table_view' : 'terminal'}
                      </span>
                      <span className="text-xs font-bold">{format.toUpperCase()}</span>
                    </button>
                  ))}
                </div>
              </section>

              <section>
                <label className="block text-sm font-bold text-[#616289] mb-4 uppercase tracking-wider">Target Source</label>
                <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-2xl p-2">
                  <div className="flex items-center gap-3 p-3 rounded-xl bg-white dark:bg-[#1a1c36] shadow-sm">
                    <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center text-primary">
                      <span className="material-symbols-outlined text-xl">database</span>
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-bold">production_db_replica</p>
                      <p className="text-[10px] text-[#616289]">Public Schema • 24 Tables</p>
                    </div>
                    <span className="material-symbols-outlined text-[#616289] cursor-pointer">expand_more</span>
                  </div>
                </div>
              </section>

              <section>
                <label className="block text-sm font-bold text-[#616289] mb-4 uppercase tracking-wider">Data Range Settings</label>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-xs font-semibold text-[#616289] mb-2 px-1">Row Limit</p>
                    <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl h-11 flex items-center px-4">
                      <input
                        className="bg-transparent border-none focus:ring-0 text-sm font-bold w-full"
                        type="number"
                        defaultValue={1000}
                      />
                    </div>
                  </div>
                  <div>
                    <p className="text-xs font-semibold text-[#616289] mb-2 px-1">Filter by Date</p>
                    <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl h-11 flex items-center px-4">
                      <select className="bg-transparent border-none focus:ring-0 w-full text-sm font-medium appearance-none">
                        <option>Last 30 Days</option>
                        <option>All Time</option>
                        <option>Custom Range</option>
                      </select>
                    </div>
                  </div>
                </div>
              </section>

              <div className="flex gap-4 pt-4">
                <button
                  className="flex-1 py-4 rounded-2xl bg-emerald-status text-white font-bold shadow-xl shadow-emerald-500/20 hover:shadow-emerald-500/30 transition-all transform active:scale-[0.98] flex items-center justify-center gap-2"
                  onClick={handleExport}
                  disabled={isExporting}
                >
                  <span className="material-symbols-outlined text-xl">bolt</span>
                  {isExporting ? 'Exporting...' : 'Start Export'}
                </button>
                <button
                  className="px-8 py-4 rounded-2xl neumorphic-card text-[#616289] font-bold hover:text-red-500 transition-colors"
                  onClick={() => setShowExportDialog(false)}
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
