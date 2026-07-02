import React from 'react';
import { useNovelStore } from '../../stores/novelStore';
import { CheckCircle2, CircleDashed } from 'lucide-react';

export const StatusBar: React.FC = () => {
  const { snapshot } = useNovelStore();

  return (
    <div className="h-7 border-t border-[var(--color-border)] bg-[var(--color-bg-primary)] flex items-center justify-between px-3 text-[11px] font-mono text-slate-400 select-none">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-1.5 hover:text-slate-200 transition-colors cursor-default">
          {snapshot.isWriting ? (
            <CircleDashed size={12} className="animate-spin text-blue-400" />
          ) : (
            <CheckCircle2 size={12} className="text-green-500" />
          )}
          <span>{snapshot.isWriting ? 'Writing in progress...' : 'Ready'}</span>
        </div>
        <div className="w-px h-3 bg-[var(--color-border)]"></div>
        <span className="hover:text-slate-200 transition-colors cursor-default">
          Ch {snapshot.currentChapter || 0} / {snapshot.totalChapters || 0}
        </span>
      </div>
      
      <div className="flex items-center gap-4">
        <span className="hover:text-slate-200 transition-colors cursor-default" title="Total Tokens Generated">
          {snapshot.totalTokens?.toLocaleString() || 0} tokens
        </span>
        <div className="w-px h-3 bg-[var(--color-border)]"></div>
        <span className="hover:text-slate-200 transition-colors cursor-default">UTF-8</span>
        <span className="hover:text-slate-200 transition-colors cursor-default">Go + React</span>
      </div>
    </div>
  );
};
