import React, { useState } from 'react';
import { useNovelStore } from '../../stores/novelStore';
import { AnswerQuestion } from '../../../wailsjs/go/main/App';

export const AskUserModal: React.FC = () => {
  const { askUserQuestions, setAskUserQuestions } = useNovelStore();

  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [notes, setNotes] = useState<Record<string, string>>({});

  if (!askUserQuestions || askUserQuestions.length === 0) {
    return null;
  }

  const handleSelect = (questionText: string, optionLabel: string, multiSelect: boolean) => {
    setAnswers((prev) => {
      if (multiSelect) {
        const current = prev[questionText] || '';
        let currentArray = current ? current.split(', ') : [];
        if (currentArray.includes(optionLabel)) {
          currentArray = currentArray.filter((o) => o !== optionLabel);
        } else {
          currentArray.push(optionLabel);
        }
        return { ...prev, [questionText]: currentArray.join(', ') };
      } else {
        return { ...prev, [questionText]: optionLabel };
      }
    });
  };

  const handleNoteChange = (questionText: string, text: string) => {
    setNotes((prev) => ({ ...prev, [questionText]: text }));
  };

  const handleSubmit = () => {
    AnswerQuestion({ answers, notes }).then(() => {
      setAskUserQuestions(null);
      setAnswers({});
      setNotes({});
    }).catch((err: any) => {
      console.error('Failed to answer:', err);
    });
  };

  const handleSkip = () => {
    AnswerQuestion({ answers: {}, notes: {} }).then(() => {
      setAskUserQuestions(null);
      setAnswers({});
      setNotes({});
    });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="w-full max-w-2xl max-h-[90vh] overflow-y-auto bg-slate-900 border border-slate-700 rounded-xl shadow-2xl p-6 flex flex-col gap-6 text-slate-300">
        <div className="flex items-center justify-between border-b border-slate-700 pb-4">
          <h2 className="text-xl font-bold text-white flex items-center gap-2">
            <span className="text-blue-400">🤖</span> AI cần bạn hỗ trợ quyết định
          </h2>
        </div>

        <div className="flex flex-col gap-8">
          {askUserQuestions.map((q, idx) => {
            const currentAnswer = answers[q.question] || '';
            const selectedLabels = currentAnswer ? currentAnswer.split(', ') : [];

            return (
              <div key={idx} className="flex flex-col gap-3">
                <h3 className="text-lg font-semibold text-slate-100">{q.header}</h3>
                <p className="text-sm text-slate-400">{q.question}</p>
                
                <div className="flex flex-col gap-2 mt-2">
                  {q.options && q.options.map((opt, oidx) => {
                    const isSelected = selectedLabels.includes(opt.label);
                    return (
                      <button
                        key={oidx}
                        onClick={() => handleSelect(q.question, opt.label, q.multiSelect)}
                        className={`text-left p-3 rounded-lg border transition-all duration-200 flex flex-col gap-1
                          ${isSelected ? 'bg-blue-900/30 border-blue-500' : 'bg-slate-800/50 border-slate-700 hover:border-slate-500 hover:bg-slate-800'}
                        `}
                      >
                        <span className={`font-medium ${isSelected ? 'text-blue-300' : 'text-slate-200'}`}>
                          {opt.label}
                        </span>
                        {opt.description && (
                          <span className="text-xs text-slate-400">{opt.description}</span>
                        )}
                      </button>
                    );
                  })}
                </div>

                <div className="mt-2 flex flex-col gap-1">
                  <label className="text-xs font-medium text-slate-500 uppercase tracking-wider">
                    Ghi chú thêm (Tuỳ chọn)
                  </label>
                  <textarea
                    className="w-full bg-slate-950 border border-slate-800 rounded-lg p-3 text-sm text-slate-300 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all resize-none min-h-[80px]"
                    placeholder="Nhập yêu cầu cụ thể của bạn..."
                    value={notes[q.question] || ''}
                    onChange={(e) => handleNoteChange(q.question, e.target.value)}
                  />
                </div>
              </div>
            );
          })}
        </div>

        <div className="flex items-center justify-end gap-3 pt-4 border-t border-slate-800 sticky bottom-0 bg-slate-900">
          <button
            onClick={handleSkip}
            className="px-4 py-2 rounded-lg text-sm font-medium text-slate-400 hover:bg-slate-800 hover:text-slate-200 transition-all"
          >
            Bỏ qua (Để AI tự quyết)
          </button>
          <button
            onClick={handleSubmit}
            className="px-5 py-2 rounded-lg text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20 transition-all flex items-center gap-2"
          >
            Gửi phản hồi <span>→</span>
          </button>
        </div>
      </div>
    </div>
  );
};
