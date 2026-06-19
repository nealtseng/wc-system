import { useState, useEffect, useCallback, useRef } from 'react';
import { fetchMatches } from '../services/api';
import { hasLiveMatches } from '../lib/matchSignals';
import type { MatchResponse } from '../types';

const LIVE_POLL_MS = 30_000;
const IDLE_POLL_MS = 120_000;

interface UseMatchesOptions {
  pollWhenLive?: boolean;
}

interface UseMatchesState {
  data: MatchResponse[] | null;
  loading: boolean;
  error: string | null;
  reload: () => Promise<void>;
  lastUpdated: Date | null;
}

export function useMatches(options?: UseMatchesOptions): UseMatchesState {
  const pollWhenLive = options?.pollWhenLive ?? false;
  const [data, setData] = useState<MatchResponse[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const dataRef = useRef<MatchResponse[] | null>(null);

  const reload = useCallback(async () => {
    try {
      const next = await fetchMatches();
      dataRef.current = next;
      setData(next);
      setError(null);
      setLastUpdated(new Date());
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    reload();
  }, [reload]);

  useEffect(() => {
    if (!pollWhenLive) return;

    const tick = () => {
      reload();
    };

    const interval = hasLiveMatches(dataRef.current) ? LIVE_POLL_MS : IDLE_POLL_MS;
    const id = setInterval(tick, interval);
    return () => clearInterval(id);
  }, [pollWhenLive, reload, data]);

  return { data, loading, error, reload, lastUpdated };
}

export { hasLiveMatches };
