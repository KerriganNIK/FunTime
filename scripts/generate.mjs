import { spawnSync } from 'node:child_process';
import { mkdirSync } from 'node:fs';
import { join } from 'node:path';
import { root, cache, goEnv } from './env.mjs';

function run(command, args) {
  const result = spawnSync(command, args, { cwd: root, env: goEnv, stdio: 'inherit' });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}
mkdirSync(join(cache, 'bin'), { recursive: true });
run('go', ['install', 'google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11']);
run('go', ['install', 'google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1']);
const ext = process.platform === 'win32' ? '.exe' : '';
run(process.execPath, [
  'node_modules/@bufbuild/buf/bin/buf', 'generate', 'contracts/proto',
  '--template', JSON.stringify({ version: 'v2', plugins: [
    { local: join(cache, 'bin', `protoc-gen-go${ext}`), out: 'contracts/gen/go', opt: ['paths=source_relative'] },
    { local: join(cache, 'bin', `protoc-gen-go-grpc${ext}`), out: 'contracts/gen/go', opt: ['paths=source_relative'] },
  ] }),
]);
run(process.execPath, ['node_modules/openapi-typescript/bin/cli.js', 'contracts/http/openapi.yaml', '-o', 'frontend/src/shared/api/generated/schema.ts']);
