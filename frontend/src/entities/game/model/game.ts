import type { components } from '@/shared/api';

export type Game = components['schemas']['Game'];

export function isGame(value: unknown): value is Game {
  if (typeof value !== 'object' || value === null) return false;
  const game = value as Record<string, unknown>;
  return ['id', 'title', 'description', 'category'].every(key => typeof game[key] === 'string' && game[key] !== '')
    && typeof game.status === 'string' && ['coming_soon', 'available', 'maintenance'].includes(game.status)
    && typeof game.accent === 'string' && ['lime', 'lavender', 'peach'].includes(game.accent);
}
