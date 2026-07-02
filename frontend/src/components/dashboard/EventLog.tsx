import React, { useEffect, useRef } from 'react';
import { useNovelStore, NovelEvent } from '../../stores/novelStore';
import { Terminal } from 'lucide-react';

export const EventLog: React.FC = () => {
  const { events } = useNovelStore();
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [events]);

  const getColorForType = (type: string) => {
    switch(type) {
      case 'info': return 'text-[var(--color-text-secondary)]';
      case 'success': return 'text-[var(--color-success)]';
      case 'error': return 'text-[var(--color-error)]';
      case 'tool-call': return 'text-purple-400';
      default: return 'text-[var(--color-text-primary)]';
    }
  };

  return (
    <div className="bg-[var(--color-bg-primary)] border border-[var(--color-border)] rounded-xl flex flex-col h-full min-h-[300px] overflow-hidden shadow-lg relative">
      <div className="bg-[var(--color-bg-secondary)] px-4 py-3 border-b border-[var(--color-border)] flex items-center gap-2">
        <Terminal size={16} className="text-blue-500" />
        <h3 className="font-semibold text-white text-sm font-mono tracking-wide">Event Log</h3>
      </div>
      <div className="flex-1 overflow-y-auto p-4 space-y-2 font-mono text-[13px] leading-relaxed scroll-smooth bg-[#050505]">
        {events.length === 0 ? (
          <p className="text-[var(--color-text-muted)] italic text-center mt-10">Waiting for events...</p>
        ) : (
          events.map((ev: NovelEvent, i: number) => (
            <div key={i} className="flex items-start gap-3 hover:bg-[var(--color-bg-hover)] p-1.5 -mx-1.5 rounded transition-colors group">
              <span className="text-[var(--color-text-muted)] whitespace-nowrap shrink-0 group-hover:text-slate-500 transition-colors">
                [{new Date(ev.timestamp).toLocaleTimeString()}]
              </span>
              <span className="font-bold whitespace-nowrap text-blue-400 shrink-0 w-24 truncate">
                {ev.agent}
              </span>
              <span className={`${getColorForType(ev.type)} flex-1 break-all`}>
                {ev.message}
              </span>
            </div>
          ))
        )}
        <div ref={bottomRef} className="h-2" />
      </div>
    </div>
  );
};
