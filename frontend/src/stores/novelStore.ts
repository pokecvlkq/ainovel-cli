import { create } from 'zustand';

export interface AgentStatus {
  name: string;
  status: 'idle' | 'running' | 'done' | 'error';
  message?: string;
}

export interface NovelEvent {
  id: string;
  agent: string;
  type: 'info' | 'success' | 'error' | 'tool-call';
  message: string;
  timestamp: number;
}

export interface NovelSnapshot {
  currentChapter: number;
  totalChapters: number;
  agents: AgentStatus[];
  isWriting: boolean;
  totalTokens: number;
  NovelName?: string;
}

interface NovelState {
  snapshot: NovelSnapshot;
  events: NovelEvent[];
  streamBuffer: string;
  setSnapshot: (snapshot: NovelSnapshot) => void;
  addEvent: (event: NovelEvent) => void;
  appendStream: (text: string) => void;
  clearStream: () => void;
}

export const useNovelStore = create<NovelState>((set) => ({
  snapshot: {
    currentChapter: 0,
    totalChapters: 10,
    agents: [
      { name: 'Coordinator', status: 'idle' },
      { name: 'Architect', status: 'idle' },
      { name: 'Writer', status: 'idle' },
      { name: 'Editor', status: 'idle' },
    ],
    isWriting: false,
    totalTokens: 0,
  },
  events: [],
  streamBuffer: '',
  setSnapshot: (snapshot) => set({ snapshot }),
  addEvent: (event) => set((state) => ({ events: [...state.events, event] })),
  appendStream: (text) => set((state) => ({ streamBuffer: state.streamBuffer + text })),
  clearStream: () => set({ streamBuffer: '' }),
}));
