import type { components } from '@/shared/api';

export type Plan = components['schemas']['WorldPlan'];
export type City = components['schemas']['WorldCity'];
export type Country = components['schemas']['WorldCountry'];
export type GameEvent = components['schemas']['WorldEvent'];
export type Meeting = components['schemas']['WorldMeeting'];
export type WorldMatch = components['schemas']['WorldMatch'];
export function isWorldMatch(value: unknown): value is WorldMatch { return typeof value === 'object' && value !== null && 'phase' in value && 'countries' in value && Array.isArray(value.countries); }
export const roles: Record<string, string> = { president: 'Президент', economy: 'Министр экономики', defense: 'Министр обороны', foreign: 'Министр иностранных дел', statistics: 'Министр статистики' };
export const coins = (cents: number) => new Intl.NumberFormat('ru', { maximumFractionDigits: 2 }).format(cents / 100);
export function planCost(p: Plan) { return p.upgrades.length * 15000 + (p.shields.length + p.bombs) * 30000 + Number(p.nuclear) * 50000 + Number(p.ecology) * 20000 + p.donations.reduce((sum, d) => sum + d.amountCents, 0); }
