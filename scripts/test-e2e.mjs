import { fork, spawn } from 'node:child_process';
import { once } from 'node:events';
import { setTimeout as delay } from 'node:timers/promises';
import { root } from './env.mjs';

async function ready() {
  try {
    const [web, api] = await Promise.all([
      fetch('http://127.0.0.1:5173', { signal: AbortSignal.timeout(1000) }),
      fetch('http://127.0.0.1:8080/readyz', { signal: AbortSignal.timeout(1000) }),
    ]);
    return web.ok && api.ok;
  } catch { return false; }
}

let server;
try {
  if (!await ready()) {
    // IPC lets the launcher stop its direct children on Windows as well as Unix.
    server = fork('scripts/dev.mjs', [], { cwd: root, stdio: ['inherit', 'inherit', 'inherit', 'ipc'] });
    const deadline = Date.now() + 120_000;
    while (!await ready()) {
      if (server.exitCode !== null || Date.now() > deadline) throw new Error('Development server did not become ready');
      await delay(300);
    }
  }
  const tests = spawn(process.execPath, ['node_modules/@playwright/test/cli.js', 'test', ...process.argv.slice(2)], {
    cwd: root, stdio: 'inherit', env: { ...process.env, FUNTIME_EXTERNAL_SERVER: '1' },
  });
  const [code] = await once(tests, 'exit');
  process.exitCode = code ?? 1;
} catch (error) {
  console.error(error);
  process.exitCode = 1;
} finally {
  if (server?.connected) {
    const stopped = once(server, 'exit');
    server.send('stop');
    await stopped;
  }
}
