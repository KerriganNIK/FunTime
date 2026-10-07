import type { Locator } from '@playwright/test';

export async function selectOrderSection(orders: Locator, name: string) {
  const navigation = orders.getByRole('navigation', { name: 'Разделы плана' });
  if (await orders.page().evaluate(() => matchMedia('(max-width: 680px)').matches)) {
    await navigation.getByRole('button', { name, exact: true }).click();
  }
}
