import React, { useState } from 'react';
import { Play, FolderOpen, Loader2, Sparkles, BookOpen } from 'lucide-react';
import { StartNovel, ResumeNovel } from '../../wailsjs/go/main/App';
import { useNavigate } from 'react-router-dom';

export const WelcomePage: React.FC = () => {
  const [prompt, setPrompt] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleStart = async () => {
    if (!prompt.trim()) return;
    setLoading(true);
    try {
      await StartNovel(prompt);
      navigate('/');
    } catch (err) {
      console.error(err);
      alert('Error starting novel: ' + err);
    } finally {
      setLoading(false);
    }
  };

  const handleResume = async () => {
    setLoading(true);
    try {
      await ResumeNovel('');
      navigate('/');
    } catch (err) {
      console.error(err);
      alert('Error resuming novel: ' + err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-8 bg-[var(--color-bg-primary)] text-white w-full">
      <div className="max-w-2xl w-full flex flex-col items-center gap-10">
        
        <div className="text-center space-y-4">
          <div className="inline-flex items-center justify-center p-4 bg-blue-500/10 rounded-full mb-2">
            <BookOpen className="text-blue-500 w-10 h-10" />
          </div>
          <h1 className="text-5xl font-bold font-mono tracking-tighter">
            AINovel <span className="bg-gradient-to-r from-blue-500 to-indigo-500 bg-clip-text text-transparent">Writer</span>
          </h1>
          <p className="text-slate-400 font-sans text-lg max-w-md mx-auto">
            Môi trường sáng tác tiểu thuyết bằng AI chuyên nghiệp, tự động hoá hoàn toàn quy trình.
          </p>
        </div>

        <div className="w-full bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-xl p-8 shadow-2xl">
          <div className="space-y-6">
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
                className="w-full h-36 p-4 bg-[var(--color-bg-panel)] border border-[var(--color-border)] rounded-lg text-[var(--color-text-primary)] font-mono text-sm resize-none focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all disabled:opacity-50"
              ></textarea>
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <button 
                onClick={handleStart}
                disabled={loading || !prompt.trim()}
                className="flex items-center justify-center gap-2 bg-blue-600 text-white px-6 py-4 rounded-lg font-bold hover:bg-blue-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer shadow-lg shadow-blue-500/20"
              >
                {loading ? <Loader2 size={20} className="animate-spin" /> : <Play size={20} />}
                Bắt đầu sáng tác
              </button>
              
              <button 
                onClick={handleResume}
                disabled={loading}
                className="flex items-center justify-center gap-2 bg-[var(--color-bg-hover)] text-white px-6 py-4 rounded-lg font-medium border border-[var(--color-border)] hover:bg-slate-800 hover:border-slate-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              >
                {loading ? <Loader2 size={20} className="animate-spin" /> : <FolderOpen size={20} />}
                Tiếp tục dự án cũ
              </button>
            </div>
          </div>
        </div>
        
        <div className="text-xs text-slate-600 font-mono">
          © 2026 AINovel Writer CLI - GUI Version
        </div>
      </div>
    </div>
  );
};
