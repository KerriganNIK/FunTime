import { test, expect, type APIRequestContext } from '@playwright/test';
import { selectOrderSection } from '../helpers/order-section';

test('a congress runs through six rounds with shared plans, secrets and final attack history', async ({ page: host, browser }, testInfo) => {
  test.setTimeout(120_000);
  const guestContext = await browser.newContext({ baseURL: 'http://127.0.0.1:5173', viewport: { width: 390, height: 844 } });
  const teammateContext = await browser.newContext({ baseURL: 'http://127.0.0.1:5173' });
  const guest = await guestContext.newPage();
  const errors: string[] = [];
  host.on('pageerror', error => errors.push(error.message)); guest.on('pageerror', error => errors.push(error.message));
  try {
    await host.goto('/');
    await host.getByRole('button', { name: 'Об игре', exact: true }).click();
    await host.getByRole('dialog').getByLabel('Ваше имя').fill('Ведущий');
    await host.getByRole('button', { name: 'Создать комнату', exact: true }).click();
    await expect(host).toHaveURL(/\/rooms\/[A-Z2-9]{6}$/);
    const code = host.url().split('/').at(-1)!;
    const path = `/api/v1/rooms/${code}`;
    const snapshot = async (request: APIRequestContext, screen = false) => { const response = await request.get(`${path}${screen ? '/screen' : ''}`); expect(response.ok()).toBeTruthy(); return response.json(); };
    const command = async (request: APIRequestContext, kind: string, payload: unknown, requestId = crypto.randomUUID()) => request.post(`${path}/commands`, { data: { kind, payload, requestId } });
    const advance = async () => { const r = await snapshot(host.request); const response = await command(host.request, 'advance', { expectedPhase: `${r.match.round}:${r.match.phase}` }); expect(response.ok(), await response.text()).toBeTruthy(); };

    await host.getByRole('combobox', { name: 'Страна', exact: true }).selectOption('norway');
    await host.getByRole('button', { name: 'Занять место' }).click();
    await guest.goto(`/rooms/${code}`);
    await guest.getByLabel('Ваше имя').fill('Делегат');
    await guest.getByRole('button', { name: 'Присоединиться', exact: true }).click();
    await guest.getByRole('combobox', { name: 'Страна', exact: true }).selectOption('germany');
    await guest.getByRole('button', { name: 'Занять место' }).click();
    await expect(guest.getByText('Вы в команде «Германия», Президент.')).toBeVisible();

    const joined = await teammateContext.request.post(`${path}/join`, { data: { name: 'Экономист' } }); expect(joined.ok()).toBeTruthy();
    expect((await command(teammateContext.request, 'select', { countryId: 'germany', role: 'president' })).status()).toBe(409);
    expect((await command(teammateContext.request, 'select', { countryId: 'germany', role: 'economy' })).ok()).toBeTruthy();
    expect((await command(guest.request, 'start', { phaseSeconds: 720 })).status()).toBe(403);
    await host.getByRole('button', { name: 'Начать игру', exact: true }).click();
    await expect(host.getByRole('heading', { name: 'Заседание ООН', exact: true })).toBeVisible();
    await host.getByRole('button', { name: 'Открыть штабы', exact: true }).click();
    await expect(guest.getByRole('heading', { name: 'Работа в штабе', exact: true })).toBeVisible();

    const first = await snapshot(guest.request);
    expect(first.match.rules.version).toBe(2);
    expect(first.match.countries.find((c: { id: string }) => c.id === 'germany').balanceCents).toBe(70000);
    expect(first.match.countries.find((c: { id: string }) => c.id === 'germany').cities).toHaveLength(5);
    const hidden = first.match.countries.find((c: { id: string }) => c.id === 'norway');
    expect(hidden).not.toHaveProperty('bombs'); expect(hidden).not.toHaveProperty('plan'); expect(hidden.cities[0]).not.toHaveProperty('shield');
    expect(hidden.cities[0]).not.toHaveProperty('level'); expect(hidden).not.toHaveProperty('intelligence');
    const publicView = await snapshot(host.request, true);
    expect(publicView).not.toHaveProperty('me'); expect(publicView.match).not.toHaveProperty('formulas');
    for (const country of publicView.match.countries) { expect(country).not.toHaveProperty('balanceCents'); expect(country).not.toHaveProperty('plan'); }

    await selectOrderSection(host.getByRole('region', { name: 'Норвегия', exact: true }), 'Оборона');
    await host.getByLabel('Запустить программу').check();
    await selectOrderSection(host.getByRole('region', { name: 'Норвегия', exact: true }), 'Экономика');
    await host.getByLabel('Пожертвовать: Германия').fill('50');
    await host.getByRole('button', { name: 'Сохранить план', exact: true }).click();
    await expect(host.getByText('План сохранён для всей команды')).toBeVisible();
    const stale = first.match.countries.find((c: { id: string }) => c.id === 'germany').plan;
    expect((await command(teammateContext.request, 'plan', { ...stale, upgrades: ['germany-1','germany-4'], spies:['norway'] })).ok()).toBeTruthy();
    expect((await command(guest.request, 'plan', stale)).status()).toBe(409);
    await expect(guest.getByLabel('Улучшить', { exact: false }).first()).toBeChecked();

    expect((await command(guest.request, 'meeting.request', { countryId: 'norway' })).ok()).toBeTruthy();
    const meetings = (await snapshot(host.request)).match.meetings;
    expect((await command(host.request, 'meeting.respond', { id: meetings[0].id, accept: true })).ok()).toBeTruthy();
    expect((await command(guest.request, 'meeting.message', { id: meetings[0].id, text: 'Предлагаем мир.' })).ok()).toBeTruthy();
    expect((await snapshot(host.request, true)).match.meetings).toEqual([]);
    await advance();
    const second = await snapshot(guest.request);
    const germanSecond = second.match.countries.find((c: { id: string }) => c.id === 'germany');
    expect(germanSecond.balanceCents).toBe(129000);
    expect(second.match.pollution).toBe(6);
    expect(germanSecond.intelligence).toHaveLength(1);
    expect(germanSecond.intelligence[0]).toMatchObject({round:1,countryId:'norway',balanceCents:109000,bombs:0,nuclearRound:2});
    await expect(guest.getByRole('heading', { name: 'Закрытые отчёты' })).toBeVisible();
    const publicSecond = (await snapshot(host.request,true)).match;
    for (const country of publicSecond.countries) { expect(country).not.toHaveProperty('intelligence'); for(const city of country.cities) expect(city).not.toHaveProperty('level'); }
    expect(publicSecond.events.filter((e: {kind:string}) => e.kind === 'spy')).toEqual([]);
    expect((await command(guest.request,'income.adjust',{baseIncomeCents:120000,expectedPhase:'2:council'})).status()).toBe(403);
    expect((await command(host.request,'income.adjust',{baseIncomeCents:120000,expectedPhase:'2:council'})).ok()).toBeTruthy();
    expect((await snapshot(host.request,true)).match.baseIncomeCents).toBe(120000);
    expect((await command(host.request,'income.adjust',{baseIncomeCents:100000,expectedPhase:'2:council'})).ok()).toBeTruthy();
    const donation = second.match.events.find((e: { kind: string }) => e.kind === 'donation');
    expect(donation.amount).toBe(5000); expect(donation).not.toHaveProperty('countryId');

    await advance();
    const now = await snapshot(host.request);
    const own = now.match.countries.find((c: { id: string }) => c.id === 'norway');
    expect((await command(host.request, 'plan', { ...own.plan, bombs: 2, launches: ['germany-1', 'germany-1'] })).ok()).toBeTruthy();
    const their = (await snapshot(guest.request)).match.countries.find((c: { id: string }) => c.id === 'germany');
    expect((await command(guest.request, 'plan', { ...their.plan, shields: ['germany-1'] })).ok()).toBeTruthy();
    await expect(guest.locator('.game-status-strip')).toContainText('2 / 6');
    await expect(guest.getByRole('checkbox', { name: /Щит/ }).first()).toBeChecked();
    await host.screenshot({ path: testInfo.outputPath('host-headquarters.png'), fullPage: true });
    await guest.screenshot({ path: testInfo.outputPath('mobile-headquarters.png'), fullPage: true });
    await guest.getByRole('button', { name: 'Включить тёмную тему' }).click();
    await guest.screenshot({ path: testInfo.outputPath('mobile-dark.png'), fullPage: true });
    await guest.getByRole('button', { name: 'Включить светлую тему' }).click();
    for (const width of [320, 390, 768]) { await guest.setViewportSize({ width, height: 900 }); expect(await guest.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `game width ${width}`).toBeTruthy(); }
    await advance();
    const attacked = (await snapshot(guest.request)).match;
    expect(attacked.countries.find((c: { id: string }) => c.id === 'germany').cities[0].destroyed).toBe(true);
    expect(attacked.events.filter((e: { kind: string }) => e.kind === 'attack')).toHaveLength(0);

    await guest.reload();
    await expect(guest.getByRole('heading', { name: 'Заседание ООН', exact: true })).toBeVisible();
    while ((await snapshot(host.request)).match.phase !== 'finished') await advance();
    await expect(guest.getByRole('heading', { name: 'Кто изменил мир', exact: true })).toBeVisible();
    const final = (await snapshot(guest.request)).match;
    expect(final.round).toBe(6); expect(final.history).toHaveLength(7);
    expect(final.events.filter((e: { kind: string; countryId: string; targetId: string }) => e.kind === 'attack' && e.countryId === 'norway' && e.targetId === 'germany')).toHaveLength(2);
    expect(final.events.find((e: { kind: string }) => e.kind === 'donation')).not.toHaveProperty('countryId');
    await guest.screenshot({ path: testInfo.outputPath('final-statistics.png'), fullPage: true });
    expect(errors).toEqual([]);
  } finally { await guestContext.close(); await teammateContext.close(); }
});
