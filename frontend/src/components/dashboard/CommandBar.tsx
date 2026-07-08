import React, { useState } from 'react';
import { useNovelStore } from '../../stores/novelStore';
import { Play, Pause, Edit3, Trash2, FolderOpen, Download, TerminalSquare, Users, Activity, HelpCircle, FileDown } from 'lucide-react';

import { ResumeNovel, PauseNovel } from '../../../wailsjs/go/main/App';

export const CommandBar: React.FC = () => {
  const { snapshot } = useNovelStore();
  const [steerInput, setSteerInput] = useState('');

  const handleStartPause = async () => {
    try {
      if (snapshot.IsRunning) {
        await PauseNovel();
      } else {
        await ResumeNovel('');
      }
    } catch (err) {
      console.error('Lỗi khi Bắt đầu/Tạm dừng:', err);
      alert('Không thể bắt đầu/tạm dừng dự án: ' + err);
    }
  };

  const handleSteer = () => {
    if (!steerInput.trim()) return;
    // Call backend API (CoCreate/Steer)
    setSteerInput('');
  };

  const handleOpenFolder = () => {
    const win = window as any;
    if (win.go && win.go.main && win.go.main.App && win.go.main.App.OpenFolder) {
      win.go.main.App.OpenFolder().catch(console.error);
    } else {
      // Fallback
    }
  };

  return (
    <div className="bg-[#0A0A0A] border border-[#262626] rounded-lg p-3 shadow-lg flex items-center gap-3">
      {/* Control Buttons */}
      <div className="flex items-center gap-2 pr-3 border-r border-[#262626]">
        <button
          onClick={handleStartPause}
          className={`flex items-center justify-center gap-2 px-4 py-2 rounded-md font-bold text-sm font-mono transition-all border ${
            snapshot.IsRunning 
              ? 'bg-amber-500/10 text-amber-500 border-amber-500/30 hover:bg-amber-500/20 hover:border-amber-500/50' 
              : 'bg-emerald-500/10 text-emerald-500 border-emerald-500/30 hover:bg-emerald-500/20 hover:border-emerald-500/50'
          }`}
          title="Bắt đầu / Tạm dừng hệ thống"
        >
          {snapshot.IsRunning ? <Pause size={16} /> : <Play size={16} />}
        </button>

        <button 
          className="p-2 rounded-md bg-[#121212] border border-[#262626] text-slate-400 hover:text-amber-400 hover:border-amber-400/30 transition-colors"
          title="cocreate - Dừng để lên kế hoạch tiếp theo"
        >
          <Users size={16} />
        </button>

        <button 
          className="p-2 rounded-md bg-[#121212] border border-[#262626] text-slate-400 hover:text-cyan-400 hover:border-cyan-400/30 transition-colors"
          title="diag - Chẩn đoán dự án"
        >
          <Activity size={16} />
        </button>

        <button 
          className="p-2 rounded-md bg-[#121212] border border-[#262626] text-slate-400 hover:text-indigo-400 hover:border-indigo-400/30 transition-colors"
          title="import - Phân tích truyện ngoài để viết tiếp"
        >
          <FileDown size={16} />
        </button>

        <button 
          className="p-2 rounded-md bg-[#121212] border border-[#262626] text-slate-400 hover:text-fuchsia-400 hover:border-fuchsia-400/30 transition-colors"
          title="importsim - Nhập hồ sơ mô phỏng có sẵn"
        >
          <FileDown size={16} /> {/* Should probably use a different icon but let's reuse */}
        </button>

        <button 
          className="p-2 rounded-md bg-[#121212] border border-[#262626] text-slate-400 hover:text-pink-400 hover:border-pink-400/30 transition-colors"
          title="model - Đổi model"
        >
          <Edit3 size={16} />
        </button>

        <button 
          className="p-2 rounded-md bg-[#121212] border border-[#262626] text-slate-400 hover:text-blue-500 hover:border-blue-500/30 transition-colors"
          title="simulate - Đọc ./simulate để tạo/cập nhật hồ sơ"
        >
          <Activity size={16} /> 
        </button>

        <button 
          className="p-2 rounded-md bg-[#121212] border border-[#262626] text-slate-400 hover:text-purple-400 hover:border-purple-400/30 transition-colors"
          title="export - Xuất bản (TXT/EPUB)"
        >
          <Download size={16} />
        </button>
        
        <button 
          className="p-2 rounded-md bg-[#121212] border border-[#262626] text-slate-400 hover:text-emerald-400 hover:border-emerald-400/30 transition-colors"
          title="help - Xem danh sách lệnh"
        >
          <HelpCircle size={16} />
        </button>

        <div className="w-px h-6 bg-[#262626] mx-1"></div>

        <button 
          onClick={handleOpenFolder}
          className="p-2 rounded-md bg-[#121212] border border-[#262626] text-slate-400 hover:text-blue-400 hover:border-blue-400/30 transition-colors"
          title="Mở thư mục dự án"
        >
          <FolderOpen size={16} />
        </button>

        <button 
          className="p-2 rounded-md bg-[#121212] border border-[#262626] text-slate-400 hover:text-red-400 hover:border-red-400/30 transition-colors"
          title="Xoá bộ nhớ đệm (Clear Cache)"
        >
          <Trash2 size={16} />
        </button>
      </div>

      {/* CLI Input replacement */}
      <div className="flex-1 flex items-center gap-2 relative">
        <div className="absolute left-3 text-slate-500">
          <TerminalSquare size={16} />
        </div>
        <input
          type="text"
          value={steerInput}
          onChange={(e) => setSteerInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSteer()}
          placeholder="Nhập yêu cầu can thiệp (Steer) rồi nhấn Enter..."
          className="w-full bg-[#121212] border border-[#262626] rounded-md py-2 pl-9 pr-24 text-sm font-mono text-slate-200 focus:outline-none focus:border-blue-500/50 placeholder:text-slate-600"
        />
        <button
          onClick={handleSteer}
          disabled={!steerInput.trim()}
          className="absolute right-1 top-1 bottom-1 px-3 bg-blue-600 hover:bg-blue-500 disabled:bg-[#262626] disabled:text-slate-500 text-white text-xs font-bold rounded flex items-center gap-1 transition-colors"
        >
          <Edit3 size={14} />
          CAN THIỆP
        </button>
      </div>
    </div>
  );
};
