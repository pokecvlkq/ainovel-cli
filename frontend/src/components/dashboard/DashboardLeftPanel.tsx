import React from 'react';
import { useNovelStore } from '../../stores/novelStore';

export const DashboardLeftPanel: React.FC = () => {
  const { snapshot, isComplete } = useNovelStore();

  const renderSectionHeader = (title: string) => (
    <div className="flex items-center gap-2 mb-3">
      <div className="w-1.5 h-1.5 bg-yellow-500 rounded-full animate-pulse" />
      <h3 className="text-yellow-500 font-bold font-mono text-sm uppercase tracking-wider">{title}</h3>
      <div className="flex-1 h-px bg-gradient-to-r from-yellow-500/30 to-transparent" />
    </div>
  );

  const renderKeyValue = (label: string, value: string | number, highlight: boolean = false) => (
    <div className="flex justify-between items-center text-xs font-mono mb-1.5">
      <span className="text-slate-400">{label}</span>
      <span className={highlight ? "text-green-400 font-bold" : "text-slate-200"}>{value}</span>
    </div>
  );

  return (
    <div className="flex flex-col gap-4 h-full">
      {/* Khối Tổng Quan */}
      <div className="border border-[#262626] bg-[#0A0A0A] rounded-lg p-3 shadow-lg flex-shrink-0">
        {renderSectionHeader('Tổng Quan')}
        <div className="space-y-1">
          {renderKeyValue('Trạng thái', isComplete ? 'Hoàn thành' : (snapshot.IsRunning ? 'Đang chạy' : 'Chờ lệnh'), snapshot.IsRunning)}
          {renderKeyValue('Giai đoạn', snapshot.Phase || 'N/A')}
          {renderKeyValue('Hoàn thành', `Chương ${snapshot.CompletedCount || 0} / ${snapshot.TotalChapters || 0}`, true)}
          {renderKeyValue('Đang viết', `Chương ${snapshot.InProgressChapter || (snapshot.CompletedCount || 0) + 1}`)}
          {renderKeyValue('Số từ', (snapshot.TotalWordCount || 0).toLocaleString())}
        </div>
      </div>

      {/* Khối Vai Trò */}
      <div className="border border-[#262626] bg-[#0A0A0A] rounded-lg p-3 shadow-lg flex-shrink-0">
        {renderSectionHeader('Vai Trò')}
        <div className="text-xs font-mono text-slate-300 whitespace-pre-wrap break-words leading-relaxed pl-3 border-l-2 border-slate-700">
          {snapshot.Flow || 'coordinator -> writer -> editor'}
        </div>
      </div>

      {/* Khối Sử Dụng (Usage) */}
      <div className="border border-[#262626] bg-[#0A0A0A] rounded-lg p-3 shadow-lg flex-shrink-0">
        {renderSectionHeader('Sử Dụng (Usage)')}
        <div className="space-y-1 mb-3">
          {renderKeyValue('Nhập vào (In)', `${((snapshot.TotalInputTokens || 0) / 1000).toFixed(1)}k`)}
          {renderKeyValue('Output (Out)', `${((snapshot.TotalOutputTokens || 0) / 1000).toFixed(1)}k`)}
          {renderKeyValue('Chi phí', `$${(snapshot.TotalCostUSD || 0).toFixed(2)}`, true)}
          {renderKeyValue('Ngân sách', `$${(snapshot.BudgetLimitUSD || 0).toFixed(2)}`)}
        </div>
        
        {/* Model Breakdown - Hardcode fallback or use real if backend provides it in the future */}
        <div className="mt-2 pt-2 border-t border-[#262626]">
          <span className="text-[10px] text-slate-500 font-mono mb-1 block">MODELS</span>
          {renderKeyValue(snapshot.ModelName || 'gemini-3.1-flash', `$${(snapshot.TotalCostUSD || 0).toFixed(2)}`)}
        </div>
      </div>

      {/* Khối Cache */}
      <div className="border border-[#262626] bg-[#0A0A0A] rounded-lg p-3 shadow-lg flex-1 min-h-0 overflow-y-auto">
        {renderSectionHeader('Bộ Nhớ Đệm (Cache)')}
        <div className="space-y-1">
          {renderKeyValue('Trúng đích', `${snapshot.ContextPercent || 0}%`, true)}
          {renderKeyValue('Tiết kiệm', `$${(snapshot.TotalSavedUSD || 0).toFixed(2)}`, true)}
          {renderKeyValue('Token Context', `${((snapshot.ContextTokens || 0) / 1000).toFixed(1)}k`)}
          {renderKeyValue('Window Size', `${((snapshot.ContextWindow || 0) / 1000).toFixed(1)}k`)}
        </div>
      </div>
      {/* Khối Tác Vụ Nhanh (Quick Actions) */}
      <div className="border border-[#262626] bg-[#0A0A0A] rounded-lg p-3 shadow-lg flex-shrink-0 mt-auto">
        {renderSectionHeader('Tác Vụ (Actions)')}
        <div className="grid grid-cols-2 gap-2 mt-2">
          <button className="bg-[#1A1A1A] hover:bg-[#2A2A2A] text-slate-300 text-xs py-1.5 px-2 rounded border border-[#333] transition-colors" title="/model">
            Đổi Model
          </button>
          <button className="bg-[#1A1A1A] hover:bg-[#2A2A2A] text-slate-300 text-xs py-1.5 px-2 rounded border border-[#333] transition-colors" title="/diag">
            Chẩn đoán
          </button>
          <button className="bg-[#1A1A1A] hover:bg-[#2A2A2A] text-slate-300 text-xs py-1.5 px-2 rounded border border-[#333] transition-colors" title="/import">
            Nhập Truyện
          </button>
          <button className="bg-[#1A1A1A] hover:bg-[#2A2A2A] text-slate-300 text-xs py-1.5 px-2 rounded border border-[#333] transition-colors" title="/plan">
            Lên Kế Hoạch
          </button>
          <button className="bg-[#1A1A1A] hover:bg-[#2A2A2A] text-slate-300 text-xs py-1.5 px-2 rounded border border-[#333] transition-colors" title="/simulate">
            Mô phỏng
          </button>
          <button className="bg-[#1A1A1A] hover:bg-[#2A2A2A] text-slate-300 text-xs py-1.5 px-2 rounded border border-[#333] transition-colors" title="/export">
            Xuất Bản
          </button>
        </div>
      </div>
    </div>
  );
};
