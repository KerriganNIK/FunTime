import { test, expect } from '@playwright/test';

test('a phone joins by invitation and submits a plan over HTTP without randomUUID', async ({ page: phone, browser, isMobile }, testInfo) => {
  test.skip(!isMobile, 'Phone interaction test');
  test.setTimeout(90_000);
  const hostContext = await browser.newContext({ baseURL: 'http://127.0.0.1:5173', viewport: { width: 1440, height: 1000 } });
  const host = await hostContext.newPage();
  const errors: string[] = [];
  const requestIds: string[] = [];
  phone.on('pageerror', error => errors.push(error.message));
  phone.on('request', request => { if (request.url().endsWith('/commands')) requestIds.push(request.postDataJSON().requestId); });
  // LAN browsers lack this secure-context method. Exercise the fallback on CI too.
  await phone.addInitScript(() => Object.defineProperty(crypto, 'randomUUID', { value: undefined, configurable: true }));
  try {
    await host.goto('/');
    await host.getByRole('button', { name: 'Об игре', exact: true }).click();
    await host.getByRole('dialog').getByLabel('Ваше имя').fill('Ведущий с компьютера');
    await host.getByRole('button', { name: 'Создать комнату', exact: true }).click();
    await expect(host).toHaveURL(/\/rooms\/[A-Z2-9]{6}$/);
    const code = host.url().split('/').at(-1)!;
    await host.getByRole('combobox', { name: 'Страна', exact: true }).selectOption('norway');
    await host.getByRole('button', { name: 'Занять место' }).click();
    await host.getByRole('button', { name: 'Пригласить игроков', exact: true }).click();
    const invitation = host.getByRole('dialog');
    const url = await invitation.getByLabel('Ссылка для игроков').inputValue();
    await expect(invitation).toContainText(code);
    await invitation.getByRole('button', { name: 'Скопировать ссылку' }).click();
    await expect(invitation.getByRole('status')).toContainText(/Ссылка скопирована|Ссылка выделена/);
    await host.getByRole('button', { name: 'Закрыть окно' }).click();

    await phone.goto(url);
    await phone.getByLabel('Ваше имя').fill('Игрок с телефона');
    expect(await phone.getByLabel('Ваше имя').evaluate(input => parseFloat(getComputedStyle(input).fontSize))).toBeGreaterThanOrEqual(16);
    await phone.getByRole('button', { name: 'Присоединиться', exact: true }).tap();
    await phone.getByRole('combobox', { name: 'Страна', exact: true }).selectOption('germany');
    await phone.getByRole('button', { name: 'Занять место' }).tap();
    await expect(phone.getByText('Вы в команде «Германия», Президент.')).toBeVisible();
    await host.getByRole('button', { name: 'Начать игру', exact: true }).click();
    await host.getByRole('button', { name: 'Открыть штабы', exact: true }).click();
    const orders = phone.getByRole('region', { name: 'Германия', exact: true });
    const navigation = orders.getByRole('navigation', { name: 'Разделы плана' });
    await expect(navigation).toBeVisible();
    await orders.locator('.game-city').first().getByRole('checkbox', { name: /Улучшить/ }).tap();
    await navigation.getByRole('button', { name: 'Экономика', exact: true }).tap();
    await orders.getByRole('checkbox', { name: /Улучшить экологию/ }).tap();
    await navigation.getByRole('button', { name: 'Оборона', exact: true }).tap();
    await expect(orders.getByRole('heading', { name: 'Ядерная программа' })).toBeVisible();
    await navigation.getByRole('button', { name: 'Города', exact: true }).tap();
    await expect(orders.locator('.game-city').first().getByRole('checkbox', { name: /Улучшить/ })).toBeChecked();
    await orders.getByRole('button', { name: 'Сохранить план' }).tap();
    await expect(orders.getByText('План сохранён для всей команды')).toBeVisible();
    const response = await phone.request.get(new URL(`/api/v1/rooms/${code}`, url).href);
    const german = (await response.json()).match.countries.find((c: { id: string }) => c.id === 'germany');
    expect(german.plan).toMatchObject({ upgrades: ['germany-1'], ecology: true });
    expect(german.planCostCents).toBe(40000);
    await phone.screenshot({ path: testInfo.outputPath('phone-headquarters.png') });
    await phone.getByRole('button', { name: 'Включить тёмную тему' }).tap();
    await phone.screenshot({ path: testInfo.outputPath('phone-headquarters-dark.png') });
    for (const viewport of [{width:320,height:640},{width:393,height:851},{width:844,height:390}]) {
      await phone.setViewportSize(viewport);
      expect(await phone.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `viewport ${viewport.width}`).toBeTruthy();
    }
    expect(requestIds.length).toBeGreaterThanOrEqual(2);
    expect(new Set(requestIds).size).toBe(requestIds.length);
    for (const id of requestIds) expect(id).toMatch(/^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$/);
    expect(errors).toEqual([]);
  } finally { await hostContext.close(); }
});
