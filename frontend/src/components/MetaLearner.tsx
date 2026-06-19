import React, { useEffect, useMemo, useState } from 'react';
import { Gauge, TrendingUp, Play, Target, BrainCircuit, Activity, Search, AlertTriangle } from 'lucide-react';
import { cn } from '../lib/utils';
import { NationFlag } from './NationFlag';
import { teamFromResponse } from '../data';
import { useNarrative } from '../hooks/useNarrative';
import { useTeams } from '../hooks/useTeams';
import { fetchMonteCarlo, fetchPrediction } from '../services/api';
import type { NarrativeRequest, PredictResponse, TeamInfo } from '../types';

function favorLabel(side: string | undefined, homeId: string, awayId: string): string {
  if (side === 'home') return homeId;
  if (side === 'away') return awayId;
  return '均勢';
}

function pct(n: number | undefined): string {
  return n != null ? (n * 100).toFixed(1) : '—';
}

export function MetaLearner() {
  const { data: teamsData, loading: teamsLoading } = useTeams();
  const teams = useMemo(() => (teamsData ?? []).map(teamFromResponse), [teamsData]);

  const [teamA, setTeamA] = useState<TeamInfo | null>(null);
  const [teamB, setTeamB] = useState<TeamInfo | null>(null);
  const [predictData, setPredictData] = useState<PredictResponse | null>(null);
  const [predictLoading, setPredictLoading] = useState(false);
  const [predictError, setPredictError] = useState<string | null>(null);

  const [mcState, setMcState] = useState<'idle' | 'running' | 'done'>('idle');
  const [mcIterations, setMcIterations] = useState(0);
  const [mcError, setMcError] = useState<string | null>(null);

  const { data: narrativeData, loading: narrativeLoading, generate: generateNarrative, reset: resetNarrative } =
    useNarrative();

  useEffect(() => {
    if (teams.length < 2) return;
    if (!teamA) setTeamA(teams[0]);
    if (!teamB) setTeamB(teams.find((t) => t.id !== teams[0].id) ?? teams[1]);
  }, [teams, teamA, teamB]);

  const hasPredicted = predictData !== null;
  const isCalculating = predictLoading;

  const buildNarrativeReq = (data: PredictResponse): NarrativeRequest => ({
    home_team: teamA!.name,
    away_team: teamB!.name,
    home_win_prob: data.p_final.home,
    draw_prob: data.p_final.draw,
    away_win_prob: data.p_final.away,
    home_elo: data.elo.home,
    away_elo: data.elo.away,
    home_gdp: teamA!.gdpValue,
    away_gdp: teamB!.gdpValue,
    home_lambda: data.poisson.home_lambda,
    away_lambda: data.poisson.away_lambda,
    w1: data.weights.w1,
    w2: data.weights.w2,
    w3: data.weights.w3,
    venue_label: data.neutral_venue !== false ? '中立場（世界盃）' : '主客場',
    poisson_favors: data.signals?.poisson_favors ?? 'even',
    final_favors: data.signals?.final_favors ?? 'even',
    signal_conflict: data.signals?.conflict ?? false,
  });

  const calculate = async () => {
    if (!teamA || !teamB) return;
    setPredictLoading(true);
    setPredictError(null);
    setPredictData(null);
    resetNarrative();
    setMcState('idle');
    setMcIterations(0);
    try {
      const data = await fetchPrediction(teamA.id, teamB.id);
      setPredictData(data);
    } catch (err) {
      setPredictError((err as Error).message);
    } finally {
      setPredictLoading(false);
    }
  };

  const runMonteCarlo = async () => {
    if (!predictData || !teamA || !teamB) return;
    setMcState('running');
    setMcError(null);
    try {
      const result = await fetchMonteCarlo(teamA.id, teamB.id, 100000);
      setPredictData(result);
      setMcIterations(result.iterations);
      setMcState('done');
      resetNarrative();
      await generateNarrative(buildNarrativeReq(result));
    } catch (err) {
      setMcError((err as Error).message);
      setMcState('idle');
    }
  };

  const weights = predictData?.weights ?? { w1: 0.3, w2: 0.3, w3: 0.4, clip: 0.05 };
  const w3Source = predictData?.sources?.w3;
  const weightLabel = `W1 ${(weights.w1 * 100).toFixed(0)}% · W2 ${(weights.w2 * 100).toFixed(0)}% · W3 ${(weights.w3 * 100).toFixed(0)}%`;
  const topPt = `0,${(-80 * weights.w3).toFixed(2)}`;
  const rightPt = `${(70 * weights.w1).toFixed(2)},${(40 * weights.w1).toFixed(2)}`;
  const leftPt = `${(-70 * weights.w2).toFixed(2)},${(40 * weights.w2).toFixed(2)}`;
  const svgPoints = `${topPt} ${rightPt} ${leftPt}`;

  const poissonWinRates = predictData
    ? {
        home: +((predictData.poisson_w2?.home_win ?? predictData.poisson.home_win) * 100).toFixed(1),
        draw: +((predictData.poisson_w2?.draw ?? predictData.poisson.draw) * 100).toFixed(1),
        away: +((predictData.poisson_w2?.away_win ?? predictData.poisson.away_win) * 100).toFixed(1),
      }
    : { home: 0, draw: 0, away: 0 };

  const blendedWinRates = predictData
    ? {
        home: +(predictData.poisson.home_win * 100).toFixed(1),
        draw: +(predictData.poisson.draw * 100).toFixed(1),
        away: +(predictData.poisson.away_win * 100).toFixed(1),
      }
    : { home: 0, draw: 0, away: 0 };

  const topScore = predictData?.poisson.top_scores[0];
  const isNeutral = predictData?.neutral_venue !== false;
  const signalConflict = predictData?.signals?.conflict ?? false;
  const w3Implied = predictData?.w3_implied;

  if (teamsLoading && teams.length === 0) {
    return (
      <div className="text-[10px] text-[#666] font-mono tracking-widest uppercase py-12 text-center">
        載入球隊資料中…
      </div>
    );
  }

  if (!teamA || !teamB) {
    return (
      <div className="text-[10px] text-orange-400 font-mono tracking-widest uppercase py-12 text-center">
        無法載入球隊清單 — 請確認後端已同步
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 ease-out flex flex-col pb-12">
      <header className="border-b border-[#2a2a2a] pb-4 shrink-0">
        <h2 className="text-xl font-bold tracking-tight text-slate-100 uppercase font-mono">量化預測實驗室</h2>
        <p className="text-[10px] text-[#666] mt-1 font-mono tracking-widest uppercase">
          附帶國旗的模糊搜索對局選擇器與多因子張力交叉比對
        </p>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 relative z-20">
        <TeamSelector
          label="目標球隊 A"
          teams={teams}
          selected={teamA}
          onSelect={(t) => {
            setTeamA(t);
            setPredictData(null);
            resetNarrative();
            setMcState('idle');
          }}
          excludeId={teamB.id}
        />
        <TeamSelector
          label="目標球隊 B"
          teams={teams}
          selected={teamB}
          onSelect={(t) => {
            setTeamB(t);
            setPredictData(null);
            resetNarrative();
            setMcState('idle');
          }}
          excludeId={teamA.id}
        />
      </div>

      {predictError && (
        <div className="border border-orange-500/30 bg-orange-500/10 p-3 text-[10px] font-mono text-orange-400">
          {predictError}
        </div>
      )}

      {!hasPredicted && !isCalculating && (
        <button
          onClick={calculate}
          className="w-full py-4 bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-400 border border-cyan-500/30 font-mono tracking-widest uppercase font-bold flex items-center justify-center gap-3 transition-colors"
        >
          <Play className="w-5 h-5" />
          啟動量化預測模型
        </button>
      )}

      {(isCalculating || hasPredicted) && (
        <div className="relative bg-[#0a0a0a] border border-[#2a2a2a] flex flex-col overflow-hidden min-h-[600px]">
          {isCalculating ? (
            <div className="absolute inset-0 z-10 bg-[#0f0f0f] flex flex-col items-center justify-center">
              <div className="w-full max-w-lg space-y-6 px-6">
                <div className="flex justify-between items-end text-[10px] font-mono text-cyan-400 uppercase tracking-widest">
                  <span>正在重載量化矩陣</span>
                  <span className="animate-pulse">POST /api/predict …</span>
                </div>
                <div className="space-y-4">
                  <div className="h-10 bg-[#1a1a1a] border border-[#2a2a2a] w-full animate-pulse" />
                  <div className="h-20 bg-[#1a1a1a] border border-[#2a2a2a] w-full animate-pulse" />
                </div>
              </div>
            </div>
          ) : (
            <div className="p-6 md:p-8 flex flex-col relative animate-in fade-in duration-300">
              <div className="flex justify-between items-center mb-8 border-b border-[#2a2a2a] pb-6 relative z-10">
                <div className="flex items-center gap-4 text-3xl md:text-5xl uppercase font-black font-mono tracking-widest">
                  <NationFlag iso2={teamA.iso2} className="w-16 h-11 md:w-24 md:h-16" />
                  <span className="text-slate-100">{teamA.id}</span>
                </div>
                <div className="text-slate-600 font-mono italic font-bold">VS</div>
                <div className="flex items-center gap-4 text-3xl md:text-5xl uppercase font-black font-mono tracking-widest flex-row-reverse">
                  <NationFlag iso2={teamB.iso2} className="w-16 h-11 md:w-24 md:h-16" />
                  <span className="text-slate-100">{teamB.id}</span>
                </div>
              </div>

              <div className="mb-6 border border-cyan-500/30 bg-cyan-500/5 px-4 py-3 space-y-2">
                <p className="text-[10px] font-mono text-cyan-400 leading-relaxed">
                  <span className="font-bold uppercase tracking-widest">下注依據 →</span>{' '}
                  下注只看<span className="text-orange-400">中欄「融合勝率 p_final」</span>；左欄為融合 λ 的比分分布；右欄 W2 基線供對照。
                </p>
                {isNeutral && (
                  <p className="text-[9px] font-mono text-[#888]">
                    場地模式：<span className="text-cyan-400">中立場</span>（世界盃預設，無主場 λ×1.1 / ELO+100 加成）
                  </p>
                )}
                {signalConflict && teamA && teamB && (
                  <p className="text-[9px] font-mono text-orange-400 border border-orange-500/30 bg-orange-500/5 px-2 py-1.5">
                    <AlertTriangle className="w-3 h-3 inline mr-1 align-text-bottom" />
                    W2 xG/ELO 進球偏向 {favorLabel(predictData?.signals?.w2_lambda_favors, teamA.id, teamB.id)}，
                    W3 盤口進球偏向 {favorLabel(predictData?.signals?.w3_lambda_favors, teamA.id, teamB.id)}（最終 λ 已加權融合）
                  </p>
                )}
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 relative z-10 w-full mb-8">
                <div className="bg-[#121212] border border-[#2a2a2a] p-5 flex flex-col">
                  <div className="flex items-center gap-2 mb-4">
                    <Target className="w-4 h-4 text-cyan-400" />
                    <h3 className="text-[11px] font-bold text-slate-200 font-mono uppercase tracking-widest">
                      預期比分 (Poisson)
                    </h3>
                  </div>
                  <div className="flex-1 flex flex-col justify-center space-y-5">
                    <div>
                      <div className="text-[9px] text-[#666] font-mono tracking-widest uppercase mb-1">
                        預期進球 λ
                      </div>
                      <div className="text-2xl font-black font-mono tracking-widest flex justify-between items-center">
                        <span className="text-slate-200">{predictData!.poisson.home_lambda.toFixed(2)}</span>
                        <span className="text-[#444] text-sm">VS</span>
                        <span className="text-slate-200">{predictData!.poisson.away_lambda.toFixed(2)}</span>
                      </div>
                      <div className="text-[9px] text-[#555] font-mono mt-1">
                        {teamA.id} vs {teamB.id} · λ 融合 · {isNeutral ? '中立場' : '主客場'}
                      </div>
                      {predictData?.lambda_layers && (
                        <div className="border border-[#2a2a2a] bg-[#0a0a0a] px-2 py-2 mt-2 space-y-1">
                          <div className="text-[8px] text-[#555] uppercase tracking-widest mb-1">λ 分層（各因子 λ）</div>
                          <div className="text-[8px] font-mono text-cyan-500/90 flex justify-between gap-2">
                            <span>W2 xG/ELO</span>
                            <span>
                              {teamA.id} {predictData.lambda_layers.w2.home.toFixed(3)} · {teamB.id}{' '}
                              {predictData.lambda_layers.w2.away.toFixed(3)}
                            </span>
                          </div>
                          <div className="text-[8px] font-mono text-purple-400/90 flex justify-between gap-2">
                            <span>W1 陣容</span>
                            <span>
                              {teamA.id} {predictData.lambda_layers.w1.home.toFixed(3)} · {teamB.id}{' '}
                              {predictData.lambda_layers.w1.away.toFixed(3)}
                            </span>
                          </div>
                          <div className="text-[8px] font-mono text-orange-400/90 flex justify-between gap-2">
                            <span>W3 盤口</span>
                            <span>
                              {teamA.id} {predictData.lambda_layers.w3.home.toFixed(3)} · {teamB.id}{' '}
                              {predictData.lambda_layers.w3.away.toFixed(3)}
                            </span>
                          </div>
                          {predictData.lambda_contrib && (
                            <div className="text-[8px] text-[#555] pt-1 border-t border-[#2a2a2a]">
                              加權後 → {teamA.id}{' '}
                              {(
                                predictData.lambda_contrib.w2.home +
                                predictData.lambda_contrib.w1.home +
                                predictData.lambda_contrib.w3.home
                              ).toFixed(2)}{' '}
                              · {teamB.id}{' '}
                              {(
                                predictData.lambda_contrib.w2.away +
                                predictData.lambda_contrib.w1.away +
                                predictData.lambda_contrib.w3.away
                              ).toFixed(2)}{' '}
                              （= 融合 λ）
                            </div>
                          )}
                          {predictData.lambda_layers?.w3.implied_total != null && (
                            <div className="text-[8px] text-[#666] pt-1 border-t border-[#2a2a2a]">
                              {predictData.lambda_layers.w3.total_source === 'totals' ? (
                                <>
                                  W3 隱含總進球（大小分 {predictData.w3_totals?.line?.toFixed(1) ?? '2.5'}）：{' '}
                                  {predictData.lambda_layers.w3.implied_total.toFixed(2)} 球
                                  {predictData.w3_totals && (
                                    <span className="text-[#555] ml-1">
                                      · O {predictData.w3_totals.over.toFixed(2)} / U{' '}
                                      {predictData.w3_totals.under.toFixed(2)}
                                    </span>
                                  )}
                                </>
                              ) : (
                                <>
                                  W3 隱含總進球（和局 proxy）：{' '}
                                  {predictData.lambda_layers.w3.implied_total.toFixed(2)} 球
                                </>
                              )}
                            </div>
                          )}
                        </div>
                      )}
                      <p className="text-[9px] text-[#666] font-mono mt-2 leading-relaxed">
                        國際賽單隊 λ 多在 0.9–1.8；W3 優先使用 The Odds API 大小分（2.5 球線），無盤口時才以和局機率估總進球。
                      </p>
                    </div>
                    {topScore && (
                      <div className="border border-[#2a2a2a] bg-[#0a0a0a] px-3 py-2">
                        <div className="text-[9px] text-[#666] font-mono tracking-widest uppercase mb-1">
                          最可能比分
                        </div>
                        <div className="flex justify-between items-center font-mono">
                          <span className="text-lg font-black text-slate-100">
                            {topScore.home} - {topScore.away}
                          </span>
                          <span className="text-sm text-cyan-400">{(topScore.prob * 100).toFixed(1)}%</span>
                        </div>
                      </div>
                    )}
                    <div>
                      <div className="text-[9px] text-[#666] font-mono tracking-widest uppercase mb-2">
                        其他可能比分 Top 3
                      </div>
                      <div className="space-y-1.5">
                        {predictData!.poisson.top_scores.slice(1, 4).map((item, idx) => (
                          <div key={idx} className="flex justify-between items-center text-[10px] font-mono">
                            <span className="text-slate-400">
                              {item.home} - {item.away}
                            </span>
                            <span className="text-[#888]">{(item.prob * 100).toFixed(1)}%</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>

                <div className="bg-[#121212] border border-orange-500/40 p-5 flex flex-col relative overflow-hidden">
                  <div className="absolute top-0 right-0 px-2 py-0.5 bg-orange-500/20 text-[8px] font-mono text-orange-400 uppercase tracking-widest">
                    下注依據
                  </div>
                  <div className="flex items-center gap-2 mb-4 mt-1">
                    <TrendingUp className="w-4 h-4 text-orange-400" />
                    <h3 className="text-[11px] font-bold text-slate-200 font-mono uppercase tracking-widest">
                      融合勝率 (p_final)
                    </h3>
                  </div>
                  <div className="flex-1 flex flex-col justify-center space-y-4">
                    <div className="text-[9px] text-orange-500/80 font-mono tracking-widest uppercase">
                      權重 {weightLabel} · 模型最終輸出
                    </div>
                    <div className="grid grid-cols-3 gap-2 text-center">
                      <div>
                        <div className="text-[9px] text-[#666] font-mono mb-1">主勝 {teamA.id}</div>
                        <div className="text-xl font-black text-orange-400 font-mono">{blendedWinRates.home}%</div>
                      </div>
                      <div>
                        <div className="text-[9px] text-[#666] font-mono mb-1">和局</div>
                        <div className="text-xl font-black text-slate-300 font-mono">{blendedWinRates.draw}%</div>
                      </div>
                      <div>
                        <div className="text-[9px] text-[#666] font-mono mb-1">客勝 {teamB.id}</div>
                        <div className="text-xl font-black text-orange-400 font-mono">{blendedWinRates.away}%</div>
                      </div>
                    </div>
                    <p className="text-[9px] text-[#555] font-mono leading-relaxed">
                      與左欄 Poisson 勝率相同（p_final = 融合 λ 蒙地卡羅）。W3 隱含勝率見下方，僅供對照盤口。
                    </p>
                    {w3Source === 'elo_fallback' && (
                      <div className="inline-flex items-center gap-1 text-[9px] font-mono text-orange-400 border border-orange-500/30 px-2 py-0.5">
                        <AlertTriangle className="w-3 h-3" />
                        W3 使用 ELO 代替（無盤口數據）
                      </div>
                    )}
                    {w3Source === 'odds' && (
                      <div className="text-[9px] font-mono text-[#666] space-y-2">
                        <div>W3 來源：The Odds API 去抽水盤口</div>
                        {w3Implied && (
                          <div className="border border-[#2a2a2a] bg-[#0a0a0a] px-2 py-2">
                            <div className="text-[8px] text-[#555] uppercase tracking-widest mb-1.5">W3 盤口 1X2 隱含勝率（對照）</div>
                            <div className="grid grid-cols-3 gap-1 text-center">
                              <div>
                                <div className="text-[#666]">{teamA.id}</div>
                                <div className="text-orange-300 font-bold">{pct(w3Implied.home)}%</div>
                              </div>
                              <div>
                                <div className="text-[#666]">和</div>
                                <div className="text-slate-300 font-bold">{pct(w3Implied.draw)}%</div>
                              </div>
                              <div>
                                <div className="text-[#666]">{teamB.id}</div>
                                <div className="text-orange-300 font-bold">{pct(w3Implied.away)}%</div>
                              </div>
                            </div>
                            {predictData?.w3_odds && (
                              <div className="text-[8px] text-[#555] mt-1.5">
                                賠率 {predictData.w3_odds.home.toFixed(2)} / {predictData.w3_odds.draw.toFixed(2)} / {predictData.w3_odds.away.toFixed(2)}
                              </div>
                            )}
                            <p className="text-[8px] text-[#555] mt-1.5 leading-relaxed">
                              盤口主勝 {pct(w3Implied.home)}% ≠ p_final {blendedWinRates.home}% 屬正常：p_final 由融合 λ 模擬，不是 1X2 加權平均。
                            </p>
                          </div>
                        )}
                      </div>
                    )}
                    <p className="text-[9px] text-[#555] font-mono leading-relaxed">
                      到「小組賽況」展開同場比賽，對照模型機率與 EV / Kelly 欄下單。
                    </p>
                  </div>
                </div>

                <div className="bg-[#121212] border border-[#2a2a2a] p-5 flex flex-col">
                  <div className="flex items-center gap-2 mb-4">
                    <Activity className="w-4 h-4 text-cyan-400" />
                    <h3 className="text-[11px] font-bold text-slate-200 font-mono uppercase tracking-widest">
                      W2 純競技基線 (xG/ELO)
                    </h3>
                  </div>
                  <div className="flex flex-col justify-center space-y-5 flex-1">
                    {predictData?.lambda_layers && (
                      <div className="text-[9px] text-[#666] font-mono border-b border-[#2a2a2a] pb-3">
                        W2 λ：{predictData.lambda_layers.w2.home.toFixed(2)} vs{' '}
                        {predictData.lambda_layers.w2.away.toFixed(2)}
                      </div>
                    )}
                    <div className="grid grid-cols-3 gap-2 text-center border-b border-[#2a2a2a] pb-3">
                      <div>
                        <div className="text-[9px] text-[#666] font-mono mb-1">主勝</div>
                        <div className="text-lg font-bold text-cyan-400 font-mono">{poissonWinRates.home}%</div>
                      </div>
                      <div>
                        <div className="text-[9px] text-[#666] font-mono mb-1">和局</div>
                        <div className="text-lg font-bold text-slate-300 font-mono">{poissonWinRates.draw}%</div>
                      </div>
                      <div>
                        <div className="text-[9px] text-[#666] font-mono mb-1">客勝</div>
                        <div className="text-lg font-bold text-orange-400/90 font-mono">{poissonWinRates.away}%</div>
                      </div>
                    </div>
                    <p className="text-[9px] text-[#555] font-mono leading-relaxed">
                      僅 W2 λ 模擬，不含 W1 / W3。W3 盤口把 USA λ 拉到{' '}
                      {predictData?.lambda_layers?.w3.home.toFixed(2) ?? '—'}（融合後{' '}
                      {predictData?.poisson.home_lambda.toFixed(2) ?? '—'}）。
                    </p>
                    <div className="flex flex-col items-center">
                      <svg width="140" height="140" viewBox="-100 -100 200 200" className="opacity-90">
                        <polygon points="0,-80 70,40 -70,40" fill="none" stroke="#2a2a2a" strokeWidth="1" />
                        <polygon points="0,-40 35,20 -35,20" fill="none" stroke="#2a2a2a" strokeWidth="1" />
                        <line x1="0" y1="0" x2="0" y2="-100" stroke="#2a2a2a" strokeWidth="1" />
                        <line x1="0" y1="0" x2="86.6" y2="50" stroke="#2a2a2a" strokeWidth="1" />
                        <line x1="0" y1="0" x2="-86.6" y2="50" stroke="#2a2a2a" strokeWidth="1" />
                        <polygon points={svgPoints} fill="rgba(6,182,212,0.2)" stroke="#06b6d4" strokeWidth="2" />
                        <text x="0" y="-85" textAnchor="middle" fill="#888" fontSize="9" fontFamily="monospace">
                          W3
                        </text>
                        <text x="75" y="50" textAnchor="start" fill="#888" fontSize="9" fontFamily="monospace">
                          W1
                        </text>
                        <text x="-75" y="50" textAnchor="end" fill="#888" fontSize="9" fontFamily="monospace">
                          W2
                        </text>
                      </svg>
                      <div className="text-[8px] text-[#555] font-mono mt-1">三因子權重三角</div>
                    </div>
                  </div>
                </div>
              </div>

              <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-5 mb-8 relative overflow-hidden group">
                <div className="flex items-center gap-2 mb-3">
                  <div className="text-purple-400 font-mono font-bold text-xs">[LLM Narrative]</div>
                  <h3 className="text-[11px] font-bold text-slate-200 font-mono uppercase tracking-widest">
                    協議偵聽終端 (Protocol Terminal)
                  </h3>
                </div>
                <div className="bg-[#121212] border border-[#1a1a1a] p-4 font-mono text-[10px] text-[#888] leading-relaxed">
                  {narrativeLoading && <p className="animate-pulse text-purple-400/70">正在生成 AI 敘事分析…</p>}
                  {narrativeData && !narrativeLoading && (
                    <>
                      <p className="mb-2">
                        <span className="text-cyan-500 font-bold">OUTCOME:</span> {narrativeData.narrative}
                      </p>
                      <p>
                        <span className="text-orange-400 font-bold">CONFIDENCE:</span>
                        <span className="text-purple-400"> {(narrativeData.confidence * 100).toFixed(1)}%</span>
                      </p>
                    </>
                  )}
                  {!narrativeData && !narrativeLoading && mcState === 'idle' && (
                    <p className="text-[#555]">執行蒙特卡洛模擬（100,000 次）後自動觸發 LLM 敘事分析。</p>
                  )}
                </div>
              </div>

              <div className="bg-[#121212] border border-[#2a2a2a] border-l-4 border-l-cyan-500 p-5 mb-8">
                <div className="flex items-center gap-2 mb-3">
                  <BrainCircuit className="w-4 h-4 text-cyan-400" />
                  <h3 className="text-[11px] font-bold text-slate-200 font-mono uppercase tracking-widest">AI敘事側寫情報彙編</h3>
                </div>
                <p className="text-xs text-[#888] font-mono leading-relaxed">
                  {narrativeData?.narrative ?? '完成蒙地卡洛高迭代仿真後，將由 LLM（NVIDIA NIM / DeepSeek）生成敘事。'}
                </p>
              </div>

              <div className="grid grid-cols-1 xl:grid-cols-3 gap-6 relative z-10 w-full mb-8">
                <div className="bg-[#121212] border border-[#2a2a2a] p-5">
                  <h3 className="text-[10px] font-bold text-slate-200 font-mono uppercase tracking-widest border-b border-[#2a2a2a] pb-3 mb-4 flex items-center gap-2">
                    <NationFlag iso2={teamA.iso2} className="w-5 h-3" /> {teamA.id} 因子調控
                  </h3>
                  <div className="text-[10px] text-[#666] font-mono tracking-widest space-y-2">
                    <div className="flex justify-between">
                      <span>ELO 等級</span>
                      <span className="text-cyan-400">{predictData!.elo.home}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>預期進球 λ</span>
                      <span className="text-cyan-400">{predictData!.poisson.home_lambda.toFixed(2)}</span>
                    </div>
                  </div>
                </div>

                <div className="bg-[#121212] border border-[#2a2a2a] p-5">
                  <h3 className="text-[10px] font-bold text-slate-200 font-mono uppercase tracking-widest border-b border-[#2a2a2a] pb-3 mb-4 flex items-center gap-2 flex-row-reverse text-right justify-end">
                    {teamB.id} 因子調控 <NationFlag iso2={teamB.iso2} className="w-5 h-3" />
                  </h3>
                  <div className="text-[10px] text-[#666] font-mono tracking-widest space-y-2">
                    <div className="flex justify-between">
                      <span>ELO 等級</span>
                      <span className="text-orange-400">{predictData!.elo.away}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>預期進球 λ</span>
                      <span className="text-orange-400">{predictData!.poisson.away_lambda.toFixed(2)}</span>
                    </div>
                  </div>
                </div>

                <div className="bg-[#0f0f0f] border border-cyan-500/30 p-5 flex flex-col justify-center relative overflow-hidden">
                  <h3 className="text-[11px] font-bold text-cyan-400 font-mono uppercase tracking-widest mb-1 text-center">
                    加速蒙特卡洛模擬
                  </h3>
                  <p className="text-[9px] text-[#666] font-mono tracking-widest uppercase mb-4 text-center">
                    後端真實 Poisson 仿真
                  </p>
                  {mcError && (
                    <p className="text-[9px] text-orange-400 font-mono text-center mb-2">{mcError}</p>
                  )}
                  {mcState === 'running' ? (
                    <div className="w-full flex-1 flex flex-col justify-center items-center gap-2">
                      <div className="text-[10px] font-mono text-cyan-400 animate-pulse">運算中… 100,000 iterations</div>
                      <div className="w-full h-2 bg-[#1a1a1a] border border-[#2a2a2a] overflow-hidden">
                        <div className="h-full bg-cyan-500 w-full animate-pulse" />
                      </div>
                    </div>
                  ) : (
                    <button
                      onClick={runMonteCarlo}
                      disabled={!predictData}
                      className="w-full py-4 border border-cyan-500 text-cyan-400 hover:bg-cyan-500/10 font-mono text-[10px] tracking-widest uppercase font-bold transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
                    >
                      {mcState === 'done' ? `已完成 ${mcIterations.toLocaleString()} 次 · 重新執行` : '執行 100,000 次仿真演算'}
                    </button>
                  )}
                </div>
              </div>

              <div className="flex items-center justify-center mb-6">
                <div className="px-4 py-1 border border-[#2a2a2a] bg-[#121212] text-[9px] text-[#666] font-mono tracking-widest uppercase">
                  底層真實因子映射 (Real Factor Metrics)
                </div>
              </div>

              <div className="space-y-12 relative z-10 w-full max-w-4xl mx-auto mb-4">
                <MetricComparisonRow
                  title="Kaggle ELO 等級 (真實競技因子)"
                  valA={predictData!.elo.home}
                  valB={predictData!.elo.away}
                  color="cyan"
                />
                <MetricComparisonRow
                  title="World Bank 人均 GDP (USD)"
                  valA={teamA.gdpValue}
                  valB={teamB.gdpValue}
                  displayA={teamA.gdp}
                  displayB={teamB.gdp}
                  color="orange"
                />
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function TeamSelector({
  label,
  teams,
  selected,
  onSelect,
  excludeId,
}: {
  label: string;
  teams: TeamInfo[];
  selected: TeamInfo;
  onSelect: (t: TeamInfo) => void;
  excludeId: string;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [search, setSearch] = useState('');

  const filtered = teams.filter(
    (t) =>
      t.id !== excludeId &&
      (t.name.toLowerCase().includes(search.toLowerCase()) || t.id.toLowerCase().includes(search.toLowerCase())),
  );

  return (
    <div className="relative">
      <div className="flex items-center justify-between mb-2">
        <label className="text-[10px] font-bold text-[#666] font-mono uppercase tracking-widest">{label}</label>
      </div>
      <div
        className="bg-[#0f0f0f] border border-[#2a2a2a] p-3 flex items-center justify-between cursor-pointer hover:border-[#444] transition-colors"
        onClick={() => setIsOpen(!isOpen)}
      >
        <div className="flex items-center gap-3 font-mono font-bold text-slate-200 uppercase tracking-widest">
          <NationFlag iso2={selected.iso2} className="w-7 h-5" />
          <span>
            {selected.name} <span className="text-[#555] text-[10px] ml-2">({selected.id})</span>
          </span>
        </div>
        <Search className="w-4 h-4 text-[#555]" />
      </div>
      {isOpen && (
        <div className="absolute top-full mt-2 w-full bg-[#121212] border border-[#333] z-50 overflow-hidden">
          <div className="p-2 border-b border-[#2a2a2a]">
            <input
              autoFocus
              type="text"
              placeholder="搜尋國家..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full bg-[#0a0a0a] border border-[#2a2a2a] p-2 text-xs font-mono tracking-wider text-slate-200 focus:border-cyan-500 outline-none"
            />
          </div>
          <div className="max-h-60 overflow-y-auto">
            {filtered.map((t) => (
              <div
                key={t.id}
                onClick={() => {
                  onSelect(t);
                  setIsOpen(false);
                  setSearch('');
                }}
                className="p-3 border-b border-[#1a1a1a] last:border-0 hover:bg-[#161616] cursor-pointer flex items-center gap-3 font-mono text-slate-300 transition-colors group"
              >
                <NationFlag iso2={t.iso2} className="w-6 h-4 opacity-70 group-hover:opacity-100" />
                <span className="uppercase tracking-widest text-xs font-bold">{t.name}</span>
                <span className="text-[10px] text-[#555]">{t.id}</span>
              </div>
            ))}
            {filtered.length === 0 && (
              <div className="p-4 text-center text-[10px] text-[#555] font-mono uppercase">無匹配的國家隊</div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function MetricComparisonRow({
  title,
  valA,
  valB,
  displayA,
  displayB,
  color,
}: {
  title: string;
  valA: number;
  valB: number;
  displayA?: string;
  displayB?: string;
  color: 'cyan' | 'orange';
}) {
  const isCyan = color === 'cyan';
  const colorClass = isCyan ? 'text-cyan-400 bg-cyan-500' : 'text-orange-400 bg-orange-500';
  const textColorClass = isCyan ? 'text-cyan-400' : 'text-orange-400';
  const total = valA + valB;
  const pctA = total > 0 ? (valA / total) * 100 : 50;
  const pctB = total > 0 ? (valB / total) * 100 : 50;

  return (
    <div className="space-y-4">
      <div className="text-center text-[10px] font-bold text-[#888] font-mono uppercase tracking-widest mb-4 flex justify-center items-center gap-2">
        {isCyan ? <Gauge className="w-3 h-3" /> : <TrendingUp className="w-3 h-3" />}
        {title}
      </div>
      <div className="flex justify-between items-end font-mono font-bold px-2">
        <div className={cn('text-xl tracking-widest', textColorClass)}>{displayA ?? valA}</div>
        <div className={cn('text-xl tracking-widest', textColorClass)}>{displayB ?? valB}</div>
      </div>
      <div className="w-full h-1.5 bg-[#161616] border border-[#2a2a2a] flex overflow-hidden">
        <div className={colorClass.split(' ')[1]} style={{ width: `${pctA}%` }} />
        <div className="h-full bg-transparent w-0.5 z-10 shrink-0" />
        <div className={cn('opacity-40', colorClass.split(' ')[1])} style={{ width: `${pctB}%` }} />
      </div>
      <div className="flex justify-between text-[9px] text-[#555] font-mono tracking-widest">
        <span>{pctA.toFixed(1)}%</span>
        <span>{pctB.toFixed(1)}%</span>
      </div>
    </div>
  );
}
