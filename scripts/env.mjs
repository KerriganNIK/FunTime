import { mkdirSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';

export const root = fileURLToPath(new URL('../', import.meta.url));
export const cache = join(root, '.cache');
mkdirSync(cache, { recursive: true });
export const goEnv = {
  ...process.env,
  GOCACHE: join(cache, 'go-build'),
  GOMODCACHE: join(cache, 'go-mod'),
  GOBIN: join(cache, 'bin'),
};
