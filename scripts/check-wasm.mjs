import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
await import('../dist/wasm_exec.js');
const go = new globalThis.Go();
const wasm = await readFile(new URL('../dist/inventory.wasm', import.meta.url));
const { instance } = await WebAssembly.instantiate(wasm, go.importObject);
void go.run(instance);
assert.equal(typeof globalThis.inventorySimulate, 'function');

const base = { stock: 120, quotaA: 60, demandA: 90, demandB: 10, majority: 'A', loss: 'A' };
for (const input of [
  base,
  { ...base, demandA: 50, demandB: 40, majority: 'none', loss: 'none' },
  { ...base, stock: 0, quotaA: 0, demandA: 3, demandB: 2, majority: 'B' },
]) {
  const json = JSON.stringify(input);
  const browser = JSON.parse(globalThis.inventorySimulate(json));
  const cli = JSON.parse(execFileSync('go', ['run', './cmd/inventory'], {
    cwd: root, input: json, encoding: 'utf8',
  }));
  assert.deepEqual(browser, cli);
}

const { results: [quorum, quota] } = JSON.parse(globalThis.inventorySimulate(JSON.stringify(base)));
assert.deepEqual([quorum.confirmed, quorum.pending, quorum.lost], [90, 10, 0]);
assert.deepEqual([quota.confirmed, quota.rejected, quota.lost, quota.stranded], [70, 30, 60, 30]);
assert.equal(typeof JSON.parse(globalThis.inventorySimulate('malformed')).error, 'string');
console.log('WASM y CLI coinciden en los tres escenarios. Contrato de error verificado.');
