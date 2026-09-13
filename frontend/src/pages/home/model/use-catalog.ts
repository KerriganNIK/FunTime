import { useEffect, useState } from 'react';
import { getGames, type Game } from '@/entities/game';

type CatalogState = { status: 'loading' } | { status: 'error' } | { status: 'ready'; games: Game[] };

export function useCatalog() {
  const [state, setState] = useState<CatalogState>({ status: 'loading' });
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    getGames(controller.signal)
      .then(games => { if (!controller.signal.aborted) setState({ status: 'ready', games }); })
      .catch(() => { if (!controller.signal.aborted) setState({ status: 'error' }); });
    return () => controller.abort();
  }, [attempt]);
  return { state, retry: () => { setState({ status: 'loading' }); setAttempt(value => value + 1); } };
}
