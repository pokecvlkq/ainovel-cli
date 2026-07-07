import React, { useEffect, useState } from 'react';
import { useNovelStore } from '../../stores/novelStore';
import { Activity, Book, Cpu } from 'lucide-react';
import { GetConfig } from '../../../wailsjs/go/main/App';

export const Header: React.FC = () => {
  const { snapshot } = useNovelStore();
  const [modelName, setModelName] = useState('...');

  useEffect(() => {
    GetConfig().then((cfg: any) => {
      // Config struct: { provider: string, model: string, ... }
      if (cfg && cfg.model) {
        setModelName(cfg.model);
      } else if (cfg && cfg.provider) {
        setModelName(cfg.provider);
      } else {
        setModelName('N/A');
      }
    }).catch(() => setModelName('N/A'));
  }, []);

  return (
    <div className="h-16 border-b border-[var(--color-border)] bg-[var(--color-bg-primary)] flex items-center justify-between px-6 shrink-0 relative z-10 shadow-sm">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 bg-[var(--color-bg-panel)] px-3 py-1.5 rounded-md border border-[var(--color-border)]">
          <Book size={16} className="text-blue-400" />
          <h1 className="text-[15px] font-bold text-white font-mono tracking-wide truncate max-w-[300px]">
            {snapshot.NovelName || 'Chưa có dự án'}
          </h1>
        </div>
        
        {snapshot.IsRunning && (
          <span className="flex items-center gap-2 text-[13px] font-bold text-blue-400 bg-blue-500/10 px-3 py-1.5 rounded border border-blue-500/20 font-mono tracking-wider shadow-[0_0_10px_rgba(59,130,246,0.2)]">
            <Activity size={14} className="animate-ping" />
            AI ĐANG VIẾT
          </span>
        )}
      </div>
      <div className="flex items-center gap-3 text-sm text-[var(--color-text-muted)] bg-[var(--color-bg-secondary)] px-3 py-1.5 rounded-full border border-[var(--color-border)]">
        <Cpu size={14} className="text-slate-400" />
        <span className="font-mono text-[13px] text-slate-300">{snapshot.ModelName || modelName}</span>
      </div>
    </div>
  );
};
