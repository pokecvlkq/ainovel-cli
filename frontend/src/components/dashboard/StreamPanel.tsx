import React, { useEffect, useRef } from 'react';
import { useNovelStore } from '../../stores/novelStore';
import { Activity } from 'lucide-react';

export const StreamPanel: React.FC = () => {
  const { streamBuffer } = useNovelStore();
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [streamBuffer]);

  return (
    <div className="bg-[var(--color-bg-primary)] border border-[var(--color-border)] rounded-xl flex flex-col h-full min-h-[300px] overflow-hidden shadow-lg relative">
      <div className="bg-[var(--color-bg-secondary)] px-4 py-3 border-b border-[var(--color-border)] flex items-center gap-2">
        <Activity size={16} className="text-green-500" />
        <h3 className="font-semibold text-white text-sm font-mono tracking-wide">Live Output Stream</h3>
      </div>
      <div className="flex-1 overflow-y-auto bg-[#030303] p-4 relative scroll-smooth group">
        <pre className="text-[13px] text-green-400 font-mono whitespace-pre-wrap break-words leading-relaxed selection:bg-green-500/30">
          {streamBuffer || <span className="text-[var(--color-text-muted)] italic opacity-50">Waiting for agent to generate content...</span>}
        </pre>
        {streamBuffer && (
          <div className="absolute bottom-4 left-4 flex items-center gap-2 opacity-50 transition-opacity">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
            </span>
            <span className="text-[10px] text-green-500 font-mono uppercase tracking-widest">Receiving</span>
          </div>
        )}
        <div ref={bottomRef} className="h-4" />
      </div>
    </div>
  );
};
