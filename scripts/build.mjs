import { execFileSync } from 'node:child_process';
import { copyFileSync, mkdirSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
const output = join(root, 'dist');

mkdirSync(output, { recursive: true });
execFileSync(process.execPath, [
  join(root, 'web', 'node_modules', 'typescript', 'bin', 'tsc'),
  '--project', join(root, 'web', 'tsconfig.json'),
], { cwd: root, stdio: 'inherit' });

execFileSync('go', ['build', '-trimpath', '-ldflags=-s -w', '-o', join(output, 'inventory.wasm'), './cmd/wasm'], {
  cwd: root,
  env: { ...process.env, GOOS: 'js', GOARCH: 'wasm' },
  stdio: 'inherit',
});

const goRoot = execFileSync('go', ['env', 'GOROOT'], { cwd: root, encoding: 'utf8' }).trim();
copyFileSync(join(goRoot, 'lib', 'wasm', 'wasm_exec.js'), join(output, 'wasm_exec.js'));
copyFileSync(join(goRoot, 'LICENSE'), join(output, 'GO-LICENSE.txt'));
for (const file of ['index.html', 'styles.css', 'dist/main.js']) {
  copyFileSync(join(root, 'web', file), join(output, file.split('/').at(-1)));
}
