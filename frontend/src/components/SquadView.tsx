import React, { useEffect, useState } from 'react';
import { NationFlag } from './NationFlag';
import { PlayerPhoto } from './PlayerPhoto';
import { ArrowLeft, UserCircle2, Activity, ShieldAlert, Presentation, Loader2, BarChart3 } from 'lucide-react';
import { cn } from '../lib/utils';
import { SquadPlayerResponse, TeamInfo, TeamResponse } from '../types';
import { fetchSquad, syncPipeline } from '../services/api';

interface SquadViewProps {
  team: TeamInfo;
  live?: TeamResponse;
  onBack: () => void;
}

const SYNC_RETRY_MS = 3000;
const SYNC_MAX_ATTEMPTS = 25;

const FM_TECH_ATTRS = ['角球', '传中', '盘带', '射门', '接球', '任意球', '头球', '远射', '界外球', '盯人', '传球', '罚点球', '抢断', '技术'];
const FM_MENTAL_ATTRS = ['侵略性', '预判', '勇敢', '镇定', '集中', '决断', '意志力', '想象力', '领导力', '无球跑动', '防守站位', '团队合作', '视野', '工作投入'];
const FM_PHYS_ATTRS = ['爆发力', '灵活', '平衡', '弹跳', '体质', '速度', '耐力', '强壮'];
const FM_GK_ATTRS = ['制空能力', '拦截传中', '沟通', '神经指数', '手控球', '大脚开球', '一对一', '反应', '出击', '击球倾向', '手抛球的能力'];
const FM_GROUP_ORDER = ['进攻', '创造', '技术', '速度', '身体', '防守', '精神', '制空'];

function isSyncPendingError(message: string): boolean {
  return message.includes('503') || message.includes('sync_in_progress') || message.includes('not synced yet');
}

function isFMPlayer(p: SquadPlayerResponse): boolean {
  return p.source === 'fmcsv' || (p.ca ?? 0) > 0;
}

function attrBar(value: number, max = 20) {
  const pct = Math.min(100, Math.max(0, (value / max) * 100));
  return (
    <div className="flex items-center gap-2 text-xs font-mono">
      <div className="w-20 h-1 bg-[#2a2a2a] overflow-hidden shrink-0">
        <div className={cn('h-full', value >= 15 ? 'bg-cyan-500' : value >= 12 ? 'bg-emerald-500' : 'bg-orange-500')} style={{ width: `${pct}%` }} />
      </div>
      <span className="w-5 text-right text-slate-300">{value}</span>
    </div>
  );
}

function AttrSection({ title, keys, attrs }: { title: string; keys: string[]; attrs: Record<string, number> }) {
  const items = keys.filter((k) => (attrs[k] ?? 0) > 0);
  if (items.length === 0) return null;
  return (
    <div className="bg-[#121212] border border-[#2a2a2a] p-4">
      <h4 className="text-[10px] font-bold text-slate-200 font-mono uppercase tracking-widest mb-3">{title}</h4>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-2">
        {items.map((k) => (
          <div key={k} className="flex items-center justify-between gap-2">
            <span className="text-[10px] text-[#888] truncate">{k}</span>
            {attrBar(attrs[k])}
          </div>
        ))}
      </div>
    </div>
  );
}

async function loadSquadWithRetry(teamId: string, cancelled: () => boolean): Promise<{ players: SquadPlayerResponse[]; source: string }> {
  for (let attempt = 0; attempt < SYNC_MAX_ATTEMPTS; attempt++) {
    if (cancelled()) {
      throw new Error('cancelled');
    }
    try {
      const data = await fetchSquad(teamId);
      return { players: data.players, source: data.source };
    } catch (err) {
      const message = (err as Error).message;
      if (!isSyncPendingError(message) || attempt === SYNC_MAX_ATTEMPTS - 1) {
        throw err;
      }
      await new Promise((resolve) => setTimeout(resolve, SYNC_RETRY_MS));
    }
  }
  return { players: [], source: 'fifa' };
}

export function SquadView({ team, live, onBack }: SquadViewProps) {
  const [selectedPlayer, setSelectedPlayer] = useState<SquadPlayerResponse | null>(null);
  const [players, setPlayers] = useState<SquadPlayerResponse[]>([]);
  const [squadSource, setSquadSource] = useState('fifa');
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setSyncing(false);
    setError(null);
    setSelectedPlayer(null);

    loadSquadWithRetry(team.id, () => cancelled)
      .then(({ players: loaded, source }) => {
        if (!cancelled) {
          setPlayers(loaded);
          setSquadSource(source);
          if (loaded.length > 0) {
            setSelectedPlayer(loaded[0]);
          }
        }
      })
      .catch((err: Error) => {
        if (!cancelled && err.message !== 'cancelled') {
          setError(err.message);
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [team.id]);

  const handleManualSync = async () => {
    setSyncing(true);
    setError(null);
    setLoading(true);
    try {
      await syncPipeline();
      const { players: loaded, source } = await loadSquadWithRetry(team.id, () => false);
      setPlayers(loaded);
      setSquadSource(source);
      setSelectedPlayer(loaded.length > 0 ? loaded[0] : null);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
      setSyncing(false);
    }
  };

  const fmSource = squadSource === 'fmcsv';

  return (
    <div className="space-y-6 animate-in fade-in duration-300 relative">
      <div className="flex items-center gap-4 border-b border-[#2a2a2a] pb-4 mb-6">
        <button
          onClick={onBack}
          className="p-2 border border-[#2a2a2a] hover:bg-[#1a1a1a] text-[#888] hover:text-slate-200 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
        </button>
        <div className="flex items-center gap-4">
          <NationFlag iso2={team.iso2} alt={team.name} className="w-12 h-8" />
          <div>
            <h2 className="text-xl font-bold tracking-tight text-slate-100 uppercase font-mono">{team.name} / {team.id}</h2>
            <p className="text-[10px] text-[#666] mt-1 font-mono tracking-widest uppercase">
              {fmSource ? 'Squad Roster · FM 球探 CSV · 大頭照 Wikimedia' : 'Squad Roster · FIFA/SoFIFA · 大頭照 Wikimedia'}
            </p>
          </div>
        </div>
      </div>

      {live && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-2">
          <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-4">
            <div className="text-[9px] text-[#666] font-mono uppercase tracking-widest mb-2">Kaggle ELO / 勝率</div>
            <div className="text-lg font-bold text-cyan-400 font-mono">{Math.round(live.elo)}</div>
            <div className="text-[10px] text-[#888] font-mono mt-1">{(live.win_rate * 100).toFixed(1)}% · {live.matches_played} 場</div>
          </div>
          <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-4">
            <div className="text-[9px] text-[#666] font-mono uppercase tracking-widest mb-2">W1 陣容強度</div>
            <div className="text-lg font-bold text-emerald-400 font-mono">{((live.player_strength ?? 0) * 100).toFixed(1)}</div>
            <div className="text-[10px] text-[#888] font-mono mt-1">Top-11 CA 正規化 · W1 delta 40%</div>
          </div>
          <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-4">
            <div className="text-[9px] text-[#666] font-mono uppercase tracking-widest mb-2">World Bank 人均 GDP</div>
            <div className="text-lg font-bold text-slate-200 font-mono">
              {live.gdp_per_capita > 0 ? `$${live.gdp_per_capita.toLocaleString(undefined, { maximumFractionDigits: 0 })}` : '—'}
            </div>
            {live.gdp_year ? (
              <div className="text-[10px] text-[#888] font-mono mt-1">{live.gdp_year} 年</div>
            ) : null}
          </div>
          <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-4">
            <div className="text-[9px] text-[#666] font-mono uppercase tracking-widest mb-2">動量 / 場均進球</div>
            <div className="text-lg font-bold font-mono text-orange-400">{live.momentum >= 0 ? '+' : ''}{live.momentum.toFixed(2)}</div>
            <div className="text-[10px] text-[#888] font-mono mt-1">{live.avg_goals_for.toFixed(2)} xG/場</div>
          </div>
        </div>
      )}

      {live?.wiki_extract && (
        <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-4 mb-4">
          <div className="text-[9px] text-[#666] font-mono uppercase tracking-widest mb-2">Wikimedia 球隊摘要</div>
          <p className="text-[11px] text-slate-300 font-mono leading-relaxed line-clamp-4">{live.wiki_extract}</p>
          {live.wiki_url && (
            <a href={live.wiki_url} target="_blank" rel="noreferrer" className="text-[9px] text-cyan-500 font-mono mt-2 inline-block hover:underline">
              查看 Wikipedia →
            </a>
          )}
        </div>
      )}

      {loading && (
        <div className="flex items-center justify-center gap-2 py-16 text-[#888] font-mono text-xs">
          <Loader2 className="w-4 h-4 animate-spin" />
          {syncing ? '正在同步名單資料…' : '載入名單…（後端首次啟動約需 30–60 秒）'}
        </div>
      )}

      {error && !loading && (
        <div className="bg-[#0f0f0f] border border-orange-500/30 p-4 text-orange-400 text-xs font-mono space-y-3">
          <p>{error}</p>
          <button
            type="button"
            onClick={handleManualSync}
            disabled={syncing}
            className="border border-orange-500/40 px-3 py-1.5 text-[10px] uppercase tracking-widest hover:bg-orange-500/10 transition-colors disabled:opacity-50"
          >
            手動觸發資料同步
          </button>
        </div>
      )}

      {!loading && !error && players.length === 0 && (
        <div className="bg-[#0f0f0f] border border-[#2a2a2a] p-6 text-center space-y-2">
          <p className="text-xs text-[#888] font-mono">
            目前沒有 {team.name}（{team.id}）的球員名單。
          </p>
          <p className="text-[10px] text-[#666] font-mono leading-relaxed">
            可將 FM 球探 CSV 放在 data/fm/players.csv，或等待 SoFIFA / FMScout 同步。
            預測仍會用 ELO / xG / 盤口運作，W1 陣容強度則 fallback 至預設值。
          </p>
        </div>
      )}

      {!loading && !error && players.length > 0 && (
        <div className="grid grid-cols-1 xl:grid-cols-5 gap-6">
          <div className="xl:col-span-3 bg-[#0f0f0f] border border-[#2a2a2a] flex flex-col">
            <div className="overflow-x-auto">
              <table className="w-full text-left font-mono whitespace-nowrap">
                <thead>
                  <tr className="text-[10px] text-[#666] uppercase tracking-widest border-b border-[#2a2a2a] bg-[#121212]">
                    <th className="pb-3 pt-4 px-4 font-normal">#</th>
                    <th className="pb-3 pt-4 px-4 font-normal">球員</th>
                    <th className="pb-3 pt-4 px-4 font-normal">位置</th>
                    <th className="pb-3 pt-4 px-4 font-normal text-right">年齡</th>
                    {fmSource ? (
                      <>
                        <th className="pb-3 pt-4 px-4 font-normal text-right">CA</th>
                        <th className="pb-3 pt-4 px-4 font-normal text-right">PA</th>
                      </>
                    ) : (
                      <th className="pb-3 pt-4 px-4 font-normal text-right">體能</th>
                    )}
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#1a1a1a]">
                  {players.map((p) => (
                    <tr
                      key={p.id}
                      onClick={() => setSelectedPlayer(p)}
                      className={cn(
                        'transition-colors cursor-pointer',
                        selectedPlayer?.id === p.id ? 'bg-[#1a1a1a] border-l-2 border-cyan-500' : 'hover:bg-[#121212] border-l-2 border-transparent',
                      )}
                    >
                      <td className="py-3 px-4">
                        <span className="text-xs font-bold text-slate-400">{p.no || '—'}</span>
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-3">
                          <PlayerPhoto wikiSlug={p.wiki_slug} name={p.name} size="sm" />
                          <div>
                            <span className={cn('text-xs font-bold block', selectedPlayer?.id === p.id ? 'text-cyan-400' : 'text-slate-200')}>
                              {p.name}
                            </span>
                            {p.club && (
                              <span className="text-[9px] text-[#666]">{p.club}</span>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <span className="text-[9px] px-2 py-0.5 border border-cyan-500/30 text-cyan-400 bg-cyan-500/10 uppercase tracking-wider">
                          {p.pos}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-right">
                        <span className="text-xs text-[#888]">{p.age || '—'}</span>
                      </td>
                      {fmSource || isFMPlayer(p) ? (
                        <>
                          <td className="py-3 px-4 text-right">
                            <span className={cn('text-xs font-bold', (p.ca ?? 0) >= 160 ? 'text-cyan-400' : 'text-orange-400')}>{p.ca || '—'}</span>
                          </td>
                          <td className="py-3 px-4 text-right">
                            <span className="text-xs text-[#888]">{p.pa || '—'}</span>
                          </td>
                        </>
                      ) : (
                        <td className="py-3 px-4 text-right">
                          <div className="flex items-center justify-end gap-2">
                            <div className="w-16 h-1 bg-[#2a2a2a] overflow-hidden">
                              <div className={cn('h-full', p.fitness >= 90 ? 'bg-cyan-500' : 'bg-orange-500')} style={{ width: `${p.fitness}%` }} />
                            </div>
                            <span className={cn('text-xs w-6', p.fitness >= 90 ? 'text-cyan-400' : 'text-orange-400')}>{p.fitness}</span>
                          </div>
                        </td>
                      )}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          <div className="xl:col-span-2 bg-[#0f0f0f] border border-[#2a2a2a] flex flex-col p-5 max-h-[85vh] overflow-y-auto">
            {selectedPlayer ? (
              <div className="animate-in fade-in duration-200 space-y-4">
                <div className="flex items-start gap-4 pb-4 border-b border-[#2a2a2a]">
                  <PlayerPhoto wikiSlug={selectedPlayer.wiki_slug} name={selectedPlayer.name} size="lg" />
                  <div>
                    <h3 className="text-lg font-bold text-slate-100 font-mono tracking-widest">{selectedPlayer.name}</h3>
                    <div className="text-[10px] text-cyan-400 font-mono mt-1 tracking-widest">
                      NO. {selectedPlayer.no || '—'} | {selectedPlayer.pos}
                    </div>
                    {selectedPlayer.club && (
                      <div className="text-[10px] text-[#888] font-mono mt-1">{selectedPlayer.club}</div>
                    )}
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div className="bg-[#121212] border border-[#2a2a2a] p-3 text-center">
                    <div className="text-[9px] text-[#666] font-mono tracking-widest mb-1">國家隊</div>
                    <div className="text-sm font-bold text-slate-300 font-mono">
                      {selectedPlayer.apps > 0 || selectedPlayer.goals > 0
                        ? `${selectedPlayer.apps} 場 / ${selectedPlayer.goals} 球`
                        : '—'}
                    </div>
                  </div>
                  <div className="bg-[#121212] border border-[#2a2a2a] p-3 text-center">
                    <div className="text-[9px] text-[#666] font-mono tracking-widest mb-1">身價</div>
                    <div className="text-sm font-bold text-emerald-400 font-mono">{selectedPlayer.value || '—'}</div>
                  </div>
                </div>

                {isFMPlayer(selectedPlayer) ? (
                  <>
                    <div className="bg-[#121212] border border-[#2a2a2a] p-4">
                      <div className="flex items-center gap-2 mb-3">
                        <ShieldAlert className="w-3.5 h-3.5 text-cyan-400" />
                        <h4 className="text-[10px] font-bold text-slate-200 font-mono uppercase tracking-widest">FM 能力值</h4>
                      </div>
                      <div className="grid grid-cols-3 gap-3 text-center">
                        <div>
                          <div className="text-[9px] text-[#666] mb-1">CA</div>
                          <div className="text-lg font-bold text-cyan-400 font-mono">{selectedPlayer.ca || '—'}</div>
                        </div>
                        <div>
                          <div className="text-[9px] text-[#666] mb-1">PA</div>
                          <div className="text-lg font-bold text-slate-300 font-mono">{selectedPlayer.pa || '—'}</div>
                        </div>
                        <div>
                          <div className="text-[9px] text-[#666] mb-1">RCA</div>
                          <div className="text-lg font-bold text-orange-400 font-mono">{selectedPlayer.rca || '—'}</div>
                        </div>
                      </div>
                      {(selectedPlayer.height ?? 0) > 0 && (
                        <div className="text-[10px] text-[#888] font-mono mt-3 text-center">身高 {selectedPlayer.height} cm</div>
                      )}
                    </div>

                    {selectedPlayer.attr_groups && Object.keys(selectedPlayer.attr_groups).length > 0 && (
                      <div className="bg-[#121212] border border-[#2a2a2a] p-4">
                        <div className="flex items-center gap-2 mb-3">
                          <BarChart3 className="w-3.5 h-3.5 text-emerald-400" />
                          <h4 className="text-[10px] font-bold text-slate-200 font-mono uppercase tracking-widest">能力分組</h4>
                        </div>
                        <div className="space-y-2">
                          {FM_GROUP_ORDER.filter((k) => selectedPlayer.attr_groups?.[k]).map((k) => (
                            <div key={k} className="flex items-center justify-between gap-2">
                              <span className="text-[10px] text-[#888]">{k}</span>
                              <div className="flex items-center gap-2 flex-1 max-w-[140px]">
                                <div className="flex-1 h-1 bg-[#2a2a2a] overflow-hidden">
                                  <div
                                    className="h-full bg-emerald-500"
                                    style={{ width: `${Math.min(100, ((selectedPlayer.attr_groups![k] ?? 0) / 20) * 100)}%` }}
                                  />
                                </div>
                                <span className="text-[10px] text-slate-300 w-8 text-right">{selectedPlayer.attr_groups![k].toFixed(1)}</span>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {selectedPlayer.attributes && (
                      <div className="space-y-3">
                        <AttrSection title="技术属性" keys={FM_TECH_ATTRS} attrs={selectedPlayer.attributes} />
                        <AttrSection title="精神属性" keys={FM_MENTAL_ATTRS} attrs={selectedPlayer.attributes} />
                        <AttrSection title="身体属性" keys={FM_PHYS_ATTRS} attrs={selectedPlayer.attributes} />
                        {(selectedPlayer.pos.includes('GK') || FM_GK_ATTRS.some((k) => (selectedPlayer.attributes?.[k] ?? 0) > 0)) && (
                          <AttrSection title="门将属性" keys={FM_GK_ATTRS} attrs={selectedPlayer.attributes} />
                        )}
                      </div>
                    )}
                  </>
                ) : (
                  <>
                    <div className="bg-[#121212] border border-[#2a2a2a] p-4">
                      <div className="flex items-center gap-2 mb-3">
                        <ShieldAlert className="w-3.5 h-3.5 text-cyan-400" />
                        <h4 className="text-[10px] font-bold text-slate-200 font-mono uppercase tracking-widest">FIFA 能力評分</h4>
                      </div>
                      <div className="space-y-3">
                        <div className="flex justify-between items-center text-xs font-mono">
                          <span className="text-[#888]">綜合 (OVR)</span>
                          <span className={cn('font-bold px-1.5 py-0.5', selectedPlayer.prof >= 16 ? 'text-cyan-400 bg-cyan-500/10' : 'text-orange-400 bg-orange-500/10')}>
                            {selectedPlayer.prof}/20
                          </span>
                        </div>
                        <div className="flex justify-between items-center text-xs font-mono">
                          <span className="text-[#888]">潛力 (POT)</span>
                          <span className={cn('font-bold px-1.5 py-0.5', selectedPlayer.imp >= 16 ? 'text-cyan-400 bg-cyan-500/10' : 'text-orange-400 bg-orange-500/10')}>
                            {selectedPlayer.imp}/20
                          </span>
                        </div>
                      </div>
                    </div>

                    {selectedPlayer.role && (
                      <div className="bg-[#121212] border border-[#2a2a2a] p-4">
                        <div className="flex items-center gap-2 mb-2">
                          <Presentation className="w-3.5 h-3.5 text-[#888]" />
                          <h4 className="text-[10px] font-bold text-slate-200 font-mono uppercase tracking-widest">球場角色</h4>
                        </div>
                        <p className="text-[11px] text-slate-300 font-mono leading-relaxed mt-2">{selectedPlayer.role}</p>
                      </div>
                    )}

                    <div className="bg-[#121212] border border-[#2a2a2a] p-4 flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Activity className="w-3.5 h-3.5 text-[#888]" />
                        <h4 className="text-[10px] font-bold text-slate-200 font-mono uppercase tracking-widest">國際聲望</h4>
                      </div>
                      <div className="text-xl font-bold font-mono text-cyan-400">{selectedPlayer.off_pitch}/10</div>
                    </div>
                  </>
                )}
              </div>
            ) : (
              <div className="flex-1 flex flex-col items-center justify-center text-center opacity-50 pb-10">
                <UserCircle2 className="w-12 h-12 text-[#444] mb-4" />
                <div className="text-xs text-[#888] font-mono tracking-widest uppercase mb-1">未選取球員</div>
                <div className="text-[9px] text-[#555] font-mono tracking-widest">點擊左側列表查看大頭照與量化報告</div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
