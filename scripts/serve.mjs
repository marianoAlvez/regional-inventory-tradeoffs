import { createReadStream } from 'node:fs';
import { createServer } from 'node:http';
import { fileURLToPath } from 'node:url';

const assets = new Map([
  ['/', ['index.html', 'text/html; charset=utf-8']],
  ['/styles.css', ['styles.css', 'text/css; charset=utf-8']],
  ['/main.js', ['main.js', 'text/javascript; charset=utf-8']],
  ['/wasm_exec.js', ['wasm_exec.js', 'text/javascript; charset=utf-8']],
  ['/inventory.wasm', ['inventory.wasm', 'application/wasm']],
]);

const port = Number(process.argv[2] ?? 4173);
createServer((request, response) => {
  const asset = assets.get(request.url);
  if (!asset) {
    response.writeHead(404).end();
    return;
  }
  const [file, type] = asset;
  const stream = createReadStream(fileURLToPath(new URL(`../dist/${file}`, import.meta.url)));
  stream.once('error', () => response.writeHead(404).end());
  stream.once('open', () => {
    response.writeHead(200, { 'Content-Type': type, 'Cache-Control': 'no-store' });
    stream.pipe(response);
  });
}).listen(port, '127.0.0.1', () => {
  console.log(`Demo en http://127.0.0.1:${port}/`);
});
