import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { once } from 'node:events';
import { after, before, test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';

let browser;
let server;
let localURL;

before(async () => {
  browser = await chromium.launch();
  const serve = fileURLToPath(new URL('../../scripts/serve.mjs', import.meta.url));
  server = spawn(process.execPath, [serve, '0'], { stdio: ['ignore', 'pipe', 'inherit'] });
  const line = await Promise.race([
    once(server.stdout, 'data').then(([data]) => data.toString()),
    once(server, 'exit').then(([code]) => { throw new Error(`El servidor terminó con código ${code}.`); }),
  ]);
  const match = line.match(/http:\/\/127\.0\.0\.1:\d+\//);
  assert.ok(match, `El servidor no informó su URL: ${line}`);
  localURL = match[0];
}, { timeout: 30_000 });

after(async () => {
  if (server && server.exitCode === null) {
    const exited = once(server, 'exit');
    server.kill();
    await exited;
  }
  await browser?.close();
});

async function openDemo(page, url = localURL) {
  const response = await page.goto(url);
  assert.equal(response.status(), 200);
  await page.locator('#report').waitFor({ state: 'visible' });
}

async function metrics(page, strategy) {
  const values = await page.locator(`[data-strategy="${strategy}"] dd`).allTextContents();
  return values.map((value) => Number(value.trim()));
}

test('un visitante compara los diseños y puede cambiar el incidente', { timeout: 30_000 }, async () => {
  const page = await browser.newPage({ viewport: { width: 360, height: 800 } });
  const errors = [];
  page.on('pageerror', (error) => errors.push(error.message));
  try {
    await openDemo(page, process.env.DEMO_URL || localURL);
    assert.deepEqual(await metrics(page, 'quorum'), [90, 10, 0, 90, 0, 30, 0]);
    assert.deepEqual(await metrics(page, 'quota'), [70, 0, 30, 70, 0, 50, 30]);

    await page.getByRole('button', { name: 'Sin mayoría' }).press('Enter');
    assert.deepEqual(await metrics(page, 'quorum'), [0, 80, 0, 0, 0, 120, 0]);
    assert.deepEqual(await metrics(page, 'quota'), [80, 0, 0, 80, 0, 40, 0]);

    await page.getByRole('button', { name: 'Pérdida de B' }).click();
    assert.deepEqual(await metrics(page, 'quorum'), [40, 60, 0, 40, 0, 80, 0]);
    assert.deepEqual(await metrics(page, 'quota'), [100, 0, 0, 40, 60, 20, 0]);

    await page.getByLabel('Pérdida permanente posterior').selectOption('A');
    await page.getByRole('button', { name: /Simular escenario/ }).click();
    assert.deepEqual(await metrics(page, 'quota'), [100, 0, 0, 60, 40, 20, 0]);
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test('una asignación inválida se bloquea y puede corregirse', { timeout: 30_000 }, async () => {
  const page = await browser.newPage();
  try {
    await openDemo(page);
    await page.getByLabel('Stock total').fill('20');
    await page.getByRole('button', { name: /Simular escenario/ }).click();
    assert.equal(await page.locator('#quota-a').evaluate((field) => field.validity.rangeOverflow), true);
    assert.match(await page.locator('#status').innerText(), /Resultados anteriores/);

    await page.getByLabel('Cupo de A').fill('10');
    await page.getByRole('button', { name: /Simular escenario/ }).click();
    assert.deepEqual(await metrics(page, 'quorum'), [20, 10, 70, 20, 0, 0, 0]);
    assert.deepEqual(await metrics(page, 'quota'), [20, 0, 80, 20, 0, 0, 0]);
  } finally {
    await page.close();
  }
});

test('un fallo del motor oculta el resultado anterior', { timeout: 30_000 }, async () => {
  const page = await browser.newPage();
  try {
    await openDemo(page);
    await page.evaluate(() => {
      window.simulateBeforeFailure = window.inventorySimulate;
      window.inventorySimulate = () => '{"error":"Error de prueba del motor."}';
    });
    await page.getByRole('button', { name: /Simular escenario/ }).click();
    await page.getByRole('alert').waitFor({ state: 'visible' });
    assert.match(await page.getByRole('alert').innerText(), /Error de prueba/);
    assert.equal(await page.locator('#report').isVisible(), false);

    await page.evaluate(() => {
      window.inventorySimulate = window.simulateBeforeFailure;
      delete window.simulateBeforeFailure;
    });
    await page.getByRole('button', { name: /Simular escenario/ }).click();
    assert.equal(await page.locator('#report').isVisible(), true);
    assert.equal(await page.getByRole('alert').isVisible(), false);
    assert.deepEqual(await metrics(page, 'quorum'), [90, 10, 0, 90, 0, 30, 0]);
  } finally {
    await page.close();
  }
});

test('si WebAssembly no carga, la demo muestra un error en lugar de cifras', { timeout: 30_000 }, async () => {
  const page = await browser.newPage();
  try {
    await page.route('**/inventory.wasm', (route) => route.abort());
    await page.goto(localURL);
    await page.getByRole('alert').waitFor({ state: 'visible' });
    assert.match(await page.getByRole('alert').innerText(), /No se pudo iniciar/);
    assert.equal(await page.locator('#report').isVisible(), false);
    assert.equal(await page.getByRole('button', { name: /Simular escenario/ }).isEnabled(), false);
  } finally {
    await page.close();
  }
});
