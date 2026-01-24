import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { MainPage } from './pages/MainPage';
import { MainLayout } from './pages/MainLayout';
import { Connections } from './pages/Connections';
import { QueryEditor } from './pages/QueryEditor';
import { TableDesigner } from './pages/TableDesigner';
import { ImportExport } from './pages/ImportExport';
import './style.css';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<MainPage />} />
        <Route path="/connections" element={<Connections />} />
        <Route element={<MainLayout />}>
          <Route path="/query/:connectionId?" element={<QueryEditor />} />
          <Route path="/table/:connectionId/:tableName" element={<QueryEditor />} />
          <Route path="/designer/:connectionId/:tableName?" element={<TableDesigner />} />
          <Route path="/import-export" element={<ImportExport />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
