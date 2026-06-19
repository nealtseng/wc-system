import type { MatchResponse } from '../types';
import { isMatchFinished, resolveMatchStatus } from './matchSignals';

export type MilestoneStatus = 'completed' | 'active' | 'upcoming';

export interface TournamentMilestone {
  id: string;
  title: string;
  subtitle: string;
  status: MilestoneStatus;
  date: string;
  finished: number;
  total: number;
}

const STAGE_DEFS: Array<{ id: string; title: string; subtitle: string; stages: string[] }> = [
  { id: 'group', title: '小組賽', subtitle: 'Group Stage', stages: ['group'] },
  { id: 'r32', title: '三十二強', subtitle: 'Round of 32', stages: ['r32'] },
  { id: 'r16', title: '十六強', subtitle: 'Round of 16', stages: ['r16'] },
  { id: 'qf', title: '八強', subtitle: 'Quarter-finals', stages: ['qf'] },
  { id: 'sf', title: '四強', subtitle: 'Semi-finals', stages: ['sf'] },
  { id: 'final', title: '決賽', subtitle: 'Final', stages: ['final'] },
];

function stageDateRange(matches: MatchResponse[]): string {
  if (matches.length === 0) return '—';
  const dates = matches
    .map((m) => m.local_date?.split(' ')[0])
    .filter(Boolean)
    .sort();
  if (dates.length === 0) return '—';
  const first = dates[0];
  const last = dates[dates.length - 1];
  const fmt = (d: string) => {
    const [mm, dd] = d.split('/');
    return mm && dd ? `${mm}/${dd}` : d;
  };
  return first === last ? fmt(first) : `${fmt(first)} – ${fmt(last)}`;
}

export function buildTournamentMilestones(matches: MatchResponse[]): TournamentMilestone[] {
  const now = Date.now();
  let activeAssigned = false;

  const milestones = STAGE_DEFS.map((def) => {
    const stageMatches = matches.filter((m) => def.stages.includes(m.stage));
    const total = stageMatches.length;
    const finished = stageMatches.filter((m) => isMatchFinished(m)).length;
    const hasLive = stageMatches.some((m) => resolveMatchStatus(m) === 'live');
    const hasUpcoming = stageMatches.some((m) => resolveMatchStatus(m) === 'scheduled');
    const hasStarted = stageMatches.some(
      (m) => isMatchFinished(m) || resolveMatchStatus(m) === 'live' || new Date(m.kickoff).getTime() <= now,
    );

    let status: MilestoneStatus = 'upcoming';
    if (total > 0 && finished === total) {
      status = 'completed';
    } else if (total > 0 && (hasLive || hasStarted || finished > 0)) {
      status = 'active';
    } else if (total > 0 && hasUpcoming && !activeAssigned) {
      status = 'active';
    }

    if (status === 'active') activeAssigned = true;

    return {
      id: def.id,
      title: def.title,
      subtitle: def.subtitle,
      status,
      date: stageDateRange(stageMatches),
      finished,
      total,
    };
  });

  // If nothing marked active, pick first non-completed stage with matches.
  if (!milestones.some((m) => m.status === 'active')) {
    const next = milestones.find((m) => m.status === 'upcoming' && m.total > 0);
    if (next) next.status = 'active';
  }

  return milestones;
}

export function tournamentProgressPercent(matches: MatchResponse[]): number {
  if (matches.length === 0) return 0;
  const finished = matches.filter((m) => isMatchFinished(m)).length;
  return Math.round((finished / matches.length) * 100);
}
