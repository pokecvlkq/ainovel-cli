import React from 'react';
import { AgentCards } from '../components/dashboard/AgentCards';
import { ProgressBar } from '../components/dashboard/ProgressBar';
import { EventLog } from '../components/dashboard/EventLog';
import { StreamPanel } from '../components/dashboard/StreamPanel';

export const DashboardPage: React.FC = () => {
  return (
    <div className="h-full flex flex-col gap-4">
      <AgentCards />
      <ProgressBar />
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 flex-1 min-h-0">
        <EventLog />
        <StreamPanel />
      </div>
    </div>
  );
};
