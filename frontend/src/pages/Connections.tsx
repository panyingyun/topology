import React, { useState, useEffect } from 'react';
import { Connection, DatabaseType } from '../types';
import { connectionService } from '../services/connectionService';
import { ConnectionCard } from '../components/ConnectionCard';

interface ConnectionDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (connection: Omit<Connection, 'id' | 'status'>) => void;
}

const ConnectionDialog: React.FC<ConnectionDialogProps> = ({ isOpen, onClose, onSave }) => {
  const [selectedType, setSelectedType] = useState<DatabaseType>('mysql');
  const [formData, setFormData] = useState({
    name: '',
    host: '127.0.0.1',
    port: 3306,
    username: 'root',
    password: '',
    database: '',
    savePassword: true,
  });
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<boolean | null>(null);

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave({
      name: formData.name,
      type: selectedType,
      host: formData.host,
      port: formData.port,
      username: formData.username,
      password: formData.savePassword ? formData.password : undefined,
      database: formData.database || undefined,
      savePassword: formData.savePassword,
    });
    onClose();
  };

  const handleTest = async () => {
    setTesting(true);
    setTestResult(null);
    const result = await connectionService.testConnection({
      name: formData.name,
      type: selectedType,
      host: formData.host,
      port: formData.port,
      username: formData.username,
      password: formData.password,
      database: formData.database || undefined,
    });
    setTestResult(result);
    setTesting(false);
  };

  const databaseTypes: Array<{ type: DatabaseType; label: string; icon: string; defaultPort: number }> = [
    { type: 'postgresql', label: 'PostgreSQL', icon: 'database', defaultPort: 5432 },
    { type: 'mysql', label: 'MySQL', icon: 'table_chart', defaultPort: 3306 },
    { type: 'redis', label: 'Redis', icon: 'bolt', defaultPort: 6379 },
    { type: 'mongodb', label: 'MongoDB', icon: 'schema', defaultPort: 27017 },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-md p-6 overflow-y-auto">
      <div className="bg-white dark:bg-[#101122] w-full max-w-4xl rounded-[24px] border border-[#E2E8F0] dark:border-white/10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden">
        <div className="px-8 pt-8 pb-4 flex items-center justify-between flex-shrink-0">
          <div>
            <h2 className="text-3xl font-extrabold text-[#111118] dark:text-white tracking-tight">Create New Connection</h2>
            <p className="text-[#616289] dark:text-gray-400 mt-1">Configure your database connection settings.</p>
          </div>
          <button
            className="w-10 h-10 rounded-full hover:bg-gray-100 dark:hover:bg-white/5 flex items-center justify-center transition-colors"
            onClick={onClose}
          >
            <span className="material-symbols-outlined text-gray-500">close</span>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto px-8 pb-8 custom-scrollbar">
          <div className="space-y-10 pt-4">
            <div>
              <label className="block text-xs font-bold uppercase tracking-widest text-[#616289] mb-4 px-1">
                1. Choose Database Engine
              </label>
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                {databaseTypes.map((db) => (
                  <div
                    key={db.type}
                    className={`p-5 rounded-[20px] border cursor-pointer transition-all flex flex-col items-center text-center ${
                      selectedType === db.type
                        ? 'bg-primary/5 border-2 border-primary ring-4 ring-primary/5'
                        : 'neumorphic-card border-white/40 dark:border-white/5 hover:border-primary/50'
                    }`}
                    onClick={() => {
                      setSelectedType(db.type);
                      setFormData({ ...formData, port: db.defaultPort });
                    }}
                  >
                    <div className={`w-12 h-12 rounded-2xl flex items-center justify-center mb-3 ${
                      db.type === 'postgresql' ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-600' :
                      db.type === 'mysql' ? 'bg-orange-100 dark:bg-orange-900/30 text-orange-600' :
                      db.type === 'redis' ? 'bg-red-100 dark:bg-red-900/30 text-red-600' :
                      'bg-green-100 dark:bg-green-900/30 text-green-600'
                    }`}>
                      <span className="material-symbols-outlined text-2xl">{db.icon}</span>
                    </div>
                    <h3 className="font-bold text-sm">{db.label}</h3>
                    <p className="text-[10px] text-[#616289] mt-1">Port: {db.defaultPort}</p>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <label className="block text-xs font-bold uppercase tracking-widest text-[#616289] mb-4 px-1">
                2. Connection Details ({databaseTypes.find(d => d.type === selectedType)?.label})
              </label>
              <div className="neumorphic-card p-8 rounded-3xl border border-white/40 dark:border-white/5">
                <div className="grid grid-cols-2 gap-x-8 gap-y-6">
                  <div>
                    <label className="block text-xs font-semibold text-[#616289] mb-2 px-1">Connection Name</label>
                    <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl h-12 flex items-center px-4">
                      <input
                        className="bg-transparent border-none focus:ring-0 w-full font-medium text-sm"
                        type="text"
                        value={formData.name}
                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                        placeholder="My Database"
                        required
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-[#616289] mb-2 px-1">Host Address</label>
                    <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl h-12 flex items-center px-4">
                      <span className="material-symbols-outlined text-gray-400 mr-3 text-sm">dns</span>
                      <input
                        className="bg-transparent border-none focus:ring-0 w-full font-medium text-sm"
                        type="text"
                        value={formData.host}
                        onChange={(e) => setFormData({ ...formData, host: e.target.value })}
                        required
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-[#616289] mb-2 px-1">Port</label>
                    <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl h-12 flex items-center px-4">
                      <span className="material-symbols-outlined text-gray-400 mr-3 text-sm">numbers</span>
                      <input
                        className="bg-transparent border-none focus:ring-0 w-full font-medium text-sm"
                        type="number"
                        value={formData.port}
                        onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) || 3306 })}
                        required
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-[#616289] mb-2 px-1">Username</label>
                    <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl h-12 flex items-center px-4">
                      <span className="material-symbols-outlined text-gray-400 mr-3 text-sm">person</span>
                      <input
                        className="bg-transparent border-none focus:ring-0 w-full font-medium text-sm"
                        type="text"
                        value={formData.username}
                        onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                        required
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-[#616289] mb-2 px-1">Password</label>
                    <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl h-12 flex items-center px-4">
                      <span className="material-symbols-outlined text-gray-400 mr-3 text-sm">lock</span>
                      <input
                        className="bg-transparent border-none focus:ring-0 w-full font-medium text-sm flex-1"
                        type="password"
                        value={formData.password}
                        onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                      />
                      <span className="material-symbols-outlined text-gray-400 cursor-pointer text-sm">visibility</span>
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-[#616289] mb-2 px-1">Database Name</label>
                    <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl h-12 flex items-center px-4">
                      <span className="material-symbols-outlined text-gray-400 mr-3 text-sm">folder</span>
                      <input
                        className="bg-transparent border-none focus:ring-0 w-full font-medium text-sm"
                        type="text"
                        value={formData.database}
                        onChange={(e) => setFormData({ ...formData, database: e.target.value })}
                        placeholder="Optional"
                      />
                    </div>
                  </div>
                </div>
                <div className="mt-6">
                  <label className="flex items-center gap-3 cursor-pointer">
                    <div
                      className={`w-10 h-5 rounded-full relative transition-all ${
                        formData.savePassword ? 'bg-primary' : 'bg-gray-300'
                      }`}
                      onClick={() => setFormData({ ...formData, savePassword: !formData.savePassword })}
                    >
                      <div
                        className={`absolute left-1 top-1 w-3 h-3 bg-white rounded-full transition-transform ${
                          formData.savePassword ? 'translate-x-5' : ''
                        }`}
                      ></div>
                    </div>
                    <span className="text-xs font-medium text-[#616289]">Save password in keychain</span>
                  </label>
                </div>
              </div>
            </div>
          </div>
        </form>

        <div className="p-6 border-t border-[#E2E8F0] dark:border-gray-800 flex items-center justify-between bg-gray-50/50 dark:bg-white/5 flex-shrink-0">
          <div className="flex items-center gap-2 text-emerald-status font-medium text-xs">
            <span className="material-symbols-outlined text-[16px]">verified_user</span>
            SSH Tunnel: Enabled
          </div>
          <div className="flex gap-3">
            <button
              className="px-6 py-2.5 rounded-xl font-semibold text-[#616289] hover:bg-gray-100 dark:hover:bg-white/5 transition-colors text-sm"
              onClick={handleTest}
              disabled={testing}
            >
              {testing ? 'Testing...' : testResult === true ? '✓ Connected' : testResult === false ? '✗ Failed' : 'Test'}
            </button>
            <button
              className="px-8 py-2.5 rounded-xl bg-primary text-white font-bold shadow-lg shadow-primary/25 hover:brightness-110 active:scale-95 transition-all text-sm"
              onClick={handleSubmit}
            >
              Connect
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export const Connections: React.FC = () => {
  const [connections, setConnections] = useState<Connection[]>([]);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    loadConnections();
  }, []);

  const loadConnections = async () => {
    const conns = await connectionService.getConnections();
    setConnections(conns);
  };

  const handleSaveConnection = async (connection: Omit<Connection, 'id' | 'status'>) => {
    await connectionService.createConnection(connection);
    await loadConnections();
  };

  const handleDeleteConnection = async (id: string) => {
    await connectionService.deleteConnection(id);
    await loadConnections();
  };

  const filteredConnections = connections.filter(conn =>
    conn.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const groupedConnections = filteredConnections.reduce((acc, conn) => {
    const group = conn.group || 'Other';
    if (!acc[group]) {
      acc[group] = [];
    }
    acc[group].push(conn);
    return acc;
  }, {} as Record<string, Connection[]>);

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

        <div className="px-6 py-4">
          <div className="neumorphic-inset bg-background-light dark:bg-background-dark rounded-xl flex items-center px-4 h-11">
            <span className="material-symbols-outlined text-gray-400 text-[20px]">search</span>
            <input
              className="bg-transparent border-none focus:ring-0 text-sm w-full placeholder:text-gray-400"
              placeholder="Search databases..."
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>

        <nav className="flex-1 overflow-y-auto px-4 custom-scrollbar">
          <div className="space-y-6">
            {Object.entries(groupedConnections).map(([group, conns]) => (
              <div key={group}>
                <div className="flex items-center justify-between px-2 mb-2">
                  <span className="text-[11px] font-bold uppercase tracking-wider text-[#616289]">{group}</span>
                  <span className="material-symbols-outlined text-[16px] cursor-pointer">expand_more</span>
                </div>
                <div className="space-y-1">
                  {conns.map((conn) => (
                    <ConnectionCard
                      key={conn.id}
                      connection={conn}
                      isSelected={false}
                      onClick={() => {}}
                      onDelete={() => handleDeleteConnection(conn.id)}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        </nav>

        <div className="p-4 border-t border-[#E2E8F0] dark:border-gray-800">
          <button
            className="w-full bg-primary text-white font-semibold py-3 rounded-xl shadow-lg shadow-primary/30 flex items-center justify-center gap-2 hover:brightness-110 transition-all"
            onClick={() => setIsDialogOpen(true)}
          >
            <span className="material-symbols-outlined text-[20px]">add_circle</span>
            New Cluster
          </button>
        </div>
      </aside>

      <main className="flex-1 overflow-y-auto bg-background-light dark:bg-background-dark">
        <div className="px-10 pt-8 flex items-center justify-between">
          <div className="flex items-center gap-2 text-sm">
            <a className="text-[#616289] hover:text-primary transition-colors" href="#">Connections</a>
            <span className="text-[#616289]">/</span>
            <span className="font-semibold text-[#111118] dark:text-white">Overview</span>
          </div>
        </div>

        <div className="max-w-6xl mx-auto px-10 py-12">
          <div className="grid grid-cols-3 gap-8 mb-12">
            <div className="neumorphic-card p-8 rounded-3xl h-48 flex flex-col justify-end">
              <p className="text-sm text-[#616289]">Active Sessions</p>
              <h3 className="text-3xl font-bold">{connections.filter(c => c.status === 'connected').length}</h3>
            </div>
            <div className="neumorphic-card p-8 rounded-3xl h-48 flex flex-col justify-end">
              <p className="text-sm text-[#616289]">Total Connections</p>
              <h3 className="text-3xl font-bold">{connections.length}</h3>
            </div>
            <div className="neumorphic-card p-8 rounded-3xl h-48 flex flex-col justify-end">
              <p className="text-sm text-[#616289]">Status</p>
              <h3 className="text-3xl font-bold">Ready</h3>
            </div>
          </div>
        </div>
      </main>

      <ConnectionDialog
        isOpen={isDialogOpen}
        onClose={() => setIsDialogOpen(false)}
        onSave={handleSaveConnection}
      />
    </div>
  );
};
