'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  getExecutions, getContext, getContextGraph, getSessions,
  getSessionReplay, ExecutionNode, ContextState, Session, Edge,
} from '@/lib/api';

function useFetch<T>(fetcher: () => Promise<T>, deps: unknown[] = []) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refetch = useCallback(async () => {
    setLoading(true);
    try {
      const result = await fetcher();
      setData(result);
      setError(null);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  }, deps);

  useEffect(() => {
    refetch();
  }, [refetch]);

  return { data, loading, error, refetch };
}

export function useExecutions(params?: { limit?: number; sessionId?: string; search?: string }) {
  return useFetch(
    () => getExecutions({
      limit: params?.limit || 50,
      session_id: params?.sessionId,
    }),
    [params?.limit, params?.sessionId, params?.search]
  );
}

export function useContext() {
  return useFetch(() => getContext(), []);
}

export function useGraph(sessionId?: string) {
  return useFetch(() => getContextGraph(sessionId), [sessionId]);
}

export function useSessions() {
  return useFetch(() => getSessions(), []);
}

export function useSession(id: string) {
  return useFetch(() => getSessionReplay(id), [id]);
}
