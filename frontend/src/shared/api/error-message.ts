export function errorMessage(error: unknown): string {
  if (typeof error === 'object' && error !== null && 'data' in error) {
    const data = error.data;
    if (typeof data === 'object' && data !== null && 'error' in data && typeof data.error === 'object' && data.error !== null && 'message' in data.error && typeof data.error.message === 'string') return data.error.message;
  }
  return 'Не удалось выполнить запрос. Проверьте соединение и повторите.';
}
