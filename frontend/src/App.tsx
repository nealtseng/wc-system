import { useState } from 'react';
import { View } from './types';
import { Sidebar } from './components/Sidebar';
import { Overview } from './components/Overview';
import { UpcomingMatches } from './components/UpcomingMatches';
import { Pipeline } from './components/Pipeline';
import { MetaLearner } from './components/MetaLearner';
import { AdminDashboard } from './components/AdminDashboard';

export default function App() {
  const [activeView, setActiveView] = useState<View>('global-dashboard');

  const renderView = () => {
    switch (activeView) {
      case 'global-dashboard': return <Overview onViewChange={setActiveView} />;
      case 'upcoming-matches': return <UpcomingMatches />;
      case 'prediction-model': return <MetaLearner />;
      case 'data-operations': return <Pipeline />;
      case 'admin-panel': return <AdminDashboard />;
      default: return <Overview />;
    }
  };

  return (
    <div className="flex h-screen w-full bg-[#0a0a0a] overflow-hidden font-sans text-slate-300 selection:bg-cyan-500/30">
      <Sidebar activeView={activeView} onViewChange={setActiveView} />
      <main className="flex-1 overflow-y-auto bg-[#121212]">
        <div className="max-w-7xl mx-auto p-4 md:p-8">
          {renderView()}
        </div>
      </main>
    </div>
  );
}
