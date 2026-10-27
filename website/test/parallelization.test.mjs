import assert from 'node:assert/strict';
import test from 'node:test';
import vm from 'node:vm';
import { readFileSync } from 'node:fs';
import { benchmarkNavigation, variantRows } from '../src/reports/benchmarks.mjs';
import { chartPalette } from '../src/charts/dashboard.mjs';
import { formatValue } from '../src/reports/format.mjs';

function fixture() {
  return {
    project: 'buildkit',
    candidates: { buildkit: [{ name: 'stable' }, { name: 'master' }, { name: 'older' }] },
    benchmarks: [128, 16, 64, 32].map((n) => ({
      name: `BenchmarkBuild/BenchmarkBuildHighParallelization${n}x`,
      description: `Parallel ${n}x`,
      metrics: {
        duration: { unit: 's', chart: 'boxplot', description: 'Time (s)' },
        alloc: { unit: 'bytes', chart: 'boxplot', description: 'Memory' },
      },
      samples: [
        { buildkit: 'master', values: { duration: n, alloc: n * 100 } },
        { buildkit: 'stable', values: { duration: n * 2, alloc: n * 200 } },
        { buildkit: 'older', values: { duration: n * 3 } },
      ],
    })),
  };
}

test('parallelism navigation groups variants numerically without changing source benchmarks', () => {
  const report = fixture();
  const original = structuredClone(report);
  const ordinary = { name: 'Other', description: 'Other' };
  const entries = benchmarkNavigation([ordinary, ...report.benchmarks]);
  assert.equal(entries.length, 2);
  assert.equal(entries[0], ordinary);
  assert.deepEqual(
    entries[1].variants.map((v) => v.parallelism),
    [16, 32, 64, 128]
  );
  assert.equal(entries[1].variants[0].benchmark, report.benchmarks[1]);
  assert.deepEqual(report, original);
});

test('parallelism medians retain candidate order, missing points and unit boundaries', () => {
  const report = fixture();
  report.benchmarks.find((b) => b.name.endsWith('32x')).samples = [
    { buildkit: 'stable', values: { duration: 64 } },
  ];
  report.benchmarks.find((b) => b.name.endsWith('64x')).metrics.duration.unit = 'ms';
  const variants = benchmarkNavigation(report.benchmarks)[0].variants;
  const rows = variantRows(report, variants, 'duration', 's');
  assert.deepEqual(
    rows.map((row) => row.ref),
    ['stable', 'master', 'older']
  );
  assert.deepEqual(
    rows[1].points.map((p) => p?.median ?? null),
    [16, null, null, 128]
  );
  assert.deepEqual(
    rows[0].points.map((p) => p?.median ?? null),
    [32, 64, null, 256]
  );
});

const source = readFileSync(new URL('../src/components/BenchmarkVariantsChart.vue', import.meta.url), 'utf8')
  .match(/<script>([\s\S]*?)<\/script>/)[1]
  .replace(/^import[\s\S]*?;$/gm, '')
  .replace('export default', 'component =');
function chart(report = fixture()) {
  const context = { BenchmarkChart: {}, variantRows, chartPalette, formatValue };
  vm.runInNewContext(source, context);
  const entry = benchmarkNavigation(report.benchmarks)[0];
  const variants = entry.variants;
  const component = context.component;
  const state = {
    ...component.data(),
    report,
    variants,
    kind: entry.kind,
    metrics: variants[0].benchmark.metrics,
    metric: 'duration',
    theme: 'dark',
    baselineRef: '',
  };
  for (const [name, getter] of Object.entries(component.computed))
    Object.defineProperty(state, name, { get: () => getter.call(state) });
  return { state, component };
}

test('grouped chart starts with baseline and master and allows extra candidate selection', () => {
  const { state, component } = chart();
  assert.equal(state.baseline, 'stable');
  assert.deepEqual(Array.from(state.metricKeys), ['duration', 'alloc']);
  assert.deepEqual(Array.from(state.selected), ['stable', 'master']);
  assert.deepEqual(Array.from(state.option.xAxis.data), ['16x', '32x', '64x', '128x']);
  assert.deepEqual(Array.from(state.option.series[1].data), [16, 32, 64, 128]);
  assert.equal(state.option.series[0].lineStyle.type, 'dashed');
  assert.equal(state.option.series[0].connectNulls, false);
  component.methods.toggleCandidate.call(state, 'older');
  assert.equal(state.option.series.length, 3);
  component.methods.toggleCandidate.call(state, 'master');
  assert.deepEqual(Array.from(state.selected), ['stable', 'older']);
  state.baselineRef = 'master';
  assert.ok(state.selected.includes('master'));
  state.metric = 'alloc';
  assert.equal(state.option.tooltip.valueFormatter(1024), '1 KiB');
  assert.equal(state.option.series.length, 2);
});

function emulatorFixture() {
  const report = fixture();
  report.benchmarks = report.benchmarks.slice(0, 2).map((benchmark, i) => ({
    ...benchmark,
    name: `BenchmarkBuild/BenchmarkBuildEmulator${i ? 'Native' : ''}`,
  }));
  return report;
}

test('emulator grouping requires both workloads and preserves historical single benchmarks', () => {
  const report = emulatorFixture();
  const original = structuredClone(report);
  const [entry] = benchmarkNavigation(report.benchmarks);
  assert.equal(entry.kind, 'emulator');
  assert.deepEqual(entry.variants.map(v => v.label), ['Native', 'Emulated']);
  assert.equal(entry.variants[0].benchmark, report.benchmarks[1]);
  assert.deepEqual(report, original);
  for (const benchmark of report.benchmarks)
    assert.equal(benchmarkNavigation([benchmark])[0], benchmark);
});

test('emulator chart compares medians as bars and leaves missing candidate values empty', () => {
  const report = emulatorFixture();
  report.benchmarks[1].samples = report.benchmarks[1].samples.filter(s => s.buildkit !== 'master');
  const { state } = chart(report);
  assert.deepEqual(Array.from(state.option.xAxis.data), ['Native', 'Emulated']);
  assert.equal(state.option.xAxis.name, 'Execution');
  assert.equal(state.option.series[0].type, 'bar');
  assert.deepEqual(Array.from(state.option.series[1].data), [null, 128]);
  state.metric = 'alloc';
  assert.deepEqual(Array.from(state.option.series[0].data), [3200, 25600]);
});
