import React, { useState, useEffect } from 'react';
import { ShieldAlert, RefreshCw } from 'lucide-react';
import { GetConfig } from '../../wailsjs/go/main/App';

export const SettingsPage: React.FC = () => {
  const [config, setConfig] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  const fetchConfig = async () => {
    setLoading(true);
    try {
      const cfg = await GetConfig();
      setConfig(cfg);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConfig();
  }, []);

  return (
    <div className="h-full flex flex-col bg-[var(--color-bg-primary)] text-white max-w-3xl mx-auto w-full pt-4">
      <div className="pb-4 border-b border-[var(--color-border)] mb-6 flex justify-between items-end">
        <div>
          <h2 className="text-xl font-bold font-mono">System Config</h2>
          <p className="text-sm text-[var(--color-text-secondary)] mt-1">Cấu hình được nạp từ file hệ thống (Read-only)</p>
        </div>
        <button 
          onClick={fetchConfig}
          className="flex items-center gap-2 px-3 py-1.5 text-sm bg-slate-800 hover:bg-slate-700 rounded border border-slate-700 transition-colors cursor-pointer"
        >
          <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          Reload
        </button>
      </div>
      
      <div className="mb-6 p-4 bg-orange-500/10 border border-orange-500/20 rounded-lg flex gap-3 text-orange-200 text-sm">
        <ShieldAlert size={20} className="shrink-0 text-orange-400" />
        <p>
          <strong>Lưu ý:</strong> Để đảm bảo an toàn cho dự án CLI cũ của bạn, giao diện Settings này hiện đang bị khoá ở chế độ <strong>Chỉ xem</strong>. Vui lòng sửa trực tiếp file config.yaml nếu cần thiết.
        </p>
      </div>
      
      <div className="space-y-6 flex-1 overflow-y-auto pb-8">
        {loading ? (
          <div className="animate-pulse space-y-4">
            <div className="h-32 bg-[var(--color-bg-panel)] rounded-lg"></div>
            <div className="h-32 bg-[var(--color-bg-panel)] rounded-lg"></div>
          </div>
        ) : config ? (
          <>
            <section className="bg-[var(--color-bg-panel)] rounded-lg border border-[var(--color-border)] overflow-hidden">
              <div className="px-6 py-4 border-b border-[var(--color-border)] bg-[var(--color-bg-hover)]">
                <h3 className="font-semibold text-blue-400 font-mono">Current Provider (Active)</h3>
              </div>
              <div className="p-6">
                <div className="grid grid-cols-3 gap-y-4 text-sm">
                  <div className="text-[var(--color-text-secondary)]">Provider</div>
                  <div className="col-span-2 font-mono">{config.Provider?.Name || 'gemini'}</div>
                  
                  <div className="text-[var(--color-text-secondary)]">Model</div>
                  <div className="col-span-2 font-mono text-blue-300">{config.Provider?.Model || 'gemini-1.5-pro-latest'}</div>
                  
                  <div className="text-[var(--color-text-secondary)]">Max Tokens</div>
                  <div className="col-span-2 font-mono">{config.Provider?.MaxTokens || 8192}</div>
                </div>
              </div>
            </section>

            <section className="bg-[var(--color-bg-panel)] rounded-lg border border-[var(--color-border)] overflow-hidden">
              <div className="px-6 py-4 border-b border-[var(--color-border)] bg-[var(--color-bg-hover)]">
                <h3 className="font-semibold text-blue-400 font-mono">Raw Configuration (JSON)</h3>
              </div>
              <div className="p-0">
                <pre className="p-4 text-xs font-mono text-slate-300 overflow-x-auto whitespace-pre-wrap leading-relaxed">
                  {JSON.stringify(config, null, 2)}
                </pre>
              </div>
            </section>
          </>
        ) : (
          <div className="text-center p-8 text-slate-500">
            Không thể tải được cấu hình.
          </div>
        )}
      </div>
    </div>
  );
};
