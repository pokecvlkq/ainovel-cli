import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useNovelStore } from '../stores/novelStore';
import { DashboardLeftPanel } from '../components/dashboard/DashboardLeftPanel';
import { DashboardCenterPanel } from '../components/dashboard/DashboardCenterPanel';
import { DashboardRightPanel } from '../components/dashboard/DashboardRightPanel';
import { BookOpen, FolderOpen, Plus, ArrowRight } from 'lucide-react';

export const DashboardPage: React.FC = () => {
  const { snapshot } = useNovelStore();
  const navigate = useNavigate();
  const hasProject = !!snapshot.NovelName;

  if (!hasProject) {
    return (
      <div className="h-full flex flex-col items-center justify-center gap-8 p-8 bg-black">
        <div className="text-center space-y-4">
          <div className="inline-flex items-center justify-center p-5 bg-blue-500/10 rounded-full border border-blue-500/20 shadow-[0_0_30px_rgba(59,130,246,0.15)] mb-2">
            <BookOpen className="text-blue-400 w-12 h-12" />
          </div>
          <h2 className="text-3xl font-bold text-white font-mono">
            Chưa có dự án nào
          </h2>
          <p className="text-slate-400 text-base max-w-md mx-auto">
            Tạo dự án mới hoặc khôi phục dự án CLI đang làm dở để bắt đầu sáng tác.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 w-full max-w-lg">
          <button
            onClick={() => navigate('/welcome')}
            className="flex items-center justify-center gap-3 bg-blue-600 hover:bg-blue-500 text-white px-6 py-4 rounded-xl font-bold transition-all cursor-pointer shadow-lg shadow-blue-600/20 hover:shadow-blue-500/30 hover:scale-[1.02] active:scale-[0.98] group border border-blue-400/50"
          >
            <Plus size={22} />
            <span>Tạo Dự Án Mới</span>
            <ArrowRight size={16} className="opacity-0 group-hover:opacity-100 transition-opacity -ml-1" />
          </button>
          
          <button
            onClick={() => navigate('/welcome')}
            className="flex items-center justify-center gap-3 bg-[#0A0A0A] hover:bg-[#121212] text-white px-6 py-4 rounded-xl font-medium border border-[#262626] hover:border-blue-500/30 transition-all cursor-pointer hover:scale-[1.02] active:scale-[0.98]"
          >
            <FolderOpen size={22} className="text-blue-400" />
            <span>Khôi Phục Dự Án</span>
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full w-full bg-black p-3 gap-3 grid grid-cols-12 overflow-hidden">
      {/* Left Panel - 3 cols (25%) */}
      <div className="col-span-3 h-full min-h-0 overflow-y-auto pr-1">
        <DashboardLeftPanel />
      </div>

      {/* Center Panel - 6 cols (50%) */}
      <div className="col-span-6 h-full min-h-0 flex flex-col gap-3">
        <DashboardCenterPanel />
      </div>

      {/* Right Panel - 3 cols (25%) */}
      <div className="col-span-3 h-full min-h-0 overflow-y-auto pl-1">
        <DashboardRightPanel />
      </div>
    </div>
  );
};
