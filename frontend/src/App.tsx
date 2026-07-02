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
import './styles/globals.css';

const AuthGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const navigate = useNavigate();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    GetSnapshot()
      .then((snap) => {
        if (!snap || !snap.NovelName) {
          navigate('/welcome');
        }
      })
      .catch(() => {
        navigate('/welcome');
      })
      .finally(() => {
        setChecking(false);
      });
  }, [navigate]);

  if (checking) {
    return <div className="h-screen w-screen bg-black flex items-center justify-center"><div className="animate-pulse text-slate-500 font-mono">Initializing Backend...</div></div>;
  }

  return <>{children}</>;
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
