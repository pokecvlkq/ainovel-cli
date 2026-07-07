import React from 'react';
import { ProgressBar } from './ProgressBar';
import { AgentCards } from './AgentCards';
import { EventLog } from './EventLog';
import { StreamPanel } from './StreamPanel';
import { CommandBar } from './CommandBar';

export const DashboardCenterPanel: React.FC = () => {
  return (
    <div className="flex flex-col gap-3 h-full overflow-hidden">
      {/* Top: Agents Status */}
      <div className="flex-shrink-0">
        <AgentCards />
      </div>

      <div className="flex-shrink-0">
        <ProgressBar />
      </div>

      {/* Middle: Logs & Stream */}
      <div className="flex-1 flex gap-3 min-h-0">
        {/* Left part of Center Panel: Event Log */}
        <div className="flex-1 flex flex-col min-w-0 border border-[#262626] bg-[#0A0A0A] rounded-lg shadow-lg overflow-hidden">
          <div className="p-2 border-b border-[#262626] bg-[#121212] flex items-center gap-2">
            <div className="w-2 h-2 bg-blue-500 rounded-full" />
            <span className="text-xs font-mono font-bold text-slate-300">SYSTEM.LOG</span>
          </div>
          <div className="flex-1 min-h-0 overflow-y-auto">
            <EventLog />
          </div>
        </div>

        {/* Right part of Center Panel: Stream Panel */}
        <div className="flex-1 flex flex-col min-w-0 border border-[#262626] bg-[#0A0A0A] rounded-lg shadow-lg overflow-hidden">
          <div className="p-2 border-b border-[#262626] bg-[#121212] flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full" />
            <span className="text-xs font-mono font-bold text-slate-300">OUTPUT.STREAM</span>
          </div>
          <div className="flex-1 min-h-0 overflow-y-auto">
            <StreamPanel />
          </div>
        </div>
      </div>

      {/* Bottom: Command Bar */}
      <div className="flex-shrink-0">
        <CommandBar />
      </div>
    </div>
  );
};
