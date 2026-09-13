import { spawnSync } from 'node:child_process';
import { mkdirSync } from 'node:fs';
import { join } from 'node:path';
import { root, cache, goEnv } from './env.mjs';

const mode = process.argv[2];
if (!['build', 'test', 'tidy'].includes(mode)) throw new Error('Expected build, test, or tidy');
mkdirSync(join(cache, 'bin'), { recursive: true });
for (const name of ['contracts/gen/go', 'services/catalog', 'services/gateway']) {
  const service = name.split('/').at(-1);
  const args = mode === 'build' && name.startsWith('services')
    ? ['build', '-o', join(cache, 'bin', `${service}${process.platform === 'win32' ? '.exe' : ''}`), './cmd/server']
    : mode === 'tidy' ? ['mod', 'tidy'] : [mode, ...(mode === 'test' ? ['-count=1'] : []), './...'];
  const result = spawnSync('go', args, { cwd: join(root, name), env: goEnv, stdio: 'inherit' });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}
