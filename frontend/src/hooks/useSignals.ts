import { useState, useEffect, useCallback } from 'react';
import { fetchSignals } from '../services/api';
import type { MatchSignal } from '../types';

export function useSignals() {
  const [data, setData] = useState<MatchSignal[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      const signals = await fetchSignals();
      setData(signals);
      setError(null);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    reload();
    const id = setInterval(reload, 60_000);
    return () => clearInterval(id);
  }, [reload]);

  return { data, loading, error, reload };
}
