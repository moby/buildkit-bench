import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { boxPlot, reportCharts } from '../src/charts/report.mjs';

// Raw samples and expected values extracted independently from the JSON and
// original go-echarts HTML for the same historical run (three candidates each).
const historical = JSON.parse(readFileSync(new URL('./fixtures/historical-charts.json', import.meta.url)));
for (const { source, report, expected } of historical) {
  test(`matches original chart statistics and ordering from ${source}`, () => {
    const [{ option }] = reportCharts(report);
    assert.deepEqual(option.title, expected.title);
    assert.deepEqual(option.xAxis.data, expected.refs);
    assert.deepEqual(option.series.map(s => ({
      name: s.name, type: s.type, yAxisIndex: s.yAxisIndex, data: s.data.map(d => d.value),
    })), expected.series);
  });
}

// The historical configuration only has overlays. Generate these expectations
// with the Go exporter over the same real samples, changing duration to a bar
// with an average line so allocation remains a standalone box plot.
const references = JSON.parse(readFileSync(new URL('./fixtures/go-reference-charts.json', import.meta.url)));
for (const { report, expected } of references) {
  test(`${report.project} standalone boxes and bars match Go-generated reference charts`, () => {
    assert.deepEqual(reportCharts(report).map(({ option }) => ({
      title: option.title, refs: option.xAxis.data,
      series: option.series.map(s => ({
        name: s.name, type: s.type, yAxisIndex: s.yAxisIndex,
        data: s.data.map(d => d.value), ...(s.markLine ? { markLine: s.markLine } : {}),
      })),
    })), expected);
  });
}

test('box plots use median halves and min/max whiskers without mutating samples', () => {
  const values = [100, 4, 2, 1, 3];
  assert.deepEqual(boxPlot(values), [1, 1.5, 3, 52, 100]);
  assert.deepEqual(values, [100, 4, 2, 1, 3]);
  assert.deepEqual(boxPlot([1, 2, 3, 4, 5, 6]), [1, 2, 3.5, 5, 6]);
  assert.deepEqual(boxPlot([2, 8]), [2, 2, 5, 8, 8]);
  assert.deepEqual(boxPlot([0]), [0, 0, 0, 0, 0]);
});

function fixture(project = 'buildkit') {
  return {
    version: 1, project,
    candidates: { [project]: [{ name: 'z' }, { name: 'a' }, { name: 'absent' }] },
    benchmarks: [{
      name: 'BenchmarkExample', description: 'Example',
      metrics: {
        duration: { chart: 'boxplot', description: 'Time (s)' },
        alloc: { chart: 'boxplot', description: 'Allocated memory (bytes)' },
        custom: { chart: 'bar', description: 'Custom', average: true },
      },
      samples: [
        { [project]: 'a', values: { duration: 4, alloc: 40, custom: -1.234565 } },
        { [project]: 'z', values: { duration: 1, alloc: 10, custom: 1.234565 } },
        { [project]: 'a', values: { duration: 6, alloc: 60, custom: -1.234565 } },
        { [project]: 'z', values: { duration: 3, alloc: 30, custom: 1.234565 } },
        { [project]: 'unknown', values: { duration: 9, alloc: 90, custom: 9 } },
      ],
    }],
  };
}

for (const project of ['buildkit', 'buildx']) {
  test(`${project} preserves candidate order, overlay axes, labels and rounded bar medians`, () => {
    const [overlay, bar] = reportCharts(fixture(project));
    assert.deepEqual(overlay.option.xAxis.data, ['z', 'a']);
    assert.deepEqual(overlay.option.series.map(s => [s.name, s.yAxisIndex]), [
      ['Time (s)', 0], ['Allocated memory (bytes)', 1],
    ]);
    assert.deepEqual(overlay.option.series[0].data[0].value, [1, 1, 2, 3, 3]);
    assert.deepEqual(overlay.option.series[1].data[0].value, [10, 10, 20, 30, 30]);
    assert.deepEqual(bar.option.series[0].data.map(d => d.value), [1.23456, -1.23456]);
    assert.equal(bar.option.series[0].markLine.data[0].type, 'average');
  });
}

test('without candidates, references sort by name; overlays only use shared references', () => {
  const report = fixture();
  report.candidates = {};
  delete report.benchmarks[0].samples[0].values.alloc;
  delete report.benchmarks[0].samples[2].values.alloc;
  const [overlay, bar] = reportCharts(report);
  assert.deepEqual(overlay.option.xAxis.data, ['unknown', 'z']);
  assert.deepEqual(bar.option.xAxis.data, ['a', 'unknown', 'z']);
});

test('bar rounding uses Go ties away from zero', () => {
  const report = fixture();
  for (const sample of report.benchmarks[0].samples) {
    sample.values.custom = sample.buildkit === 'a' ? -0.000005 : 0.000005;
  }
  assert.deepEqual(reportCharts(report)[1].option.series[0].data.map(d => d.value), [0.00001, -0.00001]);
});

test('standalone metrics, missing metrics and empty reports', () => {
  const report = fixture();
  report.benchmarks[0].samples.forEach(sample => delete sample.values.alloc);
  const charts = reportCharts(report);
  assert.deepEqual(charts.map(c => c.option.series[0].type), ['bar', 'boxplot']);
  assert.deepEqual(reportCharts({ ...report, benchmarks: [] }), []);
});

test('rejects incompatible or invalid reports instead of drawing misleading charts', () => {
  assert.throws(() => reportCharts({ ...fixture(), version: 2 }), /Unsupported report version/);
  const report = fixture();
  report.benchmarks[0].samples[0].values.duration = null;
  assert.throws(() => reportCharts(report), /Invalid duration/);
  report.benchmarks[0].samples[0].values.duration = 1;
  report.benchmarks[0].metrics.custom.chart = 'unknown';
  assert.throws(() => reportCharts(report), /Unknown chart type/);
});
