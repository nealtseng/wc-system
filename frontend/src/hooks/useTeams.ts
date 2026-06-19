import { useState, useEffect } from 'react';
import { fetchTeams } from '../services/api';
import type { TeamResponse } from '../types';

interface UseTeamsState {
  data: TeamResponse[] | null;
  loading: boolean;
  error: string | null;
}

export function useTeams(): UseTeamsState {
  const [state, setState] = useState<UseTeamsState>({
    data: null,
    loading: true,
    error: null,
  });

  useEffect(() => {
    let cancelled = false;

    fetchTeams()
      .then((data) => {
        if (!cancelled) setState({ data, loading: false, error: null });
      })
      .catch((err) => {
        if (!cancelled) setState({ data: null, loading: false, error: (err as Error).message });
      });

    return () => {
      cancelled = true;
    };
  }, []);

  return state;
}
