const assert = require('node:assert/strict');
const fs = require('node:fs');
const http = require('node:http');
const path = require('node:path');
const { spawn } = require('node:child_process');
const test = require('node:test');

async function fixture(t, manifest, files = {}, statuses = {}) {
  const scratch = path.resolve(__dirname, '../../.dev/download-results-tests');
  fs.mkdirSync(scratch, { recursive: true });
  const root = fs.mkdtempSync(path.join(scratch, 'case-'));
  const results = path.join(root, 'results');
  const requests = [];
  const server = http.createServer((req, res) => {
    requests.push(req.url);
    const status = statuses[req.url]?.shift();
    if (status) {
      res.writeHead(status).end();
    } else if (req.url === '/bench/result/manifest.json') {
      res.setHeader('Content-Type', 'application/json');
      res.end(JSON.stringify(manifest));
    } else if (Object.hasOwn(files, req.url)) {
      res.end(files[req.url]);
    } else {
      res.writeHead(404).end();
    }
  });
  t.after(async () => {
    await new Promise(resolve => server.close(resolve));
    assert.equal(path.dirname(root), scratch);
    fs.rmSync(root, { recursive: true, force: true });
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  return {
    results,
    requests,
    run(...args) {
      return new Promise((resolve, reject) => {
        const child = spawn(process.execPath, [path.join(__dirname, 'downloadResults.js'), ...args], {
          cwd: root,
          env: {
            ...process.env,
            PAGES_URL: `http://127.0.0.1:${server.address().port}/bench`,
            RESULTS_DIR: results,
          },
          windowsHide: true,
        });
        let stdout = '', stderr = '';
        child.stdout.on('data', data => { stdout += data; });
        child.stderr.on('data', data => { stderr += data; });
        child.on('error', reject);
        child.on('close', code => resolve({ code, stdout, stderr }));
      });
    },
  };
}

test('list emits only JSON and does not download files or create output', async t => {
  const f = await fixture(t, { version: 1, results: [{ name: 'run-a', files: ['buildkit.json'] }] });
  const result = await f.run('--list');
  assert.equal(result.code, 0, result.stderr);
  assert.deepEqual(JSON.parse(result.stdout), ['run-a']);
  assert.deepEqual(f.requests, ['/bench/result/manifest.json']);
  assert.ok(!fs.existsSync(f.results));
});

test('downloads all runs and preserves nested inputs, legacy HTML, and binary logs', async t => {
  const archive = Buffer.from([0, 255, 31, 139]);
  const f = await fixture(t, { version: 1, results: [
    { name: 'run-a', files: ['buildkit.json', 'raw/input file.json', 'logs.tar.gz'] },
    { name: 'run-b', files: ['index.html'] },
  ] }, {
    '/bench/result/run-a/buildkit.json': '{}',
    '/bench/result/run-a/raw/input%20file.json': '{"raw":true}',
    '/bench/result/run-a/logs.tar.gz': archive,
    '/bench/result/run-b/index.html': '<html></html>',
  });
  const result = await f.run();
  assert.equal(result.code, 0, result.stderr);
  assert.equal(fs.readFileSync(path.join(f.results, 'run-a/raw/input file.json'), 'utf8'), '{"raw":true}');
  assert.deepEqual(fs.readFileSync(path.join(f.results, 'run-a/logs.tar.gz')), archive);
  assert.equal(fs.readFileSync(path.join(f.results, 'run-b/index.html'), 'utf8'), '<html></html>');
  assert.match(result.stdout, /Downloaded 4 file\(s\) for 2 result\(s\)/);
});

test('selects named runs and rejects missing names before writing files', async t => {
  const f = await fixture(t, { results: [
    { name: 'run-a', files: ['buildkit.json'] },
    { name: 'run-b', files: ['buildx.json'] },
  ] }, { '/bench/result/run-b/buildx.json': '{}' });
  const missing = await f.run('run-b', 'missing');
  assert.equal(missing.code, 1);
  assert.match(missing.stderr, /does not contain: missing/);
  assert.ok(!fs.existsSync(f.results));
  const selected = await f.run('run-b');
  assert.equal(selected.code, 0, selected.stderr);
  assert.deepEqual(fs.readdirSync(f.results), ['run-b']);
});

test('rejects traversal anywhere in the manifest before downloading', async t => {
  const f = await fixture(t, { results: [
    { name: 'run-a', files: ['buildkit.json'] },
    { name: 'run-b', files: ['../outside.json'] },
  ] });
  const result = await f.run('run-a');
  assert.equal(result.code, 1);
  assert.match(result.stderr, /Invalid manifest file path/);
  assert.deepEqual(f.requests, ['/bench/result/manifest.json']);
  assert.ok(!fs.existsSync(f.results));
});

test('retries transient responses but fails missing files without retrying', async t => {
  const url = '/bench/result/run-a/buildkit.json';
  const f = await fixture(t, { results: [{ name: 'run-a', files: ['buildkit.json'] }] }, {
    [url]: '{}',
  }, { [url]: [503] });
  const result = await f.run();
  assert.equal(result.code, 0, result.stderr);
  assert.equal(f.requests.filter(request => request === url).length, 2);
  assert.match(result.stderr, /retrying/);

  const missing = await fixture(t, { results: [{ name: 'run-a', files: ['missing.json'] }] });
  const failure = await missing.run();
  assert.equal(failure.code, 1);
  assert.match(failure.stderr, /404/);
  assert.equal(missing.requests.length, 2);
});
