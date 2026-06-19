import { useState, useCallback } from 'react';
import { fetchNarrative } from '../services/api';
import type { NarrativeRequest, NarrativeResponse } from '../types';

interface UseNarrativeState {
  data: NarrativeResponse | null;
  loading: boolean;
  error: string | null;
}

interface UseNarrativeReturn extends UseNarrativeState {
  generate: (input: NarrativeRequest) => Promise<void>;
  reset: () => void;
}

export function useNarrative(): UseNarrativeReturn {
  const [state, setState] = useState<UseNarrativeState>({
    data: null,
    loading: false,
    error: null,
  });

  const generate = useCallback(async (input: NarrativeRequest) => {
    setState({ data: null, loading: true, error: null });
    try {
      const data = await fetchNarrative(input);
      setState({ data, loading: false, error: null });
    } catch (err) {
      setState({ data: null, loading: false, error: (err as Error).message });
    }
  }, []);

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null });
  }, []);

  return { ...state, generate, reset };
}
