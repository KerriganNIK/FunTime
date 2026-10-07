import { spawn, spawnSync } from 'node:child_process';
import { join } from 'node:path';
import { root, cache, goEnv } from './env.mjs';

// Build first so shutdown terminates service binaries, without orphaned go-run children.
const build = spawnSync(process.execPath, ['scripts/go.mjs', 'build'], { cwd: root, stdio: 'inherit' });
if (build.status !== 0) process.exit(build.status ?? 1);
const ext = process.platform === 'win32' ? '.exe' : '';
const children = [];
let stopping = false;
function stop(code = 0) {
  if (stopping) return;
  stopping = true;
  for (const child of children) child.kill('SIGTERM');
  process.exitCode = code;
  if (process.connected) process.disconnect();
}
function start(command, args, env = {}) {
  const child = spawn(command, args, { cwd: root, env: { ...goEnv, ...env }, stdio: 'inherit' });
  children.push(child);
  child.on('error', error => { console.error(error); stop(1); });
  child.on('exit', code => { if (!stopping) stop(code || 1); });
}
start(join(cache, 'bin', `catalog${ext}`), [], { GRPC_ADDR: '127.0.0.1:9001', HEALTH_ADDR: '127.0.0.1:9002' });
start(join(cache, 'bin', `world-domination${ext}`), [], { GRPC_ADDR: '127.0.0.1:9005', HEALTH_ADDR: '127.0.0.1:9006', DATA_DIR: join(cache, 'data', 'world-domination') });
start(join(cache, 'bin', `room${ext}`), [], { GRPC_ADDR: '127.0.0.1:9003', HEALTH_ADDR: '127.0.0.1:9004', DATA_DIR: join(cache, 'data', 'room') });
start(join(cache, 'bin', `gateway${ext}`), [], { HTTP_ADDR: '127.0.0.1:8080', CATALOG_ADDR: '127.0.0.1:9001' });
start(process.execPath, ['node_modules/vite/bin/vite.js', 'frontend', '--host', process.env.FUNTIME_DEV_HOST || '0.0.0.0'], { GATEWAY_URL: 'http://127.0.0.1:8080' });
process.on('SIGINT', () => stop());
process.on('SIGTERM', () => stop());
process.on('disconnect', () => stop());
process.on('message', message => {
  if (message === 'stop') {
    stop();
  }
});
