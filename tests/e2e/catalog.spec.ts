import { test, expect } from '@playwright/test';

test('real catalog, search, details, keyboard dismissal and theme persistence', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'Мировое господство', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Об игре', exact: true }).click();
  await expect(page.getByRole('dialog')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Создать комнату', exact: true })).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(page.getByRole('dialog')).not.toBeVisible();
  await expect(page.getByRole('button', { name: 'Об игре', exact: true })).toBeFocused();
  await page.getByRole('searchbox', { name: 'Найти игру' }).fill('нет такой игры');
  await expect(page.getByRole('heading', { name: 'Такой игры пока нет' })).toBeVisible();
  await page.getByRole('button', { name: 'Показать все игры' }).click();
  await page.getByRole('button', { name: 'Скоро', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Такой игры пока нет' })).toBeVisible();
  await page.getByRole('button', { name: 'Показать все игры' }).click();
  await page.getByRole('button', { name: 'Включить тёмную тему' }).click();
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  await page.reload();
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  await page.getByRole('button', { name: 'Войти по коду' }).click();
  await expect(page.getByRole('heading', { name: 'Присоединиться к друзьям' })).toBeVisible();
  await page.getByRole('button', { name: 'Закрыть окно' }).last().click();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBeTruthy();
});

test('catalog failure is recoverable and malformed payload is rejected', async ({ page }) => {
  await page.route('**/api/v1/games', route => route.fulfill({ status: 503, json: { error: { code: 'CATALOG_UNAVAILABLE' } } }));
  await page.goto('/');
  await expect(page.getByRole('alert')).toContainText('Не удалось загрузить игры');
  await page.unroute('**/api/v1/games');
  await page.getByRole('button', { name: 'Попробовать снова' }).click();
  await expect(page.getByRole('heading', { name: 'Мировое господство', exact: true })).toBeVisible();
  await page.route('**/api/v1/games', route => route.fulfill({ json: { games: [{ id: 'broken' }] } }));
  await page.reload();
  await expect(page.getByRole('alert')).toBeVisible();
});

test('empty catalog and unknown page have useful states', async ({ page }) => {
  await page.route('**/api/v1/games', route => route.fulfill({ json: { games: [] } }));
  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'Коллекция скоро появится' })).toBeVisible();
  await page.goto('/missing');
  await expect(page.getByRole('heading', { name: 'Кажется, мы вышли за карту.' })).toBeVisible();
  await page.getByRole('link', { name: 'На главную FunTime' }).click();
  await expect(page).toHaveURL('/');
});

test('small screens do not overflow horizontally', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'Мировое господство', exact: true })).toBeVisible();
  await page.evaluate(() => document.fonts.ready);
  for (const width of [320, 360, 768]) {
    await page.setViewportSize({ width, height: 900 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `viewport ${width}`).toBeTruthy();
  }
});
