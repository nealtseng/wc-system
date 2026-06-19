import { useState, useCallback } from 'react';
import { fetchPrediction } from '../services/api';
import type { PredictResponse } from '../types';

interface UsePredictionState {
  data: PredictResponse | null;
  loading: boolean;
  error: string | null;
}

interface UsePredictionReturn extends UsePredictionState {
  predict: (homeId: string, awayId: string) => Promise<void>;
  reset: () => void;
}

export function usePrediction(): UsePredictionReturn {
  const [state, setState] = useState<UsePredictionState>({
    data: null,
    loading: false,
    error: null,
  });

  const predict = useCallback(async (homeId: string, awayId: string) => {
    setState({ data: null, loading: true, error: null });
    try {
      const data = await fetchPrediction(homeId, awayId);
      setState({ data, loading: false, error: null });
    } catch (err) {
      setState({ data: null, loading: false, error: (err as Error).message });
    }
  }, []);

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null });
  }, []);

  return { ...state, predict, reset };
}
