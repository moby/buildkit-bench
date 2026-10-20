import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { formatValue, runTimestamp } from '../src/reports/format.mjs';
import { candidateRows, percentChange, comparisonContext, compareContexts } from '../src/reports/benchmarks.mjs';
import { dashboardChart } from '../src/charts/dashboard.mjs';
import { reportCharts } from '../src/charts/report.mjs';

const historical = JSON.parse(readFileSync(new URL('./fixtures/historical-charts.json', import.meta.url)));
for (const { report, expected } of historical) {
  test(`${report.project} dashboard medians and distributions retain historical statistics`, () => {
    const benchmark = report.benchmarks[0];
    const rows = candidateRows(report, benchmark, 'duration');
    assert.deepEqual(
      rows.map((row) => row.ref),
      expected.refs
    );
    assert.deepEqual(
      rows.map((row) => [row.min, row.q1, row.median, row.q3, row.max]),
      expected.series[0].data
    );
    const filtered = {
      ...benchmark,
      samples: benchmark.samples.map((s) => ({ ...s, values: { duration: s.values.duration } })),
    };
    const original = reportCharts({ ...report, benchmarks: [filtered] })[0].option;
    const snapshot = structuredClone(original);
    const median = dashboardChart(original, 'dark', rows[1], 'medians', rows);
    const distribution = dashboardChart(original, 'light', rows[0], 'distribution', rows);
    assert.deepEqual(
      median.series[0].data.map((d) => d.value),
      rows.map((row) => row.median)
    );
    assert.deepEqual(distribution.series[0].data, original.series[0].data);
    assert.equal(median.series[0].markLine.data.at(-1).yAxis, rows[1].median);
    assert.deepEqual(original, snapshot);
  });
}

test('baseline deltas preserve precision and zero baselines are undefined', () => {
  assert.equal(percentChange(90, 100), -10);
  assert.equal(percentChange(110, 100), 10);
  assert.equal(percentChange(0, 100), -100);
  assert.equal(percentChange(0, 0), null);
  assert.equal(percentChange(1, 0), null);
  assert.equal(percentChange(1, undefined), null);
  assert.ok(Math.abs(percentChange(0.0002, 0.0001) - 100) < 1e-10);
});

function fixture() {
  const execution = {
    go_version: 'go1.26',
    num_cpu: 4,
    gomaxprocs: 4,
    settings: { parallel: 1 },
    environment: { TEST_IMAGE: 'image:tag', GITHUB_RUN_ID: '123' },
    binaries: [{ project: 'buildx', ref: 'stable', name: 'buildx', sha256: 'abc' }],
  };
  const report = {
    project: 'buildkit',
    sources: { one: { execution } },
    generator: { go_version: 'go9.99' },
  };
  const benchmark = {
    samples: [
      {
        buildkit: 'master',
        buildx: 'stable',
        source_id: 'one',
        cpu: 'CPU A',
        os: 'linux',
        architecture: 'amd64',
        iterations: 5,
      },
    ],
  };
  return { report, benchmark };
}

test('history context compares execution settings, runtime and counterpart binaries', () => {
  const { report, benchmark } = fixture();
  const current = comparisonContext(report, benchmark, 'master');
  assert.equal(compareContexts(current, current).label, 'Recorded fields match');
  report.sources.one.execution.go_version = 'go1.25';
  report.sources.one.execution.binaries[0].sha256 = 'def';
  const result = compareContexts(current, comparisonContext(report, benchmark, 'master'));
  assert.deepEqual(result.changed, ['Go runtime', 'Counterpart binaries']);
  assert.equal(result.label, 'Recorded differences');
});

test('unknown execution metadata never inherits the exporter runtime or claims a match', () => {
  const { report, benchmark } = fixture();
  delete report.sources.one.execution;
  const context = comparisonContext(report, benchmark, 'master');
  assert.deepEqual(context['Go runtime'], ['Not recorded']);
  assert.equal(compareContexts(context, context).label, 'Metadata incomplete');
  assert.equal(compareContexts({}, {}).label, 'Metadata incomplete');
});

test('mixed sources retain missing fields, but CI run IDs do not imply environment changes', () => {
  const { report, benchmark } = fixture();
  const current = comparisonContext(report, benchmark, 'master');
  report.sources.one.execution.environment.GITHUB_RUN_ID = '456';
  assert.deepEqual(comparisonContext(report, benchmark, 'master'), current);
  benchmark.samples.push({ ...benchmark.samples[0], source_id: 'missing' });
  const mixed = comparisonContext(report, benchmark, 'master');
  assert.ok(compareContexts(current, mixed).missing.includes('Go runtime'));
  assert.deepEqual(mixed['Go runtime'], ['Not recorded', 'go1.26']);
});

test('history uses elapsed dates rather than equally spaced categories', () => {
  assert.equal(runTimestamp('20261013-101553') - runTimestamp('20261012-101553'), 86400000);
  assert.equal(runTimestamp('unknown'), null);
});

test('environment changes identify individual fields, including the suite revision', () => {
  const { report, benchmark } = fixture();
  report.sources.one.execution.environment.suite_revision = 'aaa';
  const current = comparisonContext(report, benchmark, 'master');
  report.sources.one.execution.environment.suite_revision = 'bbb';
  report.sources.one.execution.environment.TEST_IMAGE = 'image:new';
  const previous = comparisonContext(report, benchmark, 'master');
  assert.deepEqual(compareContexts(current, previous).changed, ['Environment: TEST_IMAGE', 'Suite revision']);
  assert.deepEqual(current['Environment: TEST_IMAGE'], ['image:tag']);
  assert.deepEqual(previous['Environment: TEST_IMAGE'], ['image:new']);
});

test('display precision distinguishes sub-millisecond changes and formats tooltip units', () => {
  assert.notEqual(formatValue(0.0309, 's'), formatValue(0.0313, 's'));
  assert.equal(formatValue(1024, 'bytes'), '1 KiB');
  const report = historical[0].report;
  const rows = candidateRows(report, report.benchmarks[0], 'duration');
  const option = dashboardChart(reportCharts(report)[0].option, 'dark', rows[0], 'medians', rows, 's');
  assert.equal(option.tooltip.valueFormatter(0.0309), '0.0309 s');
  assert.equal(option.yAxis[0].axisLabel.formatter(1), '1 s');
});

test('box plot tooltips retain the candidate heading and format all five statistics', async () => {
  const echarts = await import('echarts');
  const report = structuredClone(historical[0].report);
  report.benchmarks[0].samples = report.benchmarks[0].samples.map((sample) => ({
    ...sample,
    values: { duration: sample.values.duration },
  }));
  const rows = candidateRows(report, report.benchmarks[0], 'duration');
  const option = dashboardChart(reportCharts(report)[0].option, 'light', rows[0], 'distribution', rows, 's');
  const chart = echarts.init(null, null, { renderer: 'svg', ssr: true, width: 1000, height: 390 });
  try {
    chart.setOption(option);
    const tooltip = chart.getModel().getSeriesByIndex(0).formatTooltip(0, false);
    assert.equal(tooltip.blocks[0].name, rows[0].ref);
    assert.deepEqual(option.tooltip.valueFormatter(tooltip.blocks[0].value), []);
    assert.equal(tooltip.blocks.length, 6);
    assert.deepEqual(option.tooltip.valueFormatter([1, 2, 3, 4, 5]), ['1 s', '2 s', '3 s', '4 s', '5 s']);
    assert.equal(option.tooltip.valueFormatter(tooltip.blocks[3].value), formatValue(rows[0].median, 's'));
  } finally {
    chart.dispose();
  }
});

test('explicitly empty execution flags stay distinct from unrecorded flags', () => {
  const { report, benchmark } = fixture();
  report.sources.one.execution.settings['test.cpu'] = '';
  const context = comparisonContext(report, benchmark, 'master');
  assert.deepEqual(context['Setting: test.cpu'], ['(empty)']);
  assert.equal(compareContexts(context, context).label, 'Recorded fields match');
  delete report.sources.one.execution.settings['test.cpu'];
  assert.ok(
    compareContexts(context, comparisonContext(report, benchmark, 'master')).missing.includes(
      'Setting: test.cpu'
    )
  );
});
