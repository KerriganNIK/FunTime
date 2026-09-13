import { getJSON } from '@/shared/api';
import { isGame, type Game } from '../model/game';

export async function getGames(signal?: AbortSignal): Promise<Game[]> {
  const payload = await getJSON('/api/v1/games', signal);
  if (typeof payload !== 'object' || payload === null || !('games' in payload)
    || !Array.isArray(payload.games) || !payload.games.every(isGame)) {
    throw new Error('Invalid catalog response');
  }
  return payload.games;
}
