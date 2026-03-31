import { useState, useEffect } from 'react';

export function usePolling<T>(fetcher: () => Promise<T>, intervalMs: number) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;

    const poll = async () => {
      try {
        const result = await fetcher();
        if (active) {
          setData(result);
          setError(null);
        }
      } catch (e) {
        if (active) setError(String(e));
      }
    };

    poll();
    const id = setInterval(poll, intervalMs);
    return () => { active = false; clearInterval(id); };
  }, [fetcher, intervalMs]);

  return { data, error };
}
