import { baseApi } from '@/shared/api';
import { isGame, type Game } from '../model/game';

const gameApi = baseApi.enhanceEndpoints({ addTagTypes: ['Game'] }).injectEndpoints({
  endpoints: build => ({
    getGames: build.query<Game[], void>({
      async queryFn(_arg, _api, _extra, fetchWithBQ) {
        const result = await fetchWithBQ('games');
        if (result.error) return { error: result.error };
        const payload = result.data;
        if (typeof payload !== 'object' || payload === null || !('games' in payload)
          || !Array.isArray(payload.games) || !payload.games.every(isGame)) {
          return { error: { status: 'CUSTOM_ERROR', error: 'Invalid catalog response' } };
        }
        return { data: payload.games };
      },
      providesTags: [{ type: 'Game', id: 'LIST' }],
    }),
  }),
});

export const { useGetGamesQuery } = gameApi;
