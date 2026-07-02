import React from 'react';
import { useNovelStore } from '../../stores/novelStore';
import { Target } from 'lucide-react';

export const ProgressBar: React.FC = () => {
  const { snapshot } = useNovelStore();
  
  const percentage = snapshot.totalChapters > 0 
    ? Math.round((snapshot.currentChapter / snapshot.totalChapters) * 100) 
    : 0;

  return (
    <div className="bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-xl p-5 shadow-lg relative overflow-hidden">
      <div className="flex items-center justify-between mb-4 relative z-10">
        <div className="flex items-center gap-2">
          <Target size={18} className="text-blue-500" />
          <h3 className="font-semibold text-white tracking-wide">Overall Progress</h3>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm font-medium text-slate-400 font-mono">
            Chapter {snapshot.currentChapter} / {snapshot.totalChapters || '?'}
          </span>
          <span className="bg-blue-500/20 text-blue-400 font-bold px-2 py-0.5 rounded text-sm font-mono border border-blue-500/30">
            {percentage}%
          </span>
        </div>
      </div>
      
      <div className="w-full bg-[var(--color-bg-primary)] rounded-full h-3 border border-[var(--color-border)] relative z-10 overflow-hidden shadow-inner">
        <div 
          className="bg-gradient-to-r from-blue-600 via-blue-500 to-indigo-500 h-full rounded-full transition-all duration-700 ease-out relative" 
          style={{ width: `${percentage}%` }}
        >
          <div className="absolute inset-0 bg-white/20 animate-pulse"></div>
        </div>
      </div>
      
      {/* Background glow */}
      <div 
        className="absolute top-1/2 left-0 h-20 w-1/2 bg-blue-500/10 blur-[50px] -translate-y-1/2 rounded-full pointer-events-none transition-all duration-700"
        style={{ left: `${percentage / 2}%`, transform: 'translate(-50%, -50%)' }}
      ></div>
    </div>
  );
};
