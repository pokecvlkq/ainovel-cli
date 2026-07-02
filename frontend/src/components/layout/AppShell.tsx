import React from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { StatusBar } from './StatusBar';
import { useWailsEvents } from '../../hooks/useWailsEvents';

export const AppShell: React.FC = () => {
  // Initialize Wails events listener
  useWailsEvents();

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[var(--color-bg-primary)] text-[var(--color-text-primary)]">
      <Sidebar />
      <div className="flex-1 flex flex-col h-full min-w-0">
        <Header />
        <main className="flex-1 overflow-auto bg-[var(--color-bg-primary)] relative p-4">
          <Outlet />
        </main>
        <StatusBar />
      </div>
    </div>
  );
};
