import React from 'react';
import { useNovelStore } from '../../stores/novelStore';
import { Loader2, CheckCircle2, AlertCircle, Clock } from 'lucide-react';

export const AgentCards: React.FC = () => {
  const { snapshot } = useNovelStore();

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running': return <Loader2 className="animate-spin text-blue-500" size={20} />;
      case 'done': return <CheckCircle2 className="text-[var(--color-success)]" size={20} />;
      case 'error': return <AlertCircle className="text-[var(--color-error)]" size={20} />;
      default: return <Clock className="text-[var(--color-text-secondary)]" size={20} />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'border-blue-500/50 bg-blue-500/5 shadow-[0_0_15px_rgba(59,130,246,0.15)]';
      case 'error': return 'border-[var(--color-error)] bg-[var(--color-error)]/5';
      case 'done': return 'border-[var(--color-success)]/50 bg-[var(--color-success)]/5';
      default: return 'border-[var(--color-border)] bg-[var(--color-bg-panel)]';
    }
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {snapshot.agents.map((agent) => (
        <div 
          key={agent.name} 
          className={`rounded-xl p-5 border ${getStatusColor(agent.status)} transition-all duration-300 backdrop-blur-sm relative overflow-hidden`}
        >
          {agent.status === 'running' && (
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-600 to-indigo-400 animate-pulse" />
          )}
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-semibold text-white font-mono">{agent.name}</h3>
            {getStatusIcon(agent.status)}
          </div>
          <p className="text-sm text-[var(--color-text-secondary)] capitalize font-medium">{agent.status}</p>
          {agent.message && (
            <p className="text-xs text-[var(--color-text-primary)] mt-3 opacity-90 truncate bg-[var(--color-bg-primary)]/50 p-2 rounded-md font-mono border border-[var(--color-border)]/50">{agent.message}</p>
          )}
        </div>
      ))}
    </div>
  );
};
