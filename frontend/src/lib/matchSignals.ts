import type { MatchResponse, MatchSignal } from '../types';

/** Match window used when CDN finished flag is stale (~105 min + buffer). */
const MATCH_MS = (2 * 60 + 20) * 60 * 1000;

export function signalsByMatchId(signals: MatchSignal[] | null): Map<string, MatchSignal> {
  const map = new Map<string, MatchSignal>();
  if (!signals) return map;
  for (const s of signals) {
    map.set(s.id, s);
  }
  return map;
}

export type MatchFilter = 'active' | 'live' | 'upcoming' | 'finished' | 'all';

/** Align with backend ResolveMatchStatus when API hasn't been rebuilt yet. */
export function resolveMatchStatus(match: MatchResponse, now = Date.now()): string {
  if (match.finished || match.status === 'finished') return 'finished';
  if (match.status === 'live') return 'live';

  const te = (match.time_elapsed ?? '').toLowerCase().trim();
  if (te && te !== 'notstarted') {
    if (te.includes("'") || te.includes('half') || te === 'live') return 'live';
    if (te.includes('ft') || te.includes('full')) return 'finished';
    return match.status || te;
  }

  const kickoff = new Date(match.kickoff).getTime();
  if (Number.isNaN(kickoff)) return match.status || 'scheduled';

  if (now > kickoff + MATCH_MS) return 'finished';
  // CDN still notstarted — don't infer live from kickoff alone.
  if (te === 'notstarted' || te === '') {
    return 'scheduled';
  }
  if (now >= kickoff) return 'live';
  return 'scheduled';
}

export function isMatchFinished(match: MatchResponse, now = Date.now()): boolean {
  return resolveMatchStatus(match, now) === 'finished';
}

export function classifyMatch(match: MatchResponse, now = Date.now()): MatchFilter {
  const status = resolveMatchStatus(match, now);
  if (status === 'live') return 'live';
  if (status === 'finished') return 'finished';
  return 'upcoming';
}

export function filterMatches(
  matches: MatchResponse[],
  filter: MatchFilter,
  now = Date.now(),
): MatchResponse[] {
  if (filter === 'all') return matches;
  if (filter === 'active') {
    return matches.filter((m) => classifyMatch(m, now) !== 'finished');
  }
  return matches.filter((m) => classifyMatch(m, now) === filter);
}

export function sortMatchesForDisplay(matches: MatchResponse[], now = Date.now()): MatchResponse[] {
  const live: MatchResponse[] = [];
  const upcoming: MatchResponse[] = [];
  const finished: MatchResponse[] = [];

  for (const m of matches) {
    const bucket = classifyMatch(m, now);
    if (bucket === 'live') live.push(m);
    else if (bucket === 'finished') finished.push(m);
    else upcoming.push(m);
  }

  const byKickoff = (a: MatchResponse, b: MatchResponse) =>
    new Date(a.kickoff).getTime() - new Date(b.kickoff).getTime();

  live.sort(byKickoff);
  upcoming.sort(byKickoff);
  finished.sort((a, b) => byKickoff(b, a));

  return [...live, ...upcoming, ...finished];
}

export function hasLiveMatches(matches: MatchResponse[] | null, now = Date.now()): boolean {
  return (matches ?? []).some((m) => resolveMatchStatus(m, now) === 'live');
}

export function matchCounts(matches: MatchResponse[], now = Date.now()) {
  let live = 0;
  let upcoming = 0;
  let finished = 0;
  for (const m of matches) {
    const bucket = classifyMatch(m, now);
    if (bucket === 'live') live++;
    else if (bucket === 'finished') finished++;
    else upcoming++;
  }
  return {
    all: matches.length,
    active: live + upcoming,
    live,
    upcoming,
    finished,
  };
}
