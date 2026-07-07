import React, { useEffect, useState } from 'react';
import { Play, FolderOpen, Loader2, Sparkles, BookOpen, AlertCircle, CheckCircle2, FolderSearch } from 'lucide-react';
import { StartNovel, ResumeNovel, GetSnapshot, SelectProjectDir } from '../../wailsjs/go/main/App';
import { useNavigate } from 'react-router-dom';

export const WelcomePage: React.FC = () => {
  const [prompt, setPrompt] = useState('');
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [detectedProject, setDetectedProject] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    GetSnapshot()
      .then((snap) => {
        if (snap && snap.NovelName) {
          setDetectedProject(snap.NovelName);
        }
      })
      .catch(() => {});
  }, []);

  const handleStart = async () => {
    if (!prompt.trim()) return;
    setLoading(true);
    setErrorMsg(null);
    try {
      await StartNovel(prompt);
      navigate('/');
    } catch (err: any) {
      console.error(err);
      setErrorMsg(err?.toString() || 'Lỗi không xác định khi khởi tạo dự án');
    } finally {
      setLoading(false);
    }
  };

  const handleResume = async (customDir: string = '') => {
    setLoading(true);
    setErrorMsg(null);
    try {
      await ResumeNovel(customDir);
      navigate('/');
    } catch (err: any) {
      console.error(err);
      setErrorMsg(err?.toString() || 'Lỗi khôi phục dự án');
    } finally {
      setLoading(false);
    }
  };

  const handleBrowseFolder = async () => {
    try {
      const dir = await SelectProjectDir();
      if (dir) {
        await handleResume(dir);
      }
    } catch (err: any) {
      setErrorMsg('Lỗi chọn thư mục: ' + err);
    }
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-8 bg-[var(--color-bg-primary)] text-white w-full">
      <div className="max-w-2xl w-full flex flex-col items-center gap-8">
        
        <div className="text-center space-y-4">
          <div className="inline-flex items-center justify-center p-4 bg-blue-500/10 rounded-full mb-2 border border-blue-500/20 shadow-lg shadow-blue-500/10">
            <BookOpen className="text-blue-400 w-10 h-10" />
          </div>
          <h1 className="text-5xl font-bold font-mono tracking-tighter">
            AINovel <span className="bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">Writer</span>
          </h1>
          <p className="text-slate-400 font-sans text-lg max-w-md mx-auto">
            Môi trường sáng tác tiểu thuyết bằng AI chuyên nghiệp, tự động hoá hoàn toàn quy trình.
          </p>
        </div>

        {errorMsg && (
          <div className="w-full bg-red-950/40 border border-red-500/50 rounded-lg p-4 text-red-200 text-sm font-mono flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
            <div className="whitespace-pre-wrap">{errorMsg}</div>
          </div>
        )}

        {detectedProject && (
          <div className="w-full bg-blue-950/30 border border-blue-500/30 rounded-xl p-5 flex items-center justify-between shadow-lg">
            <div className="flex items-center gap-3">
              <CheckCircle2 className="w-6 h-6 text-emerald-400 flex-shrink-0" />
              <div>
                <div className="text-xs text-slate-400 font-mono">Đã phát hiện dự án CLI dở dang:</div>
                <div className="text-lg font-bold text-white font-sans">{detectedProject}</div>
              </div>
            </div>
            <button
              onClick={() => handleResume('')}
              disabled={loading}
              className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-lg text-sm transition-colors cursor-pointer flex items-center gap-2 shadow-md shadow-emerald-600/20"
            >
              {loading ? <Loader2 size={16} className="animate-spin" /> : <FolderOpen size={16} />}
              Mở Dự Án Này
            </button>
          </div>
        )}

        <div className="w-full bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-xl p-8 shadow-2xl space-y-6">
          <div>
            <label className="flex items-center gap-2 text-sm font-medium text-slate-300 mb-3 font-mono">
              <Sparkles size={16} className="text-blue-400" />
              Prompt Khởi Tạo Truyện Mới
            </label>
            <textarea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="Ví dụ: Truyện tiên hiệp về một thiếu niên có khả năng nhìn thấu tuổi thọ của vạn vật..."
              disabled={loading}
              className="w-full h-32 p-4 bg-[var(--color-bg-panel)] border border-[var(--color-border)] rounded-lg text-[var(--color-text-primary)] font-mono text-sm resize-none focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all disabled:opacity-50"
            ></textarea>
          </div>
          
          <div className="grid grid-cols-3 gap-3">
            <button 
              onClick={handleStart}
              disabled={loading || !prompt.trim()}
              className="flex items-center justify-center gap-2 bg-blue-600 text-white px-4 py-3.5 rounded-lg font-bold hover:bg-blue-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer shadow-lg shadow-blue-500/20 text-sm"
            >
              {loading ? <Loader2 size={18} className="animate-spin" /> : <Play size={18} />}
              Tạo Mới
            </button>
            
            <button 
              onClick={() => handleResume('')}
              disabled={loading}
              className="flex items-center justify-center gap-2 bg-[var(--color-bg-hover)] text-white px-4 py-3.5 rounded-lg font-medium border border-[var(--color-border)] hover:bg-slate-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer text-sm"
            >
              {loading ? <Loader2 size={18} className="animate-spin" /> : <FolderOpen size={18} />}
              Khôi Phục Mặc Định
            </button>

            <button 
              onClick={handleBrowseFolder}
              disabled={loading}
              className="flex items-center justify-center gap-2 bg-slate-800 text-slate-200 px-4 py-3.5 rounded-lg font-medium border border-slate-700 hover:bg-slate-700 hover:text-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer text-sm"
            >
              <FolderSearch size={18} />
              Chọn Thư Mục...
            </button>
          </div>
        </div>
        
        <div className="text-xs text-slate-500 font-mono">
          © 2026 AINovel Writer CLI - GUI Engine
        </div>
      </div>
    </div>
  );
};

