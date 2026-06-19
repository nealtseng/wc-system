import React, { useEffect, useMemo, useState } from 'react';
import { Trophy, Swords, CheckCircle2, Clock } from 'lucide-react';
import { cn } from '../lib/utils';
import { NationFlag } from './NationFlag';
import { useMatches } from '../hooks/useMatches';
import { useSignals } from '../hooks/useSignals';
import { buildTournamentMilestones, tournamentProgressPercent } from '../lib/tournament';
import { formatTaipeiDateTimeLong } from '../lib/datetime';
import type { MatchSignal } from '../types';

function useCountdown(kickoff: string | undefined) {
  const [remaining, setRemaining] = useState({ hours: 0, minutes: 0, seconds: 0 });

  useEffect(() => {
    if (!kickoff) return;

    const tick = () => {
      const diff = new Date(kickoff).getTime() - Date.now();
      if (diff <= 0) {
        setRemaining({ hours: 0, minutes: 0, seconds: 0 });
        return;
      }
      setRemaining({
        hours: Math.floor(diff / 3_600_000),
        minutes: Math.floor((diff % 3_600_000) / 60_000),
        seconds: Math.floor((diff % 60_000) / 1_000),
      });
    };

    tick();
    const id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, [kickoff]);

  return remaining;
}

function pad(n: number) {
  return String(n).padStart(2, '0');
}

function findSignal(signals: MatchSignal[] | null, matchId: number | undefined) {
  if (!signals || matchId == null) return null;
  return signals.find((s) => s.id === String(matchId));
}

function bestEvLabel(signal: MatchSignal | null): { text: string; value: number } | null {
  if (!signal) return null;
  const entries = [
    { side: '主勝', value: signal.ev.home },
    { side: '和局', value: signal.ev.draw },
    { side: '客勝', value: signal.ev.away },
  ];
  const best = entries.reduce((a, b) => (b.value > a.value ? b : a));
  return { text: best.side, value: best.value };
}

export function Overview({ onViewChange }: { onViewChange?: (view: any) => void }) {
  const { data: matchesData } = useMatches();
  const { data: signals } = useSignals();

  const focusMatch = matchesData
    ? matchesData.find((m) => new Date(m.kickoff) > new Date()) ?? matchesData[matchesData.length - 1]
    : null;
  const countdown = useCountdown(focusMatch?.kickoff);

  const milestones = useMemo(
    () => buildTournamentMilestones(matchesData ?? []),
    [matchesData],
  );
  const progressPct = tournamentProgressPercent(matchesData ?? []);
  const focusSignal = findSignal(signals, focusMatch?.id);
  const bestEv = bestEvLabel(focusSignal);

  const systemMessage = focusSignal
    ? `[系統訊息] The Odds API 盤口：主 ${focusSignal.bookmarkOdds.home.toFixed(2)} / 和 ${focusSignal.bookmarkOdds.draw.toFixed(2)} / 客 ${focusSignal.bookmarkOdds.away.toFixed(2)}。模型融合機率：主勝 ${(focusSignal.pFinal.home * 100).toFixed(1)}% / 和 ${(focusSignal.pFinal.draw * 100).toFixed(1)}% / 客勝 ${(focusSignal.pFinal.away * 100).toFixed(1)}%。${
        bestEv && bestEv.value > 0
          ? `最高 EV 為${bestEv.text} (${(bestEv.value * 100).toFixed(1)}%)。`
          : '目前無顯著正 EV 信號。'
      }`
    : focusMatch
      ? `[系統訊息] 賽程已載入，此場尚無盤口或信號資料。開賽（台灣）：${formatTaipeiDateTimeLong(focusMatch.kickoff)}`
      : '[系統訊息] 等待賽程同步…';

  return (
    <div className="flex flex-col space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-500 ease-out h-[calc(100vh-2rem)] md:h-[calc(100vh-4rem)] pb-4">
      <header className="shrink-0 border-b border-[#2a2a2a] pb-4">
        <h2 className="text-xl font-bold tracking-tight text-slate-100 uppercase font-mono">全域監控面板</h2>
        <p className="text-[10px] text-[#666] mt-1 font-mono tracking-widest uppercase">世界盃賽程進度與核心焦點對局</p>
      </header>

      <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-4 md:p-6 flex flex-col shrink-0">
        <div className="flex items-center justify-between gap-4 mb-6">
          <div className="flex items-center gap-2">
            <Trophy className="w-4 h-4 text-cyan-500/50" />
            <h3 className="text-xs font-semibold text-slate-200 uppercase font-mono tracking-widest">錦標賽賽程進度</h3>
          </div>
          <span className="text-[9px] text-[#666] font-mono tracking-widest">
            {matchesData ? `${progressPct}% 場次已結束` : '載入中…'}
          </span>
        </div>

        <div className="relative px-2 md:px-8">
          <div className="absolute top-3 left-10 right-10 h-[1px] bg-[#2a2a2a] hidden md:block" />
          <div
            className="absolute top-3 left-10 h-[1px] bg-cyan-500 hidden md:block transition-all duration-500"
            style={{ width: `${Math.max(5, progressPct)}%` }}
          />

          <div className="relative z-10 flex flex-col md:flex-row justify-between gap-4 md:gap-0">
            {milestones.map((step, idx) => {
              const isCompleted = step.status === 'completed';
              const isActive = step.status === 'active';

              return (
                <div key={step.id} className="flex md:flex-col items-center gap-3 bg-[#0f0f0f] px-2">
                  <div
                    className={cn(
                      'w-6 h-6 flex items-center justify-center border transition-all duration-300 shrink-0',
                      isCompleted
                        ? 'border-cyan-500/40 bg-[#161616] text-cyan-500 text-[10px]'
                        : isActive
                          ? 'border-cyan-400 text-[#0f0f0f] bg-cyan-400 shadow-[0_0_10px_rgba(6,182,212,0.3)] scale-110'
                          : 'border-[#333] text-[#555] bg-[#0f0f0f]',
                    )}
                  >
                    {isCompleted ? (
                      <CheckCircle2 className="w-3 h-3" />
                    ) : isActive ? (
                      <Swords className="w-3 h-3" />
                    ) : (
                      <span className="text-[10px] font-mono font-bold">{idx + 1}</span>
                    )}
                  </div>
                  <div className="md:text-center flex-1 mt-2">
                    <div
                      className={cn(
                        'text-[10px] font-bold uppercase tracking-wider font-mono',
                        isActive ? 'text-cyan-400' : isCompleted ? 'text-slate-400' : 'text-[#555]',
                      )}
                    >
                      {step.title}
                    </div>
                    <div className="text-[9px] text-[#666] font-mono tracking-widest mt-0.5">{step.subtitle}</div>
                    <div className="hidden md:block text-[9px] text-[#444] mt-1 font-mono tracking-widest">
                      {step.date}
                      {step.total > 0 ? ` · ${step.finished}/${step.total}` : ''}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-5 md:p-6 relative overflow-hidden flex flex-col justify-between flex-1 min-h-[300px]">
        <div className="absolute -top-10 -right-10 p-4 opacity-[0.02] pointer-events-none">
          <Trophy className="w-72 h-72" />
        </div>

        <div className="flex justify-between items-center mb-4 relative z-10 shrink-0 border-b border-[#2a2a2a] pb-4">
          <div className="flex items-center gap-2 px-3 py-1 bg-[#161616] border border-[#333]">
            <span className="w-1.5 h-1.5 bg-cyan-500 animate-pulse" />
            <span className="text-[10px] font-bold text-cyan-400 uppercase tracking-widest font-mono">次世代焦點賽事</span>
          </div>
          <div className="text-[9px] text-[#666] font-mono tracking-widest uppercase">
            賽事編號: {focusMatch?.wc_match_id ?? '—'} // {focusMatch?.stage ?? '—'}
          </div>
        </div>

        <div className="flex justify-between items-center relative z-10 overflow-hidden px-2 md:px-8 my-auto py-4">
          <div className="w-5/12 flex flex-col items-center sm:items-end sm:pr-8">
            <div className="flex flex-col sm:flex-row items-center gap-4 mb-4">
              <NationFlag
                iso2={focusMatch?.home_iso2 ?? 'un'}
                alt={focusMatch?.home_name}
                className="w-20 h-14 md:w-28 md:h-20 transition-all opacity-80 hover:opacity-100"
              />
              <div className="flex flex-col items-center sm:items-start text-center sm:text-left">
                <div className="text-2xl md:text-4xl font-black text-slate-100 tracking-widest uppercase font-mono">
                  {focusMatch?.home_id ?? '—'}
                </div>
              </div>
            </div>
          </div>

          <div className="w-2/12 flex flex-col items-center justify-center shrink-0 border-x border-[#2a2a2a] px-4 py-8">
            <div className="text-xl md:text-3xl font-bold text-[#333] italic mb-4 font-mono">
              {focusMatch?.home_score != null && focusMatch?.away_score != null
                ? `${focusMatch.home_score} : ${focusMatch.away_score}`
                : 'VS'}
            </div>
            <div className="flex flex-col items-center gap-2">
              {focusMatch?.status === 'live' && (
                <span className="text-[9px] font-bold text-orange-400 border border-orange-500/40 px-2 py-0.5 font-mono tracking-wider">
                  進行中 {focusMatch.time_elapsed !== 'notstarted' ? focusMatch.time_elapsed : ''}
                </span>
              )}
              {focusMatch?.finished && (
                <span className="text-[9px] font-bold text-[#888] border border-[#333] px-2 py-0.5 font-mono tracking-wider">
                  已結束
                </span>
              )}
              {bestEv && (
                <span
                  className={cn(
                    'text-[10px] font-bold px-2 py-1 font-mono whitespace-nowrap tracking-wider',
                    bestEv.value > 0
                      ? 'text-black bg-cyan-400 shadow-[0_0_10px_rgba(6,182,212,0.3)]'
                      : 'text-[#888] bg-[#161616] border border-[#333]',
                  )}
                >
                  預期價值(EV): {bestEv.value >= 0 ? '+' : ''}{(bestEv.value * 100).toFixed(1)}% ({bestEv.text})
                </span>
              )}
            </div>
          </div>

          <div className="w-5/12 flex flex-col items-center sm:items-start sm:pl-8">
            <div className="flex flex-col sm:flex-row-reverse items-center gap-4 mb-4">
              <NationFlag
                iso2={focusMatch?.away_iso2 ?? 'un'}
                alt={focusMatch?.away_name}
                className="w-20 h-14 md:w-28 md:h-20 transition-all opacity-80 hover:opacity-100"
              />
              <div className="flex flex-col items-center sm:items-end text-center sm:text-right">
                <div className="text-2xl md:text-4xl font-black text-slate-100 tracking-widest uppercase font-mono">
                  {focusMatch?.away_id ?? '—'}
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-[#121212] p-3 md:p-4 border border-[#2a2a2a] flex flex-col md:flex-row items-center justify-between relative z-10 gap-4 shrink-0">
          <div className="flex items-center gap-3">
            <div className="p-1 border border-[#333] bg-[#0a0a0a]">
              <Clock className="w-4 h-4 text-cyan-500" />
            </div>
            <div className="flex flex-col">
              <span className="text-[8px] text-[#666] uppercase tracking-widest font-mono">開賽倒數計時</span>
              <span className="text-sm md:text-base text-slate-200 font-mono font-bold leading-tight tracking-wider">
                <span className="text-cyan-400">{pad(countdown.hours)}</span>h:
                <span className="text-cyan-400">{pad(countdown.minutes)}</span>m:
                <span className="text-[#555]">{pad(countdown.seconds)}</span>s
              </span>
            </div>
          </div>
          <div className="text-[10px] text-[#888] flex-1 md:px-6 leading-relaxed text-center md:text-left font-mono md:border-l border-[#2a2a2a] md:ml-4">
            {systemMessage}
          </div>
          <button
            onClick={() => onViewChange?.('prediction-model')}
            className="whitespace-nowrap px-4 py-2 border border-cyan-500/50 text-cyan-400 hover:bg-cyan-500/10 text-[9px] font-bold uppercase tracking-widest transition-colors font-mono cursor-pointer"
          >
            跳轉量化預測實驗室
          </button>
        </div>
      </div>
    </div>
  );
}
