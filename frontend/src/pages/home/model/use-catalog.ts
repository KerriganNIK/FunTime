import { useGetGamesQuery, type Game } from '@/entities/game';

type CatalogState = { status: 'loading' } | { status: 'error' } | { status: 'ready'; games: Game[] };

export function useCatalog() {
  const { data, isFetching, isError, refetch } = useGetGamesQuery();
  const state: CatalogState = isFetching && !data ? { status: 'loading' }
    : isError ? { status: 'error' } : data ? { status: 'ready', games: data } : { status: 'loading' };
  return { state, retry: () => { void refetch(); } };
}
