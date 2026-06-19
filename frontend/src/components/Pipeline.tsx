import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Activity, Clock, DatabaseZap, RefreshCw, Globe, Database } from 'lucide-react';
import { cn } from '../lib/utils';
import { NationFlag } from './NationFlag';
import { SquadView } from './SquadView';
import { teamFromResponse } from '../data';
import { fetchPipelineStatus, fetchPipelineTargets, fetchTeams, syncFBref, syncOdds, syncPipeline } from '../services/api';
import type { PipelineScraperStatus, PipelineSchedule, TeamResponse } from '../types';

const SCRAPER_ICONS: Record<string, typeof Globe> = {
  'World Bank (GDP)': Globe,
  'Wikimedia (Squad Meta)': Database,
  'Kaggle Hist. Results': DatabaseZap,
  'Kaggle/FIFA Players': DatabaseZap,
  'WorldCup2026 (Fixtures)': DatabaseZap,
  'The Odds API': DatabaseZap,
};

function formatLastFetch(iso: string): string {
  if (!iso) return '尚未同步';
  const d = new Date(iso);
  const diffMin = Math.floor((Date.now() - d.getTime()) / 60000);
  if (diffMin < 1) return '剛剛';
  if (diffMin < 60) return `${diffMin} 分鐘前`;
  return d.toLocaleString();
}

function formatScheduleCountdown(nextAt: string): string {
  const diff = new Date(nextAt).getTime() - Date.now();
  if (diff <= 0) return '即將執行';
  const h = Math.floor(diff / 3_600_000);
  const m = Math.floor((diff % 3_600_000) / 60_000);
  const s = Math.floor((diff % 60_000) / 1_000);
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
}

function mapScraperUI(status: PipelineScraperStatus['status']): 'healthy' | 'recovering' | 'offline' {
  if (status === 'ok') return 'healthy';
  if (status === 'degraded') return 'recovering';
  return 'offline';
}

function scheduleIsImminent(nextAt: string): boolean {
  const diff = new Date(nextAt).getTime() - Date.now();
  return diff > 0 && diff <= 15 * 60_000;
}

function scheduleKey(sch: PipelineSchedule): string {
  return `${sch.wc_match_id ?? ''}-${sch.window_key ?? ''}-${sch.next_at}`;
}

export function Pipeline() {
  const [selectedTeamId, setSelectedTeamId] = useState<string | null>(null);
  const [teams, setTeams] = useState<TeamResponse[]>([]);
  const [scrapers, setScrapers] = useState<PipelineScraperStatus[]>([]);
  const [schedules, setSchedules] = useState<PipelineSchedule[]>([]);
  const [scraperTargets, setScraperTargets] = useState<Record<string, string>>({});
  const [syncing, setSyncing] = useState(false);
  const [syncingOdds, setSyncingOdds] = useState(false);
  const [syncingFBref, setSyncingFBref] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    try {
      const [teamData, pipeline, targets] = await Promise.all([
        fetchTeams(),
        fetchPipelineStatus(),
        fetchPipelineTargets(),
      ]);
      setTeams(teamData);
      setScrapers(pipeline.scrapers);
      setSchedules(pipeline.schedules);
      setScraperTargets(targets);
      setLoadError(null);
    } catch (err) {
      setLoadError((err as Error).message);
    }
  }, []);

  useEffect(() => {
    loadData();
    const id = setInterval(loadData, 30_000);
    return () => clearInterval(id);
  }, [loadData]);

  const handleSync = async () => {
    setSyncing(true);
    try {
      await syncPipeline();
      await loadData();
    } catch (err) {
      setLoadError((err as Error).message);
    } finally {
      setSyncing(false);
    }
  };

  const handleFBrefSync = async () => {
    setSyncingFBref(true);
    try {
      const res = await syncFBref();
      if (res.status !== 'ok') {
        setLoadError(res.message ?? res.error ?? 'FBref xG 同步部分失敗（已保留舊值）');
      } else {
        setLoadError(null);
      }
      await loadData();
    } catch (err) {
      setLoadError((err as Error).message);
    } finally {
      setSyncingFBref(false);
    }
  };

  const handleOddsSync = async () => {
    setSyncingOdds(true);
    try {
      const res = await syncOdds();
      if (res.status === 'error') {
        setLoadError(res.message ?? '賠率同步失敗');
      } else {
        setLoadError(null);
      }
      await loadData();
    } catch (err) {
      setLoadError((err as Error).message);
    } finally {
      setSyncingOdds(false);
    }
  };

  const teamInfos = useMemo(() => teams.map(teamFromResponse), [teams]);

  if (selectedTeamId) {
    const team = teamInfos.find((t) => t.id === selectedTeamId);
    const live = teams.find((t) => t.id === selectedTeamId);
    if (team) {
      return <SquadView team={team} live={live} onBack={() => setSelectedTeamId(null)} />;
    }
  }

  const matrixRows = teams.length > 0 ? teams : null;

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 ease-out pb-8">
      <header className="border-b border-[#2a2a2a] pb-4">
        <h2 className="text-xl font-bold tracking-tight text-slate-100 uppercase font-mono">數據管線營運</h2>
        <p className="text-[10px] text-[#666] mt-1 font-mono tracking-widest uppercase">
          World Bank · Wikimedia · Kaggle · WorldCup2026 · The Odds API
        </p>
      </header>

      {loadError && (
        <div className="border border-orange-500/30 bg-orange-500/10 p-3 text-[10px] font-mono text-orange-400">
          {loadError}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-6 flex flex-col">
          <div className="flex items-center gap-3 mb-6 border-b border-[#2a2a2a] pb-4">
            <Clock className="w-4 h-4 text-cyan-400" />
            <h3 className="text-sm font-bold text-slate-200 font-mono uppercase tracking-widest">Odds 賽前排程</h3>
          </div>
          <p className="text-[9px] text-[#555] font-mono mb-4 leading-relaxed">
            依每場開球時間自動在賽前 12h / 2h / 15m 拉取 The Odds API（後端每分鐘檢查）。
          </p>
          <div className="flex gap-2 mb-4">
            <button
              type="button"
              onClick={handleOddsSync}
              disabled={syncingOdds || syncing}
              className="flex items-center gap-2 px-3 py-1.5 border border-cyan-500/40 hover:bg-cyan-500/10 text-cyan-400 transition-colors font-mono text-[9px] uppercase tracking-widest cursor-pointer disabled:opacity-50"
            >
              <RefreshCw className={cn('w-3 h-3', syncingOdds && 'animate-spin')} />
              {syncingOdds ? '更新賠率中…' : '立即更新賠率'}
            </button>
          </div>
          <div className="space-y-4">
            {schedules.length === 0 ? (
              <p className="text-[10px] text-[#666] font-mono">目前無 upcoming 賽事的 odds 排程（或尚未同步賽程）</p>
            ) : (
              schedules.map((sch, idx) => (
                <React.Fragment key={scheduleKey(sch)}>
                  {idx > 0 && <div className="w-[1px] h-3 bg-[#2a2a2a] ml-5" />}
                  <PollingNode
                    phase={sch.label}
                    matchLabel={sch.match_label}
                    target={sch.endpoint}
                    nextRun={formatScheduleCountdown(sch.next_at)}
                    status={scheduleIsImminent(sch.next_at) ? 'active' : 'waiting'}
                  />
                </React.Fragment>
              ))
            )}
          </div>
        </div>

        <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-6 flex flex-col">
          <div className="flex items-center justify-between border-b border-[#2a2a2a] pb-4 mb-6">
            <div className="flex items-center gap-3">
              <DatabaseZap className="w-4 h-4 text-cyan-400" />
              <h3 className="text-sm font-bold text-slate-200 font-mono uppercase tracking-widest">巨集資料源同步</h3>
            </div>
            <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleSync}
              disabled={syncing || syncingFBref}
              className="flex items-center gap-2 px-3 py-1.5 border border-[#333] hover:border-cyan-500/50 hover:bg-cyan-500/10 text-[#888] hover:text-cyan-400 transition-colors font-mono text-[9px] uppercase tracking-widest cursor-pointer disabled:opacity-50"
            >
              <RefreshCw className={cn('w-3 h-3', syncing && 'animate-spin')} />
              {syncing ? '全量同步中…' : '全量同步'}
            </button>
            <button
              type="button"
              onClick={handleFBrefSync}
              disabled={syncingFBref || syncing}
              className="flex items-center gap-2 px-3 py-1.5 border border-orange-500/40 hover:bg-orange-500/10 text-orange-400 transition-colors font-mono text-[9px] uppercase tracking-widest cursor-pointer disabled:opacity-50"
            >
              <RefreshCw className={cn('w-3 h-3', syncingFBref && 'animate-spin')} />
              {syncingFBref ? 'FBref xG…' : '同步 FBref xG'}
            </button>
            </div>
          </div>

          <div className="space-y-4">
            {scrapers.length === 0 ? (
              <p className="text-[10px] text-[#666] font-mono">後端啟動中，首次同步約需 30–60 秒…</p>
            ) : (
              scrapers.map((s) => (
                <ScraperStatus
                  key={s.name}
                  name={s.name}
                  target={scraperTargets[s.name] ?? s.name}
                  status={mapScraperUI(s.status)}
                  lastRun={formatLastFetch(s.last_fetch)}
                  logs={s.message ? [`[INFO] ${s.message}`] : []}
                  icon={SCRAPER_ICONS[s.name] ?? DatabaseZap}
                />
              ))
            )}
          </div>
        </div>
      </div>

      <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-6 mt-6">
        <div className="flex items-center justify-between mb-6 border-b border-[#2a2a2a] pb-4">
          <div className="flex items-center gap-3">
            <Activity className="w-4 h-4 text-cyan-400" />
            <h3 className="text-sm font-bold text-slate-200 font-mono uppercase tracking-widest">
              Unified Momentum Matrix (多維動量矩陣)
            </h3>
          </div>
          <div className="text-[9px] text-[#666] font-mono tracking-widest uppercase flex items-center gap-2">
            <span className="w-1.5 h-1.5 bg-cyan-500 rounded-none animate-pulse" />
            資料來源：Kaggle ELO · World Bank GDP · Wiki 敘事
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left font-mono">
            <thead>
              <tr className="text-[10px] text-[#666] uppercase tracking-widest border-b border-[#2a2a2a]">
                <th className="pb-3 px-4 font-normal">國家 (Target)</th>
                <th className="pb-3 px-4 font-normal text-right">Kaggle ELO</th>
                <th className="pb-3 px-4 font-normal text-right">勝率 (Win %)</th>
                <th className="pb-3 px-4 font-normal text-right">人均 GDP (USD)</th>
                <th className="pb-3 px-4 font-normal text-right">敘事權重</th>
                <th className="pb-3 px-4 font-normal text-right">動量指標</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1a1a1a]">
              {!matrixRows ? (
                <tr>
                  <td colSpan={6} className="py-6 px-4 text-center text-[10px] text-[#555]">
                    等待後端同步完成…
                  </td>
                </tr>
              ) : (
                matrixRows.map((team) => (
                  <tr
                    key={team.id}
                    onClick={() => setSelectedTeamId(team.id)}
                    className="hover:bg-[#121212] transition-colors cursor-pointer group"
                  >
                    <td className="py-3 px-4 relative">
                      <div className="absolute left-0 top-0 bottom-0 w-0 group-hover:w-1 bg-cyan-500 transition-all" />
                      <div className="flex items-center gap-3">
                        <NationFlag iso2={team.iso2 ?? 'un'} alt={team.name} className="w-8 h-5" />
                        <span className="text-xs font-bold text-slate-200">{team.id}</span>
                        <span className="text-[10px] text-[#555] uppercase hidden sm:inline-block">{team.name}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-right">
                      <span className="text-xs text-slate-300">{Math.round(team.elo)}</span>
                    </td>
                    <td className="py-3 px-4 text-right">
                      <span className="text-xs text-[#888]">{(team.win_rate * 100).toFixed(1)}%</span>
                    </td>
                    <td className="py-3 px-4 text-right">
                      <span className="text-xs text-slate-300">
                        {team.gdp_per_capita > 0
                          ? team.gdp_per_capita.toLocaleString(undefined, { maximumFractionDigits: 0 })
                          : '—'}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-right">
                      <span className="text-xs text-cyan-400 opacity-80">{team.narrative_weight.toFixed(2)}</span>
                    </td>
                    <td className="py-3 px-4 text-right">
                      <span
                        className={cn(
                          'text-xs px-2 py-1 border',
                          team.momentum >= 0
                            ? 'text-cyan-400 border-cyan-500/30 bg-cyan-500/10'
                            : 'text-orange-400 border-orange-500/30 bg-orange-500/10',
                        )}
                      >
                        {team.momentum >= 0 ? '+' : ''}
                        {team.momentum.toFixed(2)}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function PollingNode({
  phase,
  matchLabel,
  target,
  nextRun,
  status,
}: {
  phase: string;
  matchLabel?: string;
  target: string;
  nextRun: string;
  status: string;
}) {
  return (
    <div
      className={cn(
        'p-3 border',
        status === 'active' ? 'bg-[#161616] border-cyan-500/30' : 'bg-[#121212] border-[#2a2a2a]',
      )}
    >
      <div className="flex justify-between items-start">
        <div className="flex gap-3">
          <div className="mt-1">
            {status === 'active' ? (
              <div className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75" />
                <span className="relative inline-flex h-2 w-2 bg-cyan-500" />
              </div>
            ) : (
              <div className="h-1.5 w-1.5 bg-[#444] mt-0.5" />
            )}
          </div>
          <div>
            <div
              className={cn(
                'font-bold text-xs font-mono uppercase tracking-widest',
                status === 'active' ? 'text-cyan-400' : 'text-slate-400',
              )}
            >
              {phase}
            </div>
            {matchLabel && (
              <div className="text-[9px] text-slate-300 mt-0.5 font-mono tracking-wider">{matchLabel}</div>
            )}
            <div className="text-[9px] text-[#666] mt-1 font-mono tracking-wider">{target}</div>
          </div>
        </div>
        <div className="text-right">
          <div className="text-[8px] text-[#555] mb-1 font-mono uppercase tracking-widest">下一次執行</div>
          <div className={cn('font-mono text-[10px]', status === 'active' ? 'text-slate-200' : 'text-[#555]')}>
            {nextRun}
          </div>
        </div>
      </div>
    </div>
  );
}

function ScraperStatus({
  name,
  target,
  status,
  lastRun,
  logs,
  icon: CustomIcon,
}: {
  name: string;
  target: string;
  status: 'healthy' | 'recovering' | 'offline';
  lastRun: string;
  logs: string[];
  icon: React.ComponentType<{ className?: string }>;
}) {
  const isHealthy = status === 'healthy';
  const displayStatus = status === 'healthy' ? '健康' : status === 'recovering' ? '部分成功' : '離線';
  const Icon = CustomIcon || DatabaseZap;

  return (
    <div className="border border-[#2a2a2a] bg-[#121212]">
      <div className="p-3 flex justify-between items-center">
        <div className="flex items-center gap-3">
          <Icon className={cn('w-3.5 h-3.5', isHealthy ? 'text-cyan-500' : 'text-orange-500')} />
          <div>
            <div className="text-[10px] font-bold text-slate-200 font-mono uppercase tracking-widest">{name}</div>
            <div className="text-[8px] text-[#555] font-mono mt-1 tracking-wider">{target}</div>
          </div>
        </div>
        <div className="text-right">
          <div className="text-[8px] uppercase font-bold tracking-widest mb-1 font-mono">
            <span className={isHealthy ? 'text-cyan-500' : 'text-orange-500'}>{displayStatus}</span>
          </div>
          <div className="text-[8px] text-[#666] font-mono tracking-widest">{lastRun}</div>
        </div>
      </div>

      {!isHealthy && logs.length > 0 && (
        <div className="bg-[#0a0a0a] p-3 border-t border-[#2a2a2a] text-[9px] font-mono leading-relaxed">
          {logs.map((log, idx) => (
            <div key={idx} className={cn('mb-1', log.includes('ERROR') ? 'text-orange-400' : 'text-[#666]')}>
              {log}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
