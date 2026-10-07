import { test, expect, type APIRequestContext } from '@playwright/test';
import { selectOrderSection } from '../helpers/order-section';

test('players can reach headquarters, click purchases, save them and see settlement', async ({ page: host, browser }) => {
  test.setTimeout(90_000);
  const guestContext = await browser.newContext({ baseURL: 'http://127.0.0.1:5173', viewport: { width: 390, height: 844 } });
  const guest = await guestContext.newPage();
  const errors: string[] = [];
  host.on('pageerror', error => errors.push(error.message));
  guest.on('pageerror', error => errors.push(error.message));
  try {
    const created = await host.request.post('/api/v1/rooms', { data: { gameId: 'world-domination', name: 'Проверка покупок' } });
    expect(created.ok()).toBeTruthy();
    const { code } = await created.json();
    const path = `/api/v1/rooms/${code}`;
    const command = async (request: APIRequestContext, kind: string, payload: unknown) => request.post(`${path}/commands`, { data: { kind, payload, requestId: crypto.randomUUID() } });
    const snapshot = async (request: APIRequestContext) => (await request.get(path)).json();
    expect((await command(host.request, 'select', { countryId: 'norway', role: 'president' })).ok()).toBeTruthy();
    expect((await guest.request.post(`${path}/join`, { data: { name: 'Игрок' } })).ok()).toBeTruthy();
    expect((await command(guest.request, 'select', { countryId: 'germany', role: 'president' })).ok()).toBeTruthy();
    expect((await command(host.request, 'start', { phaseSeconds: 720 })).ok()).toBeTruthy();
    await Promise.all([host.goto(`/rooms/${code}`), guest.goto(`/rooms/${code}`)]);
    const hostOrders = host.getByRole('region', { name: 'Норвегия', exact: true });
    const guestOrders = guest.getByRole('region', { name: 'Германия', exact: true });

    await expect(hostOrders.getByRole('heading', { name: 'Покупки откроются в штабе' })).toBeVisible();
    await expect(guestOrders.getByRole('status')).toContainText('когда ведущий перейдёт к фазе штаба');
    await expect(hostOrders.getByRole('checkbox')).toHaveCount(0);
    await expect(guestOrders.getByRole('checkbox')).toHaveCount(0);
    await expect(guest.getByText('Переговоры с другими странами откроются в фазе штаба.')).toBeVisible();
    const before = await snapshot(guest.request);
    expect((await command(guest.request, 'plan', before.match.countries.find((c: { id: string }) => c.id === 'germany').plan)).status()).toBe(400);

    await hostOrders.getByRole('button', { name: 'Перейти к покупкам' }).click();
    await expect(guest.getByRole('heading', { name: 'Работа в штабе', exact: true })).toBeVisible();
    const hostUpgrade = hostOrders.locator('.game-city').first().getByRole('checkbox', { name: /Улучшить/ });
    const guestUpgrade = guestOrders.locator('.game-city').first().getByRole('checkbox', { name: /Улучшить/ });
    await expect(hostUpgrade).toBeEnabled();
    await expect(guestUpgrade).toBeEnabled();

    await host.getByRole('button', { name: 'Пауза', exact: true }).click();
    await expect(guestOrders.getByRole('heading', { name: 'Игра на паузе' })).toBeVisible();
    await expect(guestUpgrade).toBeDisabled();
    await hostOrders.getByRole('button', { name: 'Снять паузу' }).click();
    await expect(guestUpgrade).toBeEnabled();

    await hostUpgrade.click();
    await expect(hostUpgrade).toBeChecked();
    await expect(hostOrders.locator('.game-purchase-preview')).toContainText('200 монет');
    await expect(hostOrders.locator('.game-purchase-preview')).toContainText('500 монет');
    await hostUpgrade.click();
    await expect(hostUpgrade).not.toBeChecked();
    await hostUpgrade.click();
    await selectOrderSection(hostOrders, 'Оборона');
    await hostOrders.getByRole('checkbox', { name: /Запустить программу/ }).click();
    await hostOrders.getByRole('button', { name: 'Сохранить план' }).click();
    await expect(hostOrders.getByText('План сохранён для всей команды')).toBeVisible();
    await expect(hostOrders.getByText('Ожидаем расчёта раунда. Покупки ещё не списаны.')).toBeVisible();
    const saved = (await snapshot(host.request)).match.countries.find((c: { id: string }) => c.id === 'norway');
    expect(saved.balanceCents).toBe(70000);
    expect(saved.plan).toMatchObject({ upgrades: ['norway-1'], nuclear: true });
    await host.reload();
    await expect(hostUpgrade).toBeChecked();

    await guestUpgrade.click();
    await guestOrders.locator('.game-city').first().getByRole('checkbox', { name: /Щит/ }).click();
    await guestOrders.locator('.game-city').nth(3).getByRole('checkbox', { name: /Улучшить/ }).click();
    await selectOrderSection(guestOrders, 'Разведка');
    await guestOrders.locator('.game-spy-options').getByRole('checkbox', { name: /Норвегия/ }).click();
    await selectOrderSection(guestOrders, 'Экономика');
    const environment = guestOrders.locator('.game-subpanel').filter({ has: guest.getByRole('heading', { name: 'Экология и поддержка' }) });
    await environment.getByRole('checkbox', { name: /Улучшить экологию/ }).click();
    await expect(guestOrders.locator('.game-purchase-preview .game-error')).toHaveText('-200 монет');
    await environment.getByRole('checkbox', { name: /Улучшить экологию/ }).click();
    await environment.getByLabel('Пожертвовать: Норвегия').fill('25');
    await expect(guestOrders.locator('.game-purchase-preview .game-error')).toHaveText('-25 монет');
    await environment.getByLabel('Пожертвовать: Норвегия').fill('0');
    await environment.getByRole('checkbox', { name: 'Норвегия', exact: true }).click();
    await guestOrders.getByRole('button', { name: 'Сохранить план' }).click();
    await expect(guestOrders.getByText('План сохранён для всей команды')).toBeVisible();

    await host.getByRole('button', { name: 'Рассчитать раунд', exact: true }).click();
    await expect(guest.getByRole('heading', { name: 'Заседание ООН', exact: true })).toBeVisible();
    const second = (await snapshot(host.request)).match;
    expect(second.countries.find((c: { id: string }) => c.id === 'norway').balanceCents).toBe(79900);
    const germanSecond = second.countries.find((c: { id: string }) => c.id === 'germany');
    expect(germanSecond.balanceCents).toBe(94000);
    expect(germanSecond.cities[0]).toMatchObject({ development: 90, shield: true });
    expect(germanSecond.intelligence).toHaveLength(1);
    await expect(guestOrders.getByRole('checkbox')).toHaveCount(0);

    await hostOrders.getByRole('button', { name: 'Перейти к покупкам' }).click();
    await selectOrderSection(hostOrders, 'Оборона');
    await hostOrders.getByRole('spinbutton', { name: /Создать бомбы/ }).fill('1');
    await hostOrders.getByRole('combobox', { name: 'Цель удара' }).selectOption('germany-1');
    await hostOrders.getByRole('button', { name: 'Добавить запуск' }).click();
    await hostOrders.getByRole('button', { name: 'Сохранить план' }).click();
    await expect(hostOrders.getByText('План сохранён для всей команды')).toBeVisible();
    await host.getByRole('button', { name: 'Рассчитать раунд', exact: true }).click();
    await expect(host.getByRole('heading', { name: 'Заседание ООН', exact: true })).toBeVisible();
    const third = (await snapshot(host.request)).match;
    expect(third.countries.find((c: { id: string }) => c.id === 'norway').balanceCents).toBe(140900);
    expect(third.countries.find((c: { id: string }) => c.id === 'germany').cities[0]).toMatchObject({ development: 87, shield: false, destroyed: false });
    expect(errors).toEqual([]);
  } finally { await guestContext.close(); }
});
