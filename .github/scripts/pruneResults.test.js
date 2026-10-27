const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const test = require('node:test');
const { retentionSettings, cutoffDate, resultDate, planPruning } = require('./pruneResults');

const now = new Date('2026-10-22T12:00:00Z');
const settings = retentionSettings({});

function fixture(t) {
  const scratch = path.resolve(__dirname, '../../.dev/prune-results-tests');
  fs.mkdirSync(scratch, { recursive: true });
  const root = fs.mkdtempSync(path.join(scratch, 'case-'));
  t.after(() => {
    assert.equal(path.dirname(root), scratch);
    fs.rmSync(root, { recursive: true, force: true });
  });
  return {
    root,
    add(name, files = {}) {
      fs.mkdirSync(path.join(root, name), { recursive: true });
      for (const [file, contents] of Object.entries(files)) {
        const target = path.join(root, name, file);
        fs.mkdirSync(path.dirname(target), { recursive: true });
        fs.writeFileSync(target, contents);
      }
    },
  };
}

test('retention settings reject invalid values instead of silently deleting runs', () => {
  assert.deepEqual(settings, { monthsToKeep: 12, logsMonthsToKeep: 3, keepResultsPerDay: 1 });
  for (const value of ['', '-1', '1.5', '3months', 'NaN', '9007199254740992']) {
    assert.throws(() => retentionSettings({ MONTHS_TO_KEEP: value }));
  }
  assert.throws(() => retentionSettings({ MONTHS_TO_KEEP: '0' }));
  assert.throws(() => retentionSettings({ LOGS_MONTHS_TO_KEEP: '0' }));
  assert.throws(() => retentionSettings({ KEEP_RESULTS_PER_DAY: '-1' }));
  assert.equal(retentionSettings({ KEEP_RESULTS_PER_DAY: '0' }).keepResultsPerDay, 0);
});

test('UTC cutoffs clamp month ends and recognize only valid timestamp directories', () => {
  assert.equal(cutoffDate(new Date('2024-03-31T23:00:00Z'), 1).toISOString(), '2024-02-29T00:00:00.000Z');
  assert.equal(cutoffDate(new Date('2025-03-31T23:00:00Z'), 1).toISOString(), '2025-02-28T00:00:00.000Z');
  for (const name of ['20260230-120000', '20261301-120000', '20261022-240000', '20261022-126000', '20261022', 'notes', '20261022-120000-backup']) {
    assert.equal(resultDate(name), null, name);
  }
  assert.equal(resultDate('20240229-120000').toISOString(), '2024-02-29T00:00:00.000Z');
});

test('keeps report and raw input history for a year, but expires only logs after three months', t => {
  const f = fixture(t);
  const files = { 'buildkit.json': '{}', 'gotestoutput.json': '{}', 'logs.tar.gz': 'archive', 'logs/build.log': 'log', 'env.txt': 'GITHUB_EVENT_NAME=schedule' };
  for (const name of ['20251021-120000', '20251022-120000', '20260721-120000', '20260722-120000', '20261022-120000']) f.add(name, files);
  const actions = planPruning(f.root, settings, now);
  assert.deepEqual(actions.map(action => action.name).sort(), [
    '20251021-120000',
    path.join('20251022-120000', 'logs'),
    path.join('20251022-120000', 'logs.tar.gz'),
    path.join('20260721-120000', 'logs'),
    path.join('20260721-120000', 'logs.tar.gz'),
  ].sort());
});

test('keeps all scheduled runs and the latest configured number of other runs', t => {
  const f = fixture(t);
  f.add('20261022-010000', { 'buildkit.json': '{"metadata":{"event":"schedule"}}' });
  f.add('20261022-020000', { 'buildx.json': '{"metadata":{"event":"schedule"}}' });
  f.add('20261022-030000', { 'env.txt': 'GITHUB_EVENT_NAME="schedule"\r\n' });
  f.add('20261022-040000', { 'env.txt': 'GITHUB_EVENT_NAME=schedule_extra' });
  f.add('20261022-050000', { 'buildkit.json': '{"metadata":{"event":"push"}}', 'env.txt': 'GITHUB_EVENT_NAME=schedule' });
  f.add('20261022-060000');
  f.add('20261021-060000');
  assert.deepEqual(planPruning(f.root, settings, now).map(a => a.name), ['20261022-050000', '20261022-040000']);
  assert.deepEqual(planPruning(f.root, { ...settings, keepResultsPerDay: 2 }, now).map(a => a.name), ['20261022-040000']);
});

test('unrecognized directories and manifest files are left alone', t => {
  const f = fixture(t);
  f.add('20260230-120000');
  f.add('backup');
  fs.writeFileSync(path.join(f.root, 'manifest.json'), '{}');
  assert.deepEqual(planPruning(f.root, settings, now), []);
});

test('CLI dry-run leaves files intact, and actual pruning preserves raw inputs', t => {
  const f = fixture(t);
  const name = new Date().toISOString().slice(0, 10).replaceAll('-', '') + '-000000';
  f.add(name, { 'buildkit.json': '{"metadata":{"event":"schedule"}}', 'gotestoutput.json': 'raw' });
  const older = cutoffDate(new Date(), 4).toISOString().slice(0, 10).replaceAll('-', '') + '-000000';
  f.add(older, { 'buildkit.json': '{}', 'gotestoutput.json': 'older raw', 'logs.tar.gz': 'archive', 'logs/build.log': 'log' });
  f.add('20000101-000000', { 'logs.tar.gz': 'old' });
  const run = (...args) => spawnSync(process.execPath, [path.join(__dirname, 'pruneResults.js'), ...args], {
    env: { ...process.env, RESULTS_DIR: f.root, MONTHS_TO_KEEP: '12', LOGS_MONTHS_TO_KEEP: '3', KEEP_RESULTS_PER_DAY: '1' },
    encoding: 'utf8',
  });
  const preview = run('--dry-run');
  assert.equal(preview.status, 0, preview.stderr);
  assert.match(preview.stdout, /Would remove .*20000101-000000/);
  assert.ok(fs.existsSync(path.join(f.root, '20000101-000000/logs.tar.gz')));
  assert.ok(fs.existsSync(path.join(f.root, older, 'logs/build.log')));
  const result = run();
  assert.equal(result.status, 0, result.stderr);
  assert.ok(!fs.existsSync(path.join(f.root, '20000101-000000')));
  assert.equal(fs.readFileSync(path.join(f.root, name, 'gotestoutput.json'), 'utf8'), 'raw');
  assert.equal(fs.readFileSync(path.join(f.root, older, 'gotestoutput.json'), 'utf8'), 'older raw');
  assert.ok(fs.existsSync(path.join(f.root, older, 'buildkit.json')));
  assert.ok(!fs.existsSync(path.join(f.root, older, 'logs.tar.gz')));
  assert.ok(!fs.existsSync(path.join(f.root, older, 'logs')));
});

test('malformed report metadata aborts before any removal', t => {
  const f = fixture(t);
  f.add('20000101-000000');
  const name = new Date().toISOString().slice(0, 10).replaceAll('-', '') + '-000000';
  f.add(name, { 'buildkit.json': 'invalid JSON' });
  const result = spawnSync(process.execPath, [path.join(__dirname, 'pruneResults.js')], {
    env: { ...process.env, RESULTS_DIR: f.root, MONTHS_TO_KEEP: '12', LOGS_MONTHS_TO_KEEP: '3', KEEP_RESULTS_PER_DAY: '1' },
    encoding: 'utf8',
  });
  assert.equal(result.status, 1);
  assert.ok(fs.existsSync(path.join(f.root, '20000101-000000')));
});

test('full-output sidecars follow log retention rather than report retention', t => {
  const f = fixture(t);
  const name = `output-${'a'.repeat(64)}.log.gz`;
  f.add('20260721-120000', { [name]: 'compressed output', 'gotestoutput.json': 'raw', 'buildkit.json': '{}' });
  f.add('20260722-120000', { [name]: 'retained on cutoff day', 'buildkit.json': '{}' });
  assert.deepEqual(planPruning(f.root, settings, now).map(a => a.name), [path.join('20260721-120000', name)]);
});
