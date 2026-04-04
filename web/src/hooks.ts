import { useState, useEffect, useRef } from 'react';

// Global SSE connection — shared across all components.
// Increments a version counter on each server event, which triggers refetches.
let globalVersion = 0;
const listeners = new Set<(v: number) => void>();
let eventSource: EventSource | null = null;

function ensureEventSource() {
  if (eventSource) return;

  const es = new EventSource('/api/events');

  es.addEventListener('obs_st', () => bump());
  es.addEventListener('rapid_wind', () => bump());
  es.addEventListener('evt_precip', () => bump());
  es.addEventListener('evt_strike', () => bump());

  // On reconnect, bump to refetch anything we missed
  es.addEventListener('connected', () => bump());

  es.onerror = () => {
    // EventSource auto-reconnects; the 'connected' event will trigger a refetch
  };

  eventSource = es;
}

function bump() {
  globalVersion++;
  for (const fn of listeners) {
    fn(globalVersion);
  }
}

export function useServerEvents(): number {
  const [version, setVersion] = useState(globalVersion);

  useEffect(() => {
    ensureEventSource();
    listeners.add(setVersion);
    return () => { listeners.delete(setVersion); };
  }, []);

  return version;
}

// Fetches data once on mount and again whenever `version` changes.
export function useFetch<T>(fetcher: () => Promise<T>, version: number) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const fetcherRef = useRef(fetcher);
  fetcherRef.current = fetcher;

  useEffect(() => {
    let active = true;

    fetcherRef.current()
      .then(result => { if (active) { setData(result); setError(null); } })
      .catch(e => { if (active) setError(String(e)); });

    return () => { active = false; };
  }, [version]);

  // Also refetch when fetcher identity changes (e.g. range switch)
  useEffect(() => {
    let active = true;

    fetcher()
      .then(result => { if (active) { setData(result); setError(null); } })
      .catch(e => { if (active) setError(String(e)); });

    return () => { active = false; };
  }, [fetcher]);

  return { data, error };
}
