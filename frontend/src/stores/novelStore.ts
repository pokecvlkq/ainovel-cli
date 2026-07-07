import { create } from 'zustand';

// Interface khớp 100% PascalCase với host.UISnapshot từ Wails backend.
// Dùng plain interface thay vì class import để tránh crash khi khởi tạo rỗng.
export interface UISnapshot {
  Provider: string;
  NovelName: string;
  ModelName: string;
  ModelContextWindow: number;
  ThinkingLevel: string;
  Style: string;
  RuntimeState: string;
  StatusLabel: string;
  Phase: string;
  Flow: string;
  CurrentChapter: number;
  TotalChapters: number;
  CompletedCount: number;
  TotalWordCount: number;
  InProgressChapter: number;
  PendingRewrites: number[];
  RewriteReason: string;
  PendingSteer: string;
  RecoveryLabel: string;
  IsRunning: boolean;
  Agents: AgentSnapshot[];
  ContextTokens: number;
  ContextWindow: number;
  ContextPercent: number;
  TotalInputTokens: number;
  TotalOutputTokens: number;
  TotalCostUSD: number;
  TotalSavedUSD: number;
  BudgetLimitUSD: number;
  Premise: string;
  Outline: OutlineSnapshot[];
  Characters: string[];
  LastCommitSummary: string;
  LastReviewSummary: string;
  RecentSummaries: string[];
  [key: string]: any; // cho phép các field khác từ backend
}

export interface AgentSnapshot {
  Name: string;
  State: string;
  TaskID: string;
  TaskKind: string;
  Summary: string;
  Tool: string;
  Turn: number;
}

export interface OutlineSnapshot {
  Chapter: number;
  Title: string;
  CoreEvent: string;
}

export interface NovelEvent {
  id: string;
  agent: string;
  type: 'info' | 'success' | 'error' | 'tool-call';
  message: string;
  timestamp: number;
}

// Giá trị mặc định an toàn — plain object, không class constructor
const emptySnapshot: UISnapshot = {
  Provider: '',
  NovelName: '',
  ModelName: '',
  ModelContextWindow: 0,
  ThinkingLevel: '',
  Style: '',
  RuntimeState: 'stopped',
  StatusLabel: '',
  Phase: '',
  Flow: '',
  CurrentChapter: 0,
  TotalChapters: 0,
  CompletedCount: 0,
  TotalWordCount: 0,
  InProgressChapter: 0,
  PendingRewrites: [],
  RewriteReason: '',
  PendingSteer: '',
  RecoveryLabel: '',
  IsRunning: false,
  Agents: [],
  ContextTokens: 0,
  ContextWindow: 0,
  ContextPercent: 0,
  TotalInputTokens: 0,
  TotalOutputTokens: 0,
  TotalCostUSD: 0,
  TotalSavedUSD: 0,
  BudgetLimitUSD: 0,
  Premise: '',
  Outline: [],
  Characters: [],
  LastCommitSummary: '',
  LastReviewSummary: '',
  RecentSummaries: [],
};

interface NovelState {
  snapshot: UISnapshot;
  events: NovelEvent[];
  streamBuffer: string;
  isComplete: boolean;
  setSnapshot: (snapshot: UISnapshot) => void;
  addEvent: (event: NovelEvent) => void;
  appendStream: (text: string) => void;
  clearStream: () => void;
  setComplete: (done: boolean) => void;
}

export const useNovelStore = create<NovelState>((set) => ({
  snapshot: { ...emptySnapshot },
  events: [],
  streamBuffer: '',
  isComplete: false,
  setSnapshot: (snapshot) => set({ snapshot }),
  addEvent: (event) => set((state) => ({ events: [...state.events, event] })),
  appendStream: (text) => set((state) => ({ streamBuffer: state.streamBuffer + text })),
  clearStream: () => set({ streamBuffer: '' }),
  setComplete: (done) => set({ isComplete: done }),
}));
