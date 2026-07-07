import React from 'react';
import { useNovelStore } from '../../stores/novelStore';
import { Loader2, CheckCircle2, AlertCircle, Clock } from 'lucide-react';

export const AgentCards: React.FC = () => {
  const { snapshot } = useNovelStore();

  // Backend trả AgentSnapshot[] với PascalCase fields: Name, State, Summary, Tool, TaskKind
  const agents = snapshot.Agents || [];

  const getStatusIcon = (state: string) => {
    switch (state) {
      case 'running': return <Loader2 className="animate-spin text-blue-500" size={20} />;
      case 'done': return <CheckCircle2 className="text-[var(--color-success)]" size={20} />;
      case 'error': return <AlertCircle className="text-[var(--color-error)]" size={20} />;
      default: return <Clock className="text-[var(--color-text-secondary)]" size={20} />;
    }
  };

  const getStatusColor = (state: string) => {
    switch (state) {
      case 'running': return 'border-blue-500/50 bg-blue-500/5 shadow-[0_0_15px_rgba(59,130,246,0.15)]';
      case 'error': return 'border-[var(--color-error)] bg-[var(--color-error)]/5';
      case 'done': return 'border-[var(--color-success)]/50 bg-[var(--color-success)]/5';
      default: return 'border-[var(--color-border)] bg-[var(--color-bg-panel)]';
    }
  };

  const predefinedAgents = ['coordinator', 'architect', 'writer', 'editor'];

  const displayAgents = predefinedAgents.map(agentName => {
    const backendAgent = agents.find(a => a.Name?.toLowerCase() === agentName);
    return backendAgent 
      ? { name: backendAgent.Name, state: backendAgent.State, summary: backendAgent.Summary, tool: backendAgent.Tool }
      : { name: agentName.charAt(0).toUpperCase() + agentName.slice(1), state: 'idle', summary: '', tool: '' };
  });

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {displayAgents.map((agent) => (
        <div 
          key={agent.name} 
          className={`rounded-xl p-5 border ${getStatusColor(agent.state)} transition-all duration-300 backdrop-blur-sm relative overflow-hidden`}
        >
          {agent.state === 'running' && (
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-600 to-indigo-400 animate-pulse" />
          )}
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-semibold text-white font-mono">{agent.name}</h3>
            {getStatusIcon(agent.state)}
          </div>
          <p className="text-sm text-[var(--color-text-secondary)] capitalize font-medium">{agent.state || 'idle'}</p>
          {agent.summary && (
            <p className="text-xs text-[var(--color-text-primary)] mt-3 opacity-90 truncate bg-[var(--color-bg-primary)]/50 p-2 rounded-md font-mono border border-[var(--color-border)]/50">{agent.summary}</p>
          )}
          {agent.tool && (
            <p className="text-[10px] text-slate-500 mt-1 font-mono truncate">🔧 {agent.tool}</p>
          )}
        </div>
      ))}
    </div>
  );
};
