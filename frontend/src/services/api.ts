import type {
  AdminStatus,
  MatchResponse,
  MatchSignal,
  MonteCarloResponse,
  NarrativeRequest,
  NarrativeResponse,
  PipelineStatus,
  PredictResponse,
  SquadResponse,
  TeamResponse,
} from '../types';

// BASE resolves to '' (empty string) so that Vite's proxy forwards /api/*
// to http://localhost:8080 during development.  In production builds where
// VITE_API_BASE_URL is set, requests go directly to the backend host.
const BASE = (import.meta.env.VITE_API_BASE_URL as string) ?? '';

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`API ${path} returned ${res.status}: ${text}`);
  }
  return res.json() as Promise<T>;
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`);
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`API ${path} returned ${res.status}: ${text}`);
  }
  return res.json() as Promise<T>;
}

export function fetchPrediction(homeId: string, awayId: string): Promise<PredictResponse> {
  return post<PredictResponse>('/api/predict', {
    home_team_id: homeId,
    away_team_id: awayId,
  });
}

export function fetchMonteCarlo(
  homeId: string,
  awayId: string,
  iterations = 100000,
): Promise<MonteCarloResponse> {
  return post<MonteCarloResponse>('/api/predict/monte-carlo', {
    home_team_id: homeId,
    away_team_id: awayId,
    iterations,
  });
}

export function fetchNarrative(input: NarrativeRequest): Promise<NarrativeResponse> {
  return post<NarrativeResponse>('/api/narrative', input);
}

export function fetchTeams(): Promise<TeamResponse[]> {
  return get<TeamResponse[]>('/api/teams');
}

export function fetchPipelineStatus(): Promise<PipelineStatus> {
  return get<PipelineStatus>('/api/pipeline/status');
}

export function syncPipeline(): Promise<{ status: string; teams: number; message?: string }> {
  return post('/api/pipeline/sync', {});
}

export function syncFBref(): Promise<{
  status: string;
  teams_ok: number;
  teams_total: number;
  teams_fail: number;
  message?: string;
  error?: string;
}> {
  return post('/api/pipeline/sync/fbref', {});
}

export function syncOdds(): Promise<{ status: string; message?: string }> {
  return post('/api/pipeline/sync/odds', {});
}

export function fetchPipelineTargets(): Promise<Record<string, string>> {
  return get<Record<string, string>>('/api/pipeline/targets');
}

export function fetchWikiThumbnail(slug: string): Promise<{ thumbnail_url: string; page_url?: string }> {
  return get(`/api/wiki/thumbnail/${encodeURIComponent(slug)}`);
}

export function fetchSquad(teamId: string): Promise<SquadResponse> {
  return get<SquadResponse>(`/api/teams/${encodeURIComponent(teamId)}/squad`);
}

export function fetchMatches(): Promise<MatchResponse[]> {
  return get<MatchResponse[]>('/api/matches');
}

export function fetchSignals(): Promise<MatchSignal[]> {
  return get<MatchSignal[]>('/api/signals');
}

export function fetchAdminStatus(): Promise<AdminStatus> {
  return get<AdminStatus>('/api/admin/status');
}

export async function triggerCalibration(): Promise<{
  w1: number;
  w2: number;
  w3: number;
  brier_score: number;
  matches_used: number;
}> {
  const res = await fetch(`${BASE}/api/pipeline/calibrate`, { method: 'POST' });
  if (!res.ok) throw new Error(`calibrate ${res.status}`);
  return res.json();
}
