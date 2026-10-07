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
export type Prices = components['schemas']['WorldPrices'];
export function planCost(p: Plan, prices: Prices) { return p.upgrades.length * prices.upgrade + p.shields.length * prices.shield + p.bombs * prices.bomb + Number(p.nuclear) * prices.nuclear + Number(p.ecology) * prices.ecology + p.donations.reduce((sum, d) => sum + d.amountCents, 0); }
export const cityRoles: Record<string, { name: string; benefit: string; color: string }> = {
  capital: { name: 'Столица', benefit: '1 уровень → дополнительная принятая встреча', color: 'violet' },
  military: { name: 'Военный город', benefit: 'Каждый уровень: бомбы дешевле на 50', color: 'rose' },
  security: { name: 'Город безопасности', benefit: 'Каждый уровень: куполы дешевле на 50', color: 'blue' },
  intelligence: { name: 'Город разведки', benefit: '1 уровень → тайный отчёт о другой стране', color: 'amber' },
  tourism: { name: 'Туристический город', benefit: 'Каждый уровень: +75 монет дохода', color: 'green' },
};
