import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';
import { indexResults } from '../scripts/indexResults.js';

async function fixture(t) {
  const scratch = fileURLToPath(new URL('../../.dev/index-results-tests/', import.meta.url));
  await fs.mkdir(scratch, { recursive: true });
  const root = await fs.mkdtemp(path.join(scratch, 'case-'));
  t.after(async () => {
    assert.equal(path.dirname(root), path.resolve(scratch));
    await fs.rm(root, { recursive: true, force: true });
  });
  return {
    root,
    async add(file) {
      const target = path.join(root, 'public/result', file);
      await fs.mkdir(path.dirname(target), { recursive: true });
      await fs.writeFile(target, '{}');
    },
    async read() {
      return {
        index: JSON.parse(await fs.readFile(path.join(root, 'src/generated/results.json'), 'utf8')),
        manifest: JSON.parse(await fs.readFile(path.join(root, 'public/result/manifest.json'), 'utf8')),
      };
    },
  };
}

test('a fresh checkout gets empty dashboard and download indexes', async t => {
  const f = await fixture(t);
  await indexResults(f.root);
  assert.deepEqual(await f.read(), {
    index: { results: [], logArchives: [], jsonReports: {}, htmlReports: {} },
    manifest: { version: 1, results: [] },
  });
});

test('indexes JSON-only, historical HTML, and mixed runs with sorted download paths', async t => {
  const f = await fixture(t);
  for (const file of [
    '20261022-120000/buildx.json',
    '20261022-120000/buildkit.json',
    '20261022-120000/raw/input.json',
    '20261020-120000/buildkit.json',
    '20261020-120000/buildkit.html',
    '20261020-120000/index.html',
    '20261020-120000/buildx.html',
    '20261020-120000/logs.tar.gz',
    '20260301-120000/index.html',
    'manifest.json',
  ]) await f.add(file);
  await indexResults(f.root);
  const { index, manifest } = await f.read();
  assert.deepEqual(index, {
    results: ['20260301-120000', '20261020-120000', '20261022-120000'],
    logArchives: ['20261020-120000'],
    jsonReports: {
      '20260301-120000': [],
      '20261020-120000': ['buildkit'],
      '20261022-120000': ['buildkit', 'buildx'],
    },
    htmlReports: {
      '20260301-120000': { buildkit: 'index.html' },
      '20261020-120000': { buildkit: 'buildkit.html', buildx: 'buildx.html' },
      '20261022-120000': {},
    },
  });
  assert.deepEqual(manifest, { version: 1, results: [
    { name: '20260301-120000', files: ['index.html'] },
    { name: '20261020-120000', files: ['buildkit.html', 'buildkit.json', 'buildx.html', 'index.html', 'logs.tar.gz'] },
    { name: '20261022-120000', files: ['buildkit.json', 'buildx.json', 'raw/input.json'] },
  ] });
  await indexResults(f.root);
  assert.deepEqual(await f.read(), { index, manifest });
});

test('reindexing removes expired logs and runs from both outputs', async t => {
  const f = await fixture(t);
  await f.add('20261020-120000/buildkit.json');
  await f.add('20261020-120000/logs.tar.gz');
  await f.add('20260301-120000/index.html');
  await indexResults(f.root);
  await fs.unlink(path.join(f.root, 'public/result/20261020-120000/logs.tar.gz'));
  await fs.unlink(path.join(f.root, 'public/result/20260301-120000/index.html'));
  await fs.rmdir(path.join(f.root, 'public/result/20260301-120000'));
  await indexResults(f.root);
  const { index, manifest } = await f.read();
  assert.deepEqual(index.results, ['20261020-120000']);
  assert.deepEqual(index.logArchives, []);
  assert.deepEqual(manifest.results, [{ name: '20261020-120000', files: ['buildkit.json'] }]);
});
