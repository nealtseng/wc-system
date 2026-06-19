import React from 'react';
import { View } from '../types';
import { cn } from '../lib/utils';
import { 
  LayoutDashboard, 
  Trophy,
  BrainCircuit,
  Database,
  ShieldAlert,
  CalendarDays
} from 'lucide-react';

interface SidebarProps {
  activeView: View;
  onViewChange: (view: View) => void;
}

export function Sidebar({ activeView, onViewChange }: SidebarProps) {
  const items = [
    { id: 'global-dashboard' as View, label: '全域監控面板', icon: LayoutDashboard },
    { id: 'upcoming-matches' as View, label: '小組賽況', icon: CalendarDays },
    { id: 'prediction-model' as View, label: '量化預測實驗室', icon: BrainCircuit },
    { id: 'data-operations' as View, label: '數據管線營運', icon: Database },
    { id: 'admin-panel' as View, label: '系統管理面板', icon: ShieldAlert },
  ];

  return (
    <div className="w-64 bg-[#080808] border-r border-[#2a2a2a] flex flex-col h-full text-slate-400 font-mono">
      <div className="p-6 border-b border-[#2a2a2a]">
        <h1 className="text-xl font-bold text-slate-100 uppercase tracking-[0.1em]">
          Alpha_Quant
          <span className="block text-[10px] text-cyan-500 mt-1 uppercase tracking-[0.2em]">
            量化預測系統
          </span>
        </h1>
      </div>

      <nav className="flex-1 space-y-0.5 mt-6 px-3">
        {items.map((item) => {
          const Icon = item.icon;
          const isActive = activeView === item.id;
          return (
            <button
              key={item.id}
              onClick={() => onViewChange(item.id)}
              className={cn(
                "w-full flex items-center gap-3 px-4 py-3 text-xs uppercase tracking-widest font-medium transition-colors border border-transparent",
                isActive 
                  ? "bg-[#161616] text-cyan-400 border-l-[3px] border-l-cyan-500" 
                  : "hover:bg-[#121212] hover:text-slate-200"
              )}
            >
              <Icon className={cn("w-4 h-4", isActive ? "text-cyan-400" : "text-[#555]")} />
              {item.label}
            </button>
          );
        })}
      </nav>

      <div className="p-4 border-t border-[#2a2a2a]">
        <div className="flex items-center justify-center gap-3 px-4 py-2 border border-[#2a2a2a] bg-[#121212]">
          <div className="w-1.5 h-1.5 bg-cyan-500 animate-pulse border border-cyan-300" />
          <span className="text-[10px] font-mono text-cyan-500 uppercase tracking-widest">系統連線中</span>
        </div>
      </div>
    </div>
  );
}
