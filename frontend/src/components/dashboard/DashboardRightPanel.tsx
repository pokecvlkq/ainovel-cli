import React from 'react';
import { useNovelStore } from '../../stores/novelStore';

export const DashboardRightPanel: React.FC = () => {
  const { snapshot } = useNovelStore();

  const renderSectionHeader = (title: string, colorClass: string = "text-emerald-500", bgClass: string = "bg-emerald-500") => (
    <div className="flex items-center gap-2 mb-3">
      <div className={`w-1.5 h-1.5 ${bgClass} rounded-full animate-pulse`} />
      <h3 className={`${colorClass} font-bold font-mono text-sm uppercase tracking-wider`}>{title}</h3>
      <div className={`flex-1 h-px bg-gradient-to-r from-${colorClass.replace('text-', '')}/30 to-transparent`} />
    </div>
  );

  return (
    <div className="flex flex-col gap-4 h-full">
      {/* Khối Đề cương */}
      <div className="border border-[#262626] bg-[#0A0A0A] rounded-lg p-3 shadow-lg flex-1 min-h-[200px] flex flex-col">
        {renderSectionHeader('Đề cương (Outline)', 'text-emerald-500', 'bg-emerald-500')}
        <div className="flex-1 overflow-y-auto pr-1 space-y-2">
          {snapshot.Outline && snapshot.Outline.length > 0 ? (
            snapshot.Outline.map((chapter: any, idx: number) => (
              <div key={idx} className="flex gap-2 text-xs font-mono">
                <span className="text-slate-500 w-8 flex-shrink-0">{String(chapter.Chapter || idx + 1).padStart(2, '0')}</span>
                <div className="flex-1">
                  <div className="text-slate-200 truncate">{chapter.Title}</div>
                  <div className="text-slate-500 text-[10px] line-clamp-1">{chapter.CoreEvent}</div>
                </div>
              </div>
            ))
          ) : (
            <div className="text-slate-500 text-xs font-mono italic">Chưa có đề cương...</div>
          )}
        </div>
      </div>

      {/* Khối Nhân vật */}
      <div className="border border-[#262626] bg-[#0A0A0A] rounded-lg p-3 shadow-lg flex-1 min-h-[150px] flex flex-col">
        {renderSectionHeader('Nhân vật (Characters)', 'text-purple-400', 'bg-purple-400')}
        <div className="flex-1 overflow-y-auto pr-1 space-y-1">
          {snapshot.Characters && snapshot.Characters.length > 0 ? (
            snapshot.Characters.map((char: string, idx: number) => (
              <div key={idx} className="text-xs font-mono text-slate-300 flex items-start gap-2">
                <span className="text-purple-500/50">▸</span>
                <span className="line-clamp-2">{char}</span>
              </div>
            ))
          ) : (
            <div className="text-slate-500 text-xs font-mono italic">Đang tải nhân vật...</div>
          )}
        </div>
      </div>

      {/* Khối Tiền đề */}
      <div className="border border-[#262626] bg-[#0A0A0A] rounded-lg p-3 shadow-lg flex-shrink-0 max-h-[200px] flex flex-col">
        {renderSectionHeader('Bối cảnh & Tiền đề', 'text-amber-500', 'bg-amber-500')}
        <div className="text-xs font-mono text-slate-400 whitespace-pre-wrap break-words leading-relaxed overflow-y-auto pr-1 flex-1">
          {snapshot.Premise || 'Chưa thiết lập tiền đề...'}
        </div>
      </div>
    </div>
  );
};
