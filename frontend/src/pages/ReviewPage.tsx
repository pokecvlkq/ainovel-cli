import React from 'react';
import { useParams } from 'react-router-dom';
import ReactDiffViewer from 'react-diff-viewer-continued';
import { Check, X, Edit3, Eye } from 'lucide-react';

export const ReviewPage: React.FC = () => {
  const { chapter } = useParams();

  const oldCode = `# Chapter ${chapter || '1'}\n\nOnce upon a time in a small village, there lived a boy.\nHe liked to play all day.`;
  const newCode = `# Chapter ${chapter || '1'}\n\nOnce upon a time in a bustling town, there lived a young man.\nHe worked hard every day.`;

  return (
    <div className="h-full flex flex-col bg-[var(--color-bg-primary)]">
      <div className="flex items-center justify-between pb-4 border-b border-[var(--color-border)] mb-4">
        <h2 className="text-xl font-bold text-white flex items-center gap-3 font-mono">
          <div className="bg-indigo-500/10 p-1.5 rounded-lg border border-indigo-500/20">
            <Eye size={20} className="text-indigo-400" />
          </div>
          Review Chapter {chapter}
          <span className="text-sm font-normal text-slate-500 bg-slate-800 px-2 py-0.5 rounded-full border border-slate-700 ml-2">Diff Mode</span>
        </h2>
        <div className="flex gap-3">
          <button className="flex items-center gap-2 bg-green-600/10 text-green-500 border border-green-500/30 px-5 py-2 rounded-lg font-bold hover:bg-green-600 hover:text-white transition-all text-sm shadow-sm">
            <Check size={16} /> Approve
          </button>
          <button className="flex items-center gap-2 bg-red-600/10 text-red-500 border border-red-500/30 px-5 py-2 rounded-lg font-bold hover:bg-red-600 hover:text-white transition-all text-sm shadow-sm">
            <X size={16} /> Reject
          </button>
          <button className="flex items-center gap-2 bg-[var(--color-bg-panel)] text-white border border-[var(--color-border)] px-4 py-2 rounded-lg font-medium hover:bg-[var(--color-bg-hover)] transition-all text-sm ml-2">
            <Edit3 size={16} /> Edit
          </button>
        </div>
      </div>
      
      <div className="flex-1 overflow-auto border border-[var(--color-border)] rounded-xl bg-[#1e1e1e] shadow-inner">
        <ReactDiffViewer 
          oldValue={oldCode} 
          newValue={newCode} 
          splitView={true} 
          useDarkTheme={true}
          leftTitle="Draft (Writer Agent)"
          rightTitle="Final (Editor Agent)"
          styles={{
            variables: {
              dark: {
                diffViewerBackground: '#1e1e1e',
                diffViewerTitleBackground: '#2d2d30',
                diffViewerTitleColor: '#cccccc',
                diffViewerTitleBorderColor: '#3e3e42',
              }
            }
          }}
        />
      </div>
    </div>
  );
};
