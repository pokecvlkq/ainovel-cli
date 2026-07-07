import React from 'react';
import { useNovelStore } from '../../stores/novelStore';
import { CheckCircle2, CircleDashed, Trophy } from 'lucide-react';

export const StatusBar: React.FC = () => {
  const { snapshot, isComplete } = useNovelStore();

  return (
    <div className="h-7 border-t border-[var(--color-border)] bg-[var(--color-bg-primary)] flex items-center justify-between px-3 text-[11px] font-mono text-slate-400 select-none">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-1.5 hover:text-slate-200 transition-colors cursor-default">
          {isComplete ? (
            <Trophy size={12} className="text-yellow-400" />
          ) : snapshot.IsRunning ? (
            <CircleDashed size={12} className="animate-spin text-blue-400" />
          ) : (
            <CheckCircle2 size={12} className="text-green-500" />
          )}
          <span>
            {isComplete ? 'Hoàn thành!' : snapshot.IsRunning ? 'Đang viết...' : 'Sẵn sàng'}
          </span>
        </div>
        <div className="w-px h-3 bg-[var(--color-border)]"></div>
        <span className="hover:text-slate-200 transition-colors cursor-default">
          Ch {snapshot.CompletedCount || snapshot.CurrentChapter || 0} / {snapshot.TotalChapters || 0}
        </span>
      </div>
      
      <div className="flex items-center gap-4">
        <span className="hover:text-slate-200 transition-colors cursor-default" title="Tổng số từ">
          {(snapshot.TotalWordCount || 0).toLocaleString()} chữ
        </span>
        <div className="w-px h-3 bg-[var(--color-border)]"></div>
        {snapshot.Phase && (
          <>
            <span className="text-blue-400 hover:text-blue-300 transition-colors cursor-default">
              {snapshot.Phase}
            </span>
            <div className="w-px h-3 bg-[var(--color-border)]"></div>
          </>
        )}
        <span className="hover:text-slate-200 transition-colors cursor-default">Go + React</span>
      </div>
    </div>
  );
};
