export type View = 'global-dashboard' | 'upcoming-matches' | 'prediction-model' | 'data-operations' | 'admin-panel';

export interface MatchSignal {
  id: string;
  homeTeam: string;
  awayTeam: string;
  homeFlag: string;
  awayFlag: string;
  kickoff: string;
  bookmarkOdds: {
    home: number;
    draw: number;
    away: number;
  };
  impliedProb: {
    home: number;
    draw: number;
    away: number;
  };
  pFinal: {
    home: number;
    draw: number;
    away: number;
  };
  ev: {
    home: number;
    draw: number;
    away: number;
  };
  weights: {
    w1: number; // Insider
    w2: number; // Poisson
    w3: number; // Macro
    clip: number;
  };
  kelly_fraction: {
    home: number;
    draw: number;
    away: number;
  };
}

export interface TeamInfo {
  id: string;
  name: string;
  iso2: string;
  group: string;
  elo: number;
  gdp: string;
  gdpValue: number;
}

// ---------- Backend API response types ----------

export interface ScoreProbability {
  home: number;
  away: number;
  prob: number;
}

export interface PredictResponse {
  home_team_id: string;
  away_team_id: string;
  neutral_venue?: boolean;
  elo: { home: number; away: number };
  poisson: {
    home_lambda: number;
    away_lambda: number;
    home_win: number;
    draw: number;
    away_win: number;
    top_scores: ScoreProbability[];
  };
  /** W2-only baseline (xG/ELO λ, no W1/W3). Differs from p_final when other factors pull. */
  poisson_w2?: { home_win: number; draw: number; away_win: number };
  weights: { w1: number; w2: number; w3: number; clip: number };
  p_final: { home: number; draw: number; away: number };
  w3_implied?: { home: number; draw: number; away: number };
  w3_odds?: { home: number; draw: number; away: number };
  w3_totals?: { line: number; over: number; under: number };
  blend?: {
    w2: { home: number; draw: number; away: number };
    w3: { home: number; draw: number; away: number };
    w1_delta: number;
  };
  signals?: {
    poisson_favors: 'home' | 'away' | 'even';
    final_favors: 'home' | 'away' | 'even';
    w2_lambda_favors?: 'home' | 'away' | 'even';
    w3_lambda_favors?: 'home' | 'away' | 'even';
    conflict: boolean;
  };
  sources?: {
    elo: string;
    xg: string;
    gdp?: string;
    wiki?: string;
    w1_delta: number;
    w3: 'odds' | 'elo_fallback';
  };
  lambda_layers?: {
    w2: { home: number; away: number };
    w1: { home: number; away: number };
    w3: { home: number; away: number; implied_total: number; total_source?: 'totals' | 'draw_proxy' };
  };
  lambda_blend?: { home: number; away: number };
  lambda_contrib?: {
    w2: { home: number; away: number };
    w1: { home: number; away: number };
    w3: { home: number; away: number };
  };
}

export interface NarrativeRequest {
  home_team: string;
  away_team: string;
  home_win_prob: number;
  draw_prob: number;
  away_win_prob: number;
  home_elo: number;
  away_elo: number;
  home_gdp: number;
  away_gdp: number;
  home_lambda: number;
  away_lambda: number;
  w1: number;
  w2: number;
  w3: number;
  venue_label?: string;
  poisson_favors?: string;
  final_favors?: string;
  signal_conflict?: boolean;
}

export interface NarrativeResponse {
  narrative: string;
  confidence: number;
}

export interface TeamResponse {
  id: string;
  name: string;
  elo: number;
  gdp_per_capita: number;
  gdp_year?: number;
  wiki_extract?: string;
  wiki_url?: string;
  avg_goals_for: number;
  avg_xg_for?: number;
  avg_xg_source?: string;
  avg_xg_match_count?: number;
  win_rate: number;
  momentum: number;
  player_strength?: number;
  matches_played: number;
  narrative_weight: number;
  wc_group: string;
  iso2: string;
}

export interface MatchResponse {
  id: number;
  wc_match_id: string;
  home_id: string;
  away_id: string;
  home_iso2: string;
  away_iso2: string;
  home_name: string;
  away_name: string;
  stadium_name: string;
  stage: string;
  matchday: number;
  local_date: string;
  kickoff: string;
  home_score: number | null;
  away_score: number | null;
  finished: boolean;
  time_elapsed: string;
  status: string;
}

export interface PipelineScraperStatus {
  name: string;
  status: 'ok' | 'degraded' | 'offline';
  last_fetch: string;
  message?: string;
}

export interface PipelineSchedule {
  label: string;
  next_at: string;
  endpoint: string;
  match_label?: string;
  window_key?: string;
  wc_match_id?: string;
  synced?: boolean;
}

export interface PipelineStatus {
  scrapers: PipelineScraperStatus[];
  schedules: PipelineSchedule[];
  updated_at: string;
}

export interface SquadPlayerResponse {
  id: string;
  no: number;
  name: string;
  wiki_slug: string;
  pos: string;
  age: number;
  fitness: number;
  apps: number;
  goals: number;
  value: string;
  prof: number;
  imp: number;
  role: string;
  off_pitch: number;
  image_url?: string;
  source?: string;
  ca?: number;
  pa?: number;
  rca?: number;
  club?: string;
  height?: number;
  attributes?: Record<string, number>;
  attr_groups?: Record<string, number>;
}

export interface SquadResponse {
  team_id: string;
  source: string;
  players: SquadPlayerResponse[];
}

export interface AdminStatus {
  teams_count: number;
  matches_count: number;
  odds_count: number;
  signals_count: number;
  scrapers_ok: number;
  scrapers_total: number;
  scrapers_degraded: number;
  last_sync: string;
  system_healthy: boolean;
  scrapers: PipelineScraperStatus[];
}

export interface MonteCarloResponse extends PredictResponse {
  iterations: number;
}
