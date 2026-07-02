import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import Editor from '@monaco-editor/react';
import { Save, Download, FileText, CheckCircle2 } from 'lucide-react';

export const EditorPage: React.FC = () => {
  const { chapter } = useParams();
  const [content, setContent] = useState<string>('# Chapter ' + (chapter || '1') + '\n\nOnce upon a time...');
  const [isSaving, setIsSaving] = useState(false);
  const [isSaved, setIsSaved] = useState(false);

  const handleSave = () => {
    setIsSaving(true);
    setTimeout(() => {
      setIsSaving(false);
      setIsSaved(true);
      setTimeout(() => setIsSaved(false), 2000);
    }, 800);
  };

  return (
    <div className="h-full flex flex-col bg-[var(--color-bg-primary)]">
      <div className="flex items-center justify-between pb-4 border-b border-[var(--color-border)] mb-4">
        <h2 className="text-xl font-bold text-white flex items-center gap-3 font-mono">
          <div className="bg-blue-500/10 p-1.5 rounded-lg border border-blue-500/20">
            <FileText size={20} className="text-blue-500" />
          </div>
          Edit Chapter {chapter}
        </h2>
        <div className="flex gap-3">
          <button className="flex items-center gap-2 bg-[var(--color-bg-panel)] text-white px-4 py-2 rounded-lg font-medium border border-[var(--color-border)] hover:bg-[var(--color-bg-hover)] transition-colors text-sm">
            <Download size={16} />
            Export
          </button>
          <button 
            onClick={handleSave}
            className={`flex items-center gap-2 px-5 py-2 rounded-lg font-bold transition-all text-sm shadow-lg ${
              isSaved 
                ? 'bg-green-500 text-white shadow-green-500/20' 
                : 'bg-blue-600 text-white hover:bg-blue-500 shadow-blue-500/20'
            }`}
          >
            {isSaved ? <CheckCircle2 size={16} /> : <Save size={16} />}
            {isSaving ? 'Saving...' : isSaved ? 'Saved!' : 'Save'}
          </button>
        </div>
      </div>
      
      <div className="flex-1 border border-[var(--color-border)] rounded-xl overflow-hidden shadow-inner bg-[#1e1e1e]">
        <Editor
          height="100%"
          defaultLanguage="markdown"
          theme="vs-dark"
          value={content}
          onChange={(value) => setContent(value || '')}
          options={{
            wordWrap: 'on',
            minimap: { enabled: false },
            fontSize: 15,
            fontFamily: "'Fira Code', monospace",
            fontLigatures: true,
            padding: { top: 24, bottom: 24 },
            lineHeight: 1.6,
            scrollbar: {
              vertical: 'visible',
              horizontal: 'hidden'
            }
          }}
        />
      </div>
      <div className="pt-3 pb-1 flex justify-between items-center text-[11px] text-[var(--color-text-secondary)] font-mono uppercase tracking-wider">
        <span>Markdown</span>
        <span>
          {content.length} chars • {content.split(/\s+/).filter(w => w.length > 0).length} words
        </span>
      </div>
    </div>
  );
};
