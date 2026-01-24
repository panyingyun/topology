import React, { useState, useEffect } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Connection, Table } from '../types';
import { connectionService } from '../services/connectionService';
import { tableService } from '../services/tableService';
import { Sidebar } from '../components/Sidebar';
import { Header } from '../components/Header';

export const MainLayout: React.FC = () => {
  const [connections, setConnections] = useState<Connection[]>([]);
  const [selectedConnectionId, setSelectedConnectionId] = useState<string | undefined>();
  const [selectedTable, setSelectedTable] = useState<string | undefined>();
  const [tables, setTables] = useState<Table[]>([]);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    loadConnections();
  }, []);

  useEffect(() => {
    if (selectedConnectionId) {
      loadTables(selectedConnectionId);
    }
  }, [selectedConnectionId]);

  const loadConnections = async () => {
    const conns = await connectionService.getConnections();
    setConnections(conns);
    if (conns.length > 0 && !selectedConnectionId) {
      setSelectedConnectionId(conns[0].id);
    }
  };

  const loadTables = async (connectionId: string) => {
    const tableList = await tableService.getTables(connectionId);
    setTables(tableList);
  };

  const handleConnectionSelect = (connectionId: string) => {
    setSelectedConnectionId(connectionId);
    setSelectedTable(undefined);
  };

  const handleTableSelect = (tableName: string) => {
    setSelectedTable(tableName);
    if (selectedConnectionId) {
      navigate(`/table/${selectedConnectionId}/${tableName}`);
    }
  };

  const handleNewConnection = () => {
    navigate('/');
  };

  const handleNewTable = () => {
    if (selectedConnectionId) {
      navigate(`/designer/${selectedConnectionId}`);
    }
  };

  const handleSettings = () => {
    console.log('Settings clicked');
  };

  const currentConnection = connections.find(c => c.id === selectedConnectionId);

  return (
    <div className="flex flex-col h-screen overflow-hidden">
      <Header
        currentConnection={currentConnection}
        currentTable={selectedTable}
        onNewConnection={handleNewConnection}
      />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar
          connections={connections}
          selectedConnectionId={selectedConnectionId}
          selectedTable={selectedTable}
          tables={tables}
          onConnectionSelect={handleConnectionSelect}
          onTableSelect={handleTableSelect}
          onNewConnection={handleNewConnection}
          onNewTable={handleNewTable}
          onSettings={handleSettings}
        />
        <main className="flex-1 flex flex-col overflow-hidden bg-background-light dark:bg-background-dark">
          <Outlet context={{ connectionId: selectedConnectionId, tableName: selectedTable }} />
        </main>
      </div>
    </div>
  );
};
