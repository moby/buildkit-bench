import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import vm from 'node:vm';
import { markRaw } from 'vue';
import * as charts from '../src/charts/dashboard.mjs';
import * as benchmarks from '../src/reports/benchmarks.mjs';
import * as formatting from '../src/reports/format.mjs';
import { reportCharts } from '../src/charts/report.mjs';

const source = readFileSync(new URL('../src/components/HistoryPanel.vue', import.meta.url), 'utf8')
  .match(/<script>([\s\S]*?)<\/script>/)[1]
  .replace(/^import[\s\S]*?;$/gm, '')
  .replace('export default', 'component =');
const fixture = () => ({
  version: 1,
  project: 'buildkit',
  candidates: { buildkit: [{ name: 'master', sha: 'abc' }] },
  benchmarks: [
    {
      name: 'Simple',
      metrics: { duration: { unit: 's', description: 'Time (s)', chart: 'boxplot' } },
      samples: [{ buildkit: 'master', values: { duration: 2 } }],
    },
  ],
});
function panel(fetch) {
  const context = {
    ...charts,
    ...benchmarks,
    ...formatting,
    fetch,
    markRaw,
    reportCharts,
    AbortController,
    BenchmarkChart: {},
    UiIcon: {},
    process: { env: { BASE_URL: '/buildkit-bench/' } },
    resultsData: {
      results: ['20261013-101553', '20260612-124918', '20260315-043927'],
      jsonReports: { '20261013-101553': ['buildkit'], '20260612-124918': ['buildkit'] },
    },
  };
  vm.runInNewContext(source, context);
  const component = context.component;
  const report = fixture();
  const state = {
    ...component.data(),
    report,
    benchmark: report.benchmarks[0],
    metric: 'duration',
    theme: 'dark',
    run: '20261013-101553',
  };
  for (const [key, getter] of Object.entries(component.computed))
    Object.defineProperty(state, key, { get: () => getter.call(state) });
  state.loadHistory = component.methods.loadHistory.bind(state);
  return { state, component, resultsData: context.resultsData };
}

test('history uses shared metric and reference props without overwriting them on fallback', () => {
  const { state } = panel();
  state.benchmark.metrics.alloc = { unit: 'bytes', description: 'Memory', chart: 'boxplot' };
  state.benchmark.samples.push({ buildkit: 'other', values: { duration: 3, alloc: 100 } });
  state.report.candidates.buildkit.push({ name: 'other' });
  state.metric = 'alloc';
  state.reference = 'other';
  assert.equal(state.selectedMetric, 'alloc');
  assert.equal(state.selectedReference, 'other');
  state.metric = 'missing';
  state.reference = 'missing';
  assert.equal(state.selectedMetric, 'duration');
  assert.equal(state.selectedReference, 'master');
  assert.equal(state.reference, 'missing');
});

test('history excludes HTML-only runs and keeps partial failures visible with Pages paths', async () => {
  const calls = [];
  const { state } = panel(async (url) => {
    calls.push(url);
    return { ok: false, status: 404 };
  });
  await state.loadHistory();
  assert.deepEqual(calls, ['/buildkit-bench/result/20260612-124918/buildkit.json']);
  assert.equal(state.points.length, 2);
  assert.equal(state.points[0].problem, 'Request failed');
  assert.equal(state.validPoints, 1);
  assert.equal(state.reviewed, false);
  assert.equal(state.option.series[0].connectNulls, false);
  assert.equal(state.option.series[0].data[0][1], null);
});

test('history rejects project/schema mismatches and does not compare changed units', async () => {
  for (const change of [{ project: 'buildx' }, { version: 999 }]) {
    const { state } = panel(async () => ({ ok: true, json: async () => ({ ...fixture(), ...change }) }));
    await state.loadHistory();
    assert.equal(state.points[0].problem, 'Request failed');
  }
  const older = fixture();
  older.benchmarks[0].metrics.duration.unit = 'ms';
  const { state } = panel(async () => ({ ok: true, json: async () => older }));
  await state.loadHistory();
  assert.equal(state.points[0].problem, 'Metric unavailable or unit changed');
  assert.equal(state.points[0].row, undefined);
});

test('a superseded history request cannot replace the new run', async () => {
  const pending = [];
  const { state } = panel((url, { signal }) => new Promise((resolve) => pending.push({ resolve, signal })));
  const first = state.loadHistory();
  state.run = '20260612-124918';
  await state.loadHistory();
  assert.equal(pending[0].signal.aborted, true);
  pending[0].resolve({ ok: true, json: async () => fixture() });
  await first;
  assert.equal(state.entries.length, 1);
  assert.equal(state.entries[0].name, '20260612-124918');
  assert.equal(state.loading, false);
});

test('history limits the window to six JSON runs and changes require context review again', async () => {
  const { state, component, resultsData } = panel(async () => ({ ok: true, json: async () => fixture() }));
  resultsData.results = Array.from({ length: 9 }, (_, i) => `2026100${i + 1}-100000`);
  resultsData.jsonReports = Object.fromEntries(resultsData.results.map((name) => [name, ['buildkit']]));
  await state.loadHistory();
  assert.equal(state.entries.length, 6);
  assert.equal(state.entries[0].name, '20261004-100000');
  for (const key of ['selectedMetric', 'benchmark.name']) {
    state.reviewed = true;
    component.watch[key].call(state);
    assert.equal(state.reviewed, false);
  }
  state.reviewed = true;
  component.watch.selectedReference.handler.call(state, 'different');
  assert.equal(state.reviewed, false);
  assert.equal(state.reference, undefined);
});
