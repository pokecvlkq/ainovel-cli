import React, { useState } from 'react';
import { Send, Bot, User, MessageSquare } from 'lucide-react';

interface Message {
  id: number;
  sender: 'ai' | 'user';
  text: string;
}

export const ChatPage: React.FC = () => {
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState<Message[]>([
    { id: 1, sender: 'ai', text: 'Hello! I am the Coordinator Agent. How can I help you adjust the story?' }
  ]);

  const handleSend = () => {
    if (!input.trim()) return;
    
    const newMsg: Message = { id: Date.now(), sender: 'user', text: input };
    setMessages([...messages, newMsg]);
    setInput('');
    
    setTimeout(() => {
      setMessages(prev => [...prev, {
        id: Date.now(),
        sender: 'ai',
        text: 'I have noted your request. I will update the outline accordingly.'
      }]);
    }, 1000);
  };

  return (
    <div className="h-full flex flex-col bg-[var(--color-bg-primary)] max-w-4xl mx-auto w-full pt-4">
      <div className="pb-4 border-b border-[var(--color-border)] mb-6">
        <h2 className="text-xl font-bold text-white flex items-center gap-3 font-mono">
          <div className="bg-purple-500/10 p-1.5 rounded-lg border border-purple-500/20">
            <MessageSquare size={20} className="text-purple-400" />
          </div>
          CoCreate Chat
        </h2>
        <p className="text-sm text-[var(--color-text-secondary)] mt-1">Bàn luận với AI về cốt truyện, gợi ý sửa đổi, hoặc lên ý tưởng.</p>
      </div>
      
      <div className="flex-1 overflow-y-auto space-y-6 pr-2 mb-6 scroll-smooth">
        {messages.map((msg) => (
          <div key={msg.id} className={`flex gap-4 ${msg.sender === 'user' ? 'flex-row-reverse' : ''}`}>
            <div className={`p-2 rounded-xl h-10 w-10 flex items-center justify-center flex-shrink-0 border shadow-sm ${
              msg.sender === 'ai' 
                ? 'bg-purple-600/10 text-purple-400 border-purple-500/30' 
                : 'bg-blue-600/10 text-blue-400 border-blue-500/30'
            }`}>
              {msg.sender === 'ai' ? <Bot size={20} /> : <User size={20} />}
            </div>
            <div className={`p-4 rounded-2xl max-w-[80%] shadow-sm ${
              msg.sender === 'ai' 
                ? 'bg-[var(--color-bg-panel)] text-white border border-[var(--color-border)] rounded-tl-sm' 
                : 'bg-blue-600 text-white rounded-tr-sm border border-blue-500'
            }`}>
              <p className="text-[15px] whitespace-pre-wrap leading-relaxed">{msg.text}</p>
            </div>
          </div>
        ))}
      </div>
      
      <div className="bg-[var(--color-bg-panel)] p-2 rounded-xl flex items-center gap-2 border border-[var(--color-border)] shadow-lg focus-within:border-blue-500/50 transition-colors">
        <input 
          type="text" 
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          className="flex-1 bg-transparent border-none outline-none text-white px-4 py-2 placeholder-[var(--color-text-muted)] text-[15px]"
          placeholder="Type your message..."
        />
        <button 
          onClick={handleSend}
          disabled={!input.trim()}
          className="p-3 bg-blue-600 text-white rounded-lg hover:bg-blue-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-md shadow-blue-500/20"
        >
          <Send size={18} />
        </button>
      </div>
    </div>
  );
};
