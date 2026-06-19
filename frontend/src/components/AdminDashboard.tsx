import React, { useEffect, useState } from 'react';
import { Activity, Database, Server, Zap } from 'lucide-react';
import { cn } from '../lib/utils';
import { fetchAdminStatus } from '../services/api';
import type { AdminStatus } from '../types';

export function AdminDashboard() {
  const [status, setStatus] = useState<AdminStatus | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchAdminStatus()
      .then(setStatus)
      .catch((err: Error) => setError(err.message));
  }, []);

  const signalsCount = status?.signals_count ?? 0;
  const matchesCount = status?.matches_count ?? 0;
  const oddsPct = status && status.matches_count > 0
    ? Math.min(100, Math.round((status.odds_count / status.matches_count) * 100))
    : 0;

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 ease-out flex flex-col">
      <header className="border-b border-[#2a2a2a] pb-4">
        <h2 className="text-xl font-bold tracking-tight text-slate-100 uppercase font-mono">系統管理面板</h2>
        <p className="text-[10px] text-[#666] mt-1 font-mono tracking-widest uppercase">基礎設施監控與資料管線狀態</p>
      </header>

      {error && (
        <div className="border border-orange-500/30 bg-orange-500/10 p-3 text-[10px] font-mono text-orange-400">
          {error}
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <MetricCard
          title="已同步球隊"
          value={status ? String(status.teams_count) : '—'}
          subtitle="catalog + pipeline"
          icon={Database}
          trend="neutral"
        />
        <MetricCard
          title="賽程場次"
          value={status ? String(matchesCount) : '—'}
          subtitle="WorldCup2026 DB"
          icon={Activity}
          trend="neutral"
        />
        <MetricCard
          title="活躍 EV 信號"
          value={status ? String(signalsCount) : '—'}
          subtitle="含盤口之近期場次"
          icon={Zap}
          trend={signalsCount > 0 ? 'up' : 'neutral'}
        />
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 mt-2">
        <div className="md:col-span-2 bg-[#0f0f0f] border border-[#2a2a2a] p-6 shadow-none">
          <div className="flex items-center gap-3 mb-6 border-b border-[#2a2a2a] pb-4">
            <Server className="w-4 h-4 text-cyan-500" />
            <h3 className="text-sm font-bold text-slate-100 uppercase font-mono tracking-widest">資料管線覆蓋率</h3>
          </div>

          <div className="space-y-8">
            <div className="space-y-3">
              <div className="flex items-center justify-between font-mono">
                <div className="flex items-center gap-2">
                  <Database className="w-3.5 h-3.5 text-cyan-400" />
                  <span className="text-[10px] font-bold text-slate-300 uppercase tracking-widest">Odds 寫入率</span>
                </div>
                <span className="text-[10px] text-[#666] tracking-widest">
                  {status ? `${status.odds_count} / ${matchesCount}` : '—'}
                </span>
              </div>
              <div className="w-full bg-[#161616] border border-[#2a2a2a] h-1 overflow-hidden">
                <div className="bg-cyan-500 h-1 transition-all" style={{ width: `${oddsPct}%` }} />
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between font-mono">
                <div className="flex items-center gap-2">
                  <Activity className="w-3.5 h-3.5 text-slate-400" />
                  <span className="text-[10px] font-bold text-slate-300 uppercase tracking-widest">Scraper 健康</span>
                </div>
                <span className="text-[10px] text-[#666] tracking-widest">
                  {status ? `${status.scrapers_ok} ok · ${status.scrapers_degraded} degraded` : '—'}
                </span>
              </div>
              <div className="w-full bg-[#161616] border border-[#2a2a2a] h-1 overflow-hidden">
                <div
                  className="bg-slate-400 h-1 transition-all"
                  style={{
                    width: status && status.scrapers_total > 0
                      ? `${Math.round((status.scrapers_ok / status.scrapers_total) * 100)}%`
                      : '0%',
                  }}
                />
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between font-mono">
                <div className="flex items-center gap-2">
                  <Zap className="w-3.5 h-3.5 text-orange-400" />
                  <span className="text-[10px] font-bold text-slate-300 uppercase tracking-widest">Signals 輸出</span>
                </div>
                <span className="text-[10px] text-[#666] tracking-widest">{signalsCount} 場</span>
              </div>
              <div className="w-full bg-[#161616] border border-[#2a2a2a] h-1 overflow-hidden">
                <div
                  className="bg-orange-500 h-1 transition-all"
                  style={{ width: `${Math.min(100, signalsCount * 5)}%` }}
                />
              </div>
            </div>
          </div>
        </div>

        <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-6 flex flex-col justify-between shadow-none">
          <div>
            <h3 className="text-sm font-bold text-slate-100 mb-2 uppercase font-mono tracking-widest">系統健康狀態</h3>
            <p className="text-[10px] text-[#666] font-mono tracking-wide leading-relaxed">
              {status?.last_sync
                ? `最後同步：${new Date(status.last_sync).toLocaleString()}`
                : '尚未完成首次同步'}
            </p>
          </div>
          <div className="bg-[#121212] border border-[#2a2a2a] p-4 mt-6">
            <div className="flex items-center gap-3">
              <div className="relative flex h-2 w-2">
                {status?.system_healthy ? (
                  <>
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75" />
                    <span className="relative inline-flex h-2 w-2 bg-cyan-500" />
                  </>
                ) : (
                  <span className="relative inline-flex h-2 w-2 bg-orange-500" />
                )}
              </div>
              <span
                className={cn(
                  'text-[10px] font-bold uppercase tracking-widest font-mono',
                  status?.system_healthy ? 'text-cyan-400' : 'text-orange-400',
                )}
              >
                狀態: {status?.system_healthy ? '綠燈運行中' : '待同步 / 部分降級'}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function MetricCard({
  title,
  value,
  subtitle,
  icon: Icon,
  trend,
}: {
  title: string;
  value: string;
  subtitle: string;
  icon: React.ComponentType<{ className?: string }>;
  trend: 'up' | 'down' | 'neutral';
}) {
  return (
    <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-5">
      <div className="flex items-center justify-between mb-4">
        <span className="text-[10px] font-bold text-[#666] uppercase tracking-widest font-mono">{title}</span>
        <Icon
          className={cn(
            'w-4 h-4',
            trend === 'up' ? 'text-cyan-400' : trend === 'down' ? 'text-orange-400' : 'text-[#555]',
          )}
        />
      </div>
      <div className="text-2xl font-bold text-slate-100 font-mono tracking-tight">{value}</div>
      <div className="text-[9px] text-[#555] font-mono mt-2 uppercase tracking-widest">{subtitle}</div>
    </div>
  );
}
