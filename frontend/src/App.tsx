import React, { useEffect, useState } from 'react';
import { BrowserRouter, Routes, Route, useNavigate } from 'react-router-dom';
import { AppShell } from './components/layout/AppShell';
import { DashboardPage } from './pages/DashboardPage';
import { WelcomePage } from './pages/WelcomePage';
import { EditorPage } from './pages/EditorPage';
import { ReviewPage } from './pages/ReviewPage';
import { ChatPage } from './pages/ChatPage';
import { SettingsPage } from './pages/SettingsPage';
import { GetSnapshot } from '../wailsjs/go/main/App';
import { useNovelStore } from './stores/novelStore';
import { AskUserModal } from './components/dashboard/AskUserModal';
import './styles/globals.css';

import { EventsOn } from '../wailsjs/runtime/runtime';

const AuthGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const navigate = useNavigate();
  const [checking, setChecking] = useState(true);
  const { setSnapshot, addEvent, appendStream, clearStream, setComplete, setAskUserQuestions } = useNovelStore();

  useEffect(() => {
    GetSnapshot()
      .then((snap) => {
        if (!snap || !snap.NovelName) {
          navigate('/welcome');
        } else {
          setSnapshot(snap);
        }
      })
      .catch(() => {
        navigate('/welcome');
      })
      .finally(() => {
        setChecking(false);
      });
  }, [navigate, setSnapshot]);

  // Lắng nghe sự kiện từ Backend
  useEffect(() => {
    const unsubSnapshot = EventsOn('novel:snapshot', (snap: any) => {
      if (snap) setSnapshot(snap);
    });
    const unsubEvent = EventsOn('novel:event', (ev: any) => {
      if (ev) {
        addEvent({
          id: ev.ID || '',
          agent: ev.Agent || 'SYSTEM',
          type: ev.Level || 'info',
          message: ev.Summary || ev.Detail || '',
          timestamp: new Date(ev.Time).getTime() || Date.now(),
        });
      }
    });
    const unsubStream = EventsOn('novel:stream', (delta: any) => {
      if (delta) appendStream(delta);
    });
    const unsubStreamClear = EventsOn('novel:stream-clear', () => {
      clearStream();
    });
    const unsubDone = EventsOn('novel:done', () => {
      setComplete(true);
    });
    const unsubAskUser = EventsOn('novel:ask_user', (questions: any) => {
      if (questions && questions.length > 0) {
        setAskUserQuestions(questions);
      }
    });

    return () => {
      unsubSnapshot();
      unsubEvent();
      unsubStream();
      unsubStreamClear();
      unsubDone();
      unsubAskUser();
    };
  }, [setSnapshot, addEvent, appendStream, clearStream, setComplete, setAskUserQuestions]);

  if (checking) {
    return <div className="h-screen w-screen bg-black flex items-center justify-center"><div className="animate-pulse text-slate-500 font-mono">Đang kết nối Backend...</div></div>;
  }

  return (
    <>
      {children}
      <AskUserModal />
    </>
  );
};

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/welcome" element={<WelcomePage />} />
        <Route path="/" element={<AuthGuard><AppShell /></AuthGuard>}>
          <Route index element={<DashboardPage />} />
          <Route path="editor/:chapter" element={<EditorPage />} />
          <Route path="review/:chapter" element={<ReviewPage />} />
          <Route path="chat" element={<ChatPage />} />
          <Route path="settings" element={<SettingsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
