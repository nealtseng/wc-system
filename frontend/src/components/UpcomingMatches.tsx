import React, { useMemo, useState } from 'react';
import {
  CalendarDays,
  ChevronDown,
  ChevronUp,
  Hash,
  Loader2,
  MapPin,
  Radio,
  RefreshCw,
  TrendingUp,
} from 'lucide-react';
import { cn } from '../lib/utils';
import { formatTaipeiDateTime, formatTaipeiDateTimeLong, formatVenueLocalTime } from '../lib/datetime';
import {
  filterMatches,
  matchCounts,
  resolveMatchStatus,
  signalsByMatchId,
  sortMatchesForDisplay,
  type MatchFilter,
} from '../lib/matchSignals';
import { useMatches } from '../hooks/useMatches';
import { useSignals } from '../hooks/useSignals';
import type { MatchResponse, MatchSignal } from '../types';
import { NationFlag } from './NationFlag';

const STAGE_LABELS: Record<string, string> = {
  group: '小組賽',
  r32: '三十二強',
  r16: '十六強',
  qf: '八強',
  sf: '四強',
  third: '季軍戰',
  final: '決賽',
};

const FILTER_LABELS: Record<MatchFilter, string> = {
  active: '進行中 & 未開賽',
  live: '進行中',
  upcoming: '未開賽',
  finished: '已結束',
  all: '全部',
};

export function UpcomingMatches() {
  const { data: matchesData, loading, error, reload, lastUpdated } = useMatches({ pollWhenLive: true });
  const { data: signals, reload: reloadSignals } = useSignals();
  const [filter, setFilter] = useState<MatchFilter>('active');
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [refreshing, setRefreshing] = useState(false);

  const signalMap = useMemo(() => signalsByMatchId(signals), [signals]);

  const displayMatches = useMemo(() => {
    const sorted = sortMatchesForDisplay(matchesData ?? []);
    return filterMatches(sorted, filter);
  }, [matchesData, filter]);

  const counts = useMemo(() => matchCounts(matchesData ?? []), [matchesData]);

  const handleRefresh = async () => {
    setRefreshing(true);
    await Promise.all([reload(), reloadSignals()]);
    setRefreshing(false);
  };

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 ease-out">
      <header className="border-b border-[#2a2a2a] pb-4 flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold tracking-tight text-slate-100 uppercase font-mono">小組賽況</h2>
          <p className="text-[10px] text-[#666] mt-1 font-mono tracking-widest uppercase">
            完整賽程 · 台灣時間 (UTC+8) · 即時比分 · The Odds API 盤口
          </p>
        </div>
        <button
          type="button"
          onClick={handleRefresh}
          disabled={refreshing}
          className="flex items-center gap-2 border border-[#2a2a2a] px-3 py-1.5 text-[10px] font-mono uppercase tracking-widest text-[#888] hover:text-cyan-400 hover:border-cyan-500/40 transition-colors disabled:opacity-50"
        >
          <RefreshCw className={cn('w-3.5 h-3.5', refreshing && 'animate-spin')} />
          重新整理
          {lastUpdated && (
            <span className="text-[#555] normal-case">
              · {lastUpdated.toLocaleTimeString('zh-TW', { hour: '2-digit', minute: '2-digit' })}
            </span>
          )}
        </button>
      </header>

      <div className="flex flex-wrap gap-2">
        {(Object.keys(FILTER_LABELS) as MatchFilter[]).map((key) => (
          <button
            key={key}
            type="button"
            onClick={() => setFilter(key)}
            className={cn(
              'px-3 py-1.5 text-[10px] font-mono uppercase tracking-widest border transition-colors',
              filter === key
                ? 'border-cyan-500/50 bg-cyan-500/10 text-cyan-400'
                : 'border-[#2a2a2a] text-[#666] hover:border-[#444] hover:text-slate-300',
            )}
          >
            {FILTER_LABELS[key]}
            <span className="ml-1.5 text-[#555]">({counts[key]})</span>
          </button>
        ))}
      </div>

      {error && (
        <div className="border border-orange-500/30 bg-orange-500/10 p-3 text-[10px] font-mono text-orange-400">
          {error}
        </div>
      )}

      {loading && !matchesData ? (
        <div className="flex items-center justify-center py-16 text-[#666]">
          <Loader2 className="w-5 h-5 animate-spin mr-2" />
          <span className="text-[10px] font-mono tracking-widest uppercase">載入賽程中…</span>
        </div>
      ) : displayMatches.length === 0 ? (
        <div className="text-center py-16 text-[#666] text-[10px] font-mono tracking-widest uppercase">
          此篩選條件下沒有比賽
        </div>
      ) : (
        <div className="space-y-3">
          {displayMatches.map((match) => (
            <MatchRow
              key={match.id}
              match={match}
              signal={signalMap.get(String(match.id)) ?? null}
              expanded={expandedId === match.id}
              onToggle={() => setExpandedId((id) => (id === match.id ? null : match.id))}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function MatchRow({
  match,
  signal,
  expanded,
  onToggle,
}: {
  match: MatchResponse;
  signal: MatchSignal | null;
  expanded: boolean;
  onToggle: () => void;
}) {
  const stageLabel = STAGE_LABELS[match.stage] ?? match.stage;
  const status = resolveMatchStatus(match);
  const isLive = status === 'live';
  const isFinished = status === 'finished';
  const score =
    match.home_score != null && match.away_score != null
      ? `${match.home_score} : ${match.away_score}`
      : null;

  return (
    <div
      className={cn(
        'bg-[#0f0f0f] border transition-colors',
        isLive ? 'border-orange-500/40' : 'border-[#2a2a2a]',
        expanded && 'border-cyan-500/30',
      )}
    >
      <button
        type="button"
        onClick={onToggle}
        className="w-full text-left p-4 hover:bg-[#121212] transition-colors"
      >
        <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
          <div className="flex items-center gap-2 flex-wrap">
            <CalendarDays className="w-3.5 h-3.5 text-cyan-400 shrink-0" />
            <span className="text-[10px] font-bold text-slate-300 font-mono tracking-widest">
              {formatTaipeiDateTime(match.kickoff)} 台灣
            </span>
            {isLive && (
              <span className="flex items-center gap-1 text-[8px] font-mono font-bold px-1.5 py-0.5 border text-orange-400 border-orange-500/40 bg-orange-500/10 tracking-wider">
                <Radio className="w-3 h-3 animate-pulse" />
                LIVE {match.time_elapsed !== 'notstarted' ? match.time_elapsed : ''}
              </span>
            )}
            {isFinished && !isLive && (
              <span className="text-[8px] font-mono font-bold px-1.5 py-0.5 border text-[#888] border-[#333] tracking-wider">
                已結束
              </span>
            )}
            {!isLive && !isFinished && (
              <span className="text-[8px] font-mono font-bold px-1.5 py-0.5 border text-[#666] border-[#333] tracking-wider">
                未開賽
              </span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <span className="text-[9px] font-mono bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 px-2 py-0.5 tracking-wider">
              {stageLabel} · MD{match.matchday || '—'}
            </span>
            {expanded ? (
              <ChevronUp className="w-4 h-4 text-[#666]" />
            ) : (
              <ChevronDown className="w-4 h-4 text-[#666]" />
            )}
          </div>
        </div>

        <div className="flex items-center justify-between gap-4">
          <TeamCell id={match.home_id} name={match.home_name} iso2={match.home_iso2} align="left" />
          <div className="flex flex-col items-center shrink-0 px-2">
            <div
              className={cn(
                'text-xl md:text-2xl font-bold font-mono',
                isLive ? 'text-orange-400' : score ? 'text-slate-200' : 'text-[#444] italic',
              )}
            >
              {score ?? 'VS'}
            </div>
            {signal && (
              <div className="text-[9px] text-[#666] font-mono mt-1 whitespace-nowrap">
                {signal.bookmarkOdds.home.toFixed(2)} / {signal.bookmarkOdds.draw.toFixed(2)} /{' '}
                {signal.bookmarkOdds.away.toFixed(2)}
              </div>
            )}
          </div>
          <TeamCell id={match.away_id} name={match.away_name} iso2={match.away_iso2} align="right" />
        </div>
      </button>

      {expanded && <MatchDetail match={match} signal={signal} stageLabel={stageLabel} />}
    </div>
  );
}

function TeamCell({
  id,
  name,
  iso2,
  align,
}: {
  id: string;
  name: string;
  iso2: string;
  align: 'left' | 'right';
}) {
  return (
    <div className={cn('flex items-center gap-3 flex-1 min-w-0', align === 'right' && 'flex-row-reverse')}>
      <NationFlag iso2={iso2 || 'un'} alt={name} className="w-12 h-8 shrink-0" />
      <div className={cn('min-w-0', align === 'right' ? 'text-right' : 'text-left')}>
        <div className="text-sm font-black text-slate-100 font-mono tracking-widest uppercase truncate">{id}</div>
        <div className="text-[9px] text-[#666] font-mono tracking-widest uppercase truncate">{name}</div>
      </div>
    </div>
  );
}

function MatchDetail({
  match,
  signal,
  stageLabel,
}: {
  match: MatchResponse;
  signal: MatchSignal | null;
  stageLabel: string;
}) {
  const bestEv = signal
    ? [
        { label: '主勝', value: signal.ev.home },
        { label: '和局', value: signal.ev.draw },
        { label: '客勝', value: signal.ev.away },
      ].reduce((a, b) => (b.value > a.value ? b : a))
    : null;

  return (
    <div className="border-t border-[#2a2a2a] bg-[#0a0a0a] p-4 space-y-4 animate-in fade-in duration-200">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <InfoBlock title="開賽時間">
          <p className="text-xs text-slate-200 font-mono">{formatTaipeiDateTimeLong(match.kickoff)}（台灣）</p>
          <p className="text-[10px] text-[#666] font-mono mt-1">
            場地當地：{formatVenueLocalTime(match.local_date)}
          </p>
        </InfoBlock>
        <InfoBlock title="賽事資訊">
          <div className="flex items-center gap-2 text-[10px] text-[#888] font-mono">
            <MapPin className="w-3.5 h-3.5 shrink-0" />
            <span className="truncate">{match.stadium_name || '—'}</span>
          </div>
          <div className="flex items-center gap-2 text-[10px] text-[#888] font-mono mt-1">
            <Hash className="w-3.5 h-3.5 shrink-0" />
            <span>
              #{match.wc_match_id} · {stageLabel}
            </span>
          </div>
        </InfoBlock>
      </div>

      {match.status === 'live' && (
        <div className="border border-orange-500/30 bg-orange-500/5 p-3 flex items-center gap-3">
          <Radio className="w-4 h-4 text-orange-400 animate-pulse shrink-0" />
          <div>
            <div className="text-[10px] font-bold text-orange-400 font-mono uppercase tracking-widest">即時賽況</div>
            <div className="text-xs text-slate-200 font-mono mt-0.5">
              比分 {match.home_score ?? 0} : {match.away_score ?? 0}
              {match.time_elapsed && match.time_elapsed !== 'notstarted' ? ` · ${match.time_elapsed}` : ''}
            </div>
            <div className="text-[9px] text-[#666] font-mono mt-1">每 30 秒自動更新 · 資料來源 WorldCup2026 CDN</div>
          </div>
        </div>
      )}

      {signal ? (
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <TrendingUp className="w-4 h-4 text-cyan-400" />
            <span className="text-[10px] font-bold text-slate-200 font-mono uppercase tracking-widest">
              The Odds API 盤口 & 模型 EV
            </span>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left font-mono text-[10px]">
              <thead>
                <tr className="text-[#666] uppercase tracking-widest border-b border-[#2a2a2a]">
                  <th className="pb-2 pr-4 font-normal">結果</th>
                  <th className="pb-2 pr-4 font-normal text-right">賠率</th>
                  <th className="pb-2 pr-4 font-normal text-right">盤口機率</th>
                  <th className="pb-2 pr-4 font-normal text-right">模型機率</th>
                  <th className="pb-2 pr-4 font-normal text-right">EV</th>
                  <th className="pb-2 font-normal text-right text-purple-400/80">Kelly</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#1a1a1a]">
                <OddsRow
                  label={`主勝 ${match.home_id}`}
                  odds={signal.bookmarkOdds.home}
                  implied={signal.impliedProb.home}
                  model={signal.pFinal.home}
                  ev={signal.ev.home}
                  kelly={signal.kelly_fraction?.home ?? 0}
                  highlight={bestEv?.label === '主勝'}
                />
                <OddsRow
                  label="和局"
                  odds={signal.bookmarkOdds.draw}
                  implied={signal.impliedProb.draw}
                  model={signal.pFinal.draw}
                  ev={signal.ev.draw}
                  kelly={signal.kelly_fraction?.draw ?? 0}
                  highlight={bestEv?.label === '和局'}
                />
                <OddsRow
                  label={`客勝 ${match.away_id}`}
                  odds={signal.bookmarkOdds.away}
                  implied={signal.impliedProb.away}
                  model={signal.pFinal.away}
                  ev={signal.ev.away}
                  kelly={signal.kelly_fraction?.away ?? 0}
                  highlight={bestEv?.label === '客勝'}
                />
              </tbody>
            </table>
          </div>

          {bestEv && bestEv.value > 0 && (
            <div className="text-[10px] font-mono text-cyan-400 border border-cyan-500/30 bg-cyan-500/5 px-3 py-2">
              最高正 EV：{bestEv.label} {(bestEv.value * 100).toFixed(1)}%
            </div>
          )}
        </div>
      ) : (
        <div className="text-[10px] text-[#666] font-mono border border-[#2a2a2a] p-3">
          此場尚無 The Odds API 盤口資料。請至「數據管線營運」觸發同步，或該場次尚未開放博彩盤口。
        </div>
      )}
    </div>
  );
}

function InfoBlock({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-[#121212] border border-[#2a2a2a] p-3">
      <div className="text-[9px] text-[#666] font-mono uppercase tracking-widest mb-2">{title}</div>
      {children}
    </div>
  );
}

function OddsRow({
  label,
  odds,
  implied,
  model,
  ev,
  highlight,
  kelly,
}: {
  label: string;
  odds: number;
  implied: number;
  model: number;
  ev: number;
  highlight: boolean;
  kelly: number;
}) {
  return (
    <tr className={cn(highlight && 'bg-cyan-500/5')}>
      <td className="py-2 pr-4 text-slate-300">{label}</td>
      <td className="py-2 pr-4 text-right text-slate-200 font-bold">{odds.toFixed(2)}</td>
      <td className="py-2 pr-4 text-right text-[#888]">{(implied * 100).toFixed(1)}%</td>
      <td className="py-2 pr-4 text-right text-cyan-400/80">{(model * 100).toFixed(1)}%</td>
      <td
        className={cn(
          'py-2 pr-4 text-right font-bold',
          ev > 0 ? 'text-cyan-400' : ev < 0 ? 'text-orange-400/80' : 'text-[#666]',
        )}
      >
        {ev >= 0 ? '+' : ''}
        {(ev * 100).toFixed(1)}%
      </td>
      <td className="py-2 text-right font-bold text-purple-400/80">
        {kelly > 0.001 ? `${(kelly * 100).toFixed(1)}%` : '—'}
      </td>
    </tr>
  );
}
