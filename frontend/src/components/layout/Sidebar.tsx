import React from 'react';
import { NavLink } from 'react-router-dom';
import { BookOpen, LayoutDashboard, MessageSquare, Settings, Edit3, Eye, ChevronLeft, ChevronRight } from 'lucide-react';
import { useUIStore } from '../../stores/uiStore';

export const Sidebar: React.FC = () => {
  const { sidebarCollapsed, toggleSidebar } = useUIStore();

  const navItems = [
    { name: 'Dashboard', path: '/', icon: <LayoutDashboard size={20} /> },
    { name: 'Editor', path: '/editor/1', icon: <Edit3 size={20} /> },
    { name: 'Review', path: '/review/1', icon: <Eye size={20} /> },
    { name: 'Chat', path: '/chat', icon: <MessageSquare size={20} /> },
    { name: 'Settings', path: '/settings', icon: <Settings size={20} /> },
  ];

  return (
    <div className={`flex flex-col bg-[var(--color-bg-secondary)] border-r border-[var(--color-border)] transition-all duration-300 ${sidebarCollapsed ? 'w-16' : 'w-64'}`}>
      <div className="flex items-center justify-between p-4 border-b border-[var(--color-border)] h-14">
        {!sidebarCollapsed && (
          <div className="flex items-center gap-2 text-blue-500 font-bold text-lg font-mono truncate">
            <BookOpen size={20} />
            <span className="text-white">AINovel</span>
          </div>
        )}
        <button onClick={toggleSidebar} className="p-1 hover:bg-[var(--color-bg-panel)] rounded text-[var(--color-text-secondary)] hover:text-white transition-colors cursor-pointer ml-auto">
          {sidebarCollapsed ? <ChevronRight size={20} /> : <ChevronLeft size={20} />}
        </button>
      </div>
      <nav className="flex-1 py-4 flex flex-col gap-2 px-2">
        {navItems.map((item) => (
          <NavLink
            key={item.name}
            to={item.path}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2.5 rounded-md transition-colors font-medium text-sm ${
                isActive
                  ? 'bg-blue-600/10 text-blue-500 border border-blue-500/20'
                  : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-hover)] hover:text-white border border-transparent'
              } ${sidebarCollapsed ? 'justify-center' : ''}`
            }
            title={sidebarCollapsed ? item.name : undefined}
          >
            {item.icon}
            {!sidebarCollapsed && <span>{item.name}</span>}
          </NavLink>
        ))}
      </nav>
    </div>
  );
};
