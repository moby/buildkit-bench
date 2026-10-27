import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import vm from 'node:vm';
import { createSSRApp, markRaw } from 'vue';
import { renderToString } from 'vue/server-renderer';
import { reportCharts } from '../src/charts/report.mjs';
import * as charts from '../src/charts/dashboard.mjs';
import * as benchmarks from '../src/reports/benchmarks.mjs';
import * as formatting from '../src/reports/format.mjs';
import { createRouter, createMemoryHistory } from 'vue-router';

// Exercise the component's actual loading methods/computed properties without
// requiring a DOM or mounting ECharts in Node.
const source = readFileSync(new URL('../src/components/ResultView.vue', import.meta.url), 'utf8')
  .match(/<script>([\s\S]*?)<\/script>/)[1]
  .replace(/^import[\s\S]*?;$/gm, '')
  .replace(/\bimport\(/g, 'loadMetadata(')
  .replace('export default', 'component =');

function view(fetch, loadMetadata) {
  const context = {
    fetch,
    loadMetadata,
    markRaw,
    reportCharts,
    AbortController,
    ...charts,
    ...benchmarks,
    ...formatting,
    BenchmarkChart: {},
    BenchmarkVariantsChart: {},
    ResultsList: {},
    RunDetails: {},
    HistoryPanel: {},
    RunOverview: {},
    UiIcon: {},
    require: (path) => path,
    process: { env: { BASE_URL: '/buildkit-bench/' } },
    resultsData: {
      results: ['both'],
      logArchives: ['both'],
      jsonReports: { both: ['buildkit', 'buildx'], kit: ['buildkit'], legacy: [] },
      htmlReports: {
        both: { buildkit: 'buildkit.html', buildx: 'buildx.html' },
        legacy: { buildkit: 'index.html' },
      },
    },
  };
  vm.runInNewContext(source, context);
  const component = context.component;
  const state = { ...component.data(), $route: { params: { result: 'both' }, query: {} } };
  state.$router = { push(location) { state.$route = location; return Promise.resolve(); } };
  for (const [key, getter] of Object.entries(component.computed)) {
    Object.defineProperty(state, key, typeof getter === 'function'
      ? { get: () => getter.call(state) }
      : { get: () => getter.get.call(state), set: value => getter.set.call(state, value) });
  }
  for (const [key, method] of Object.entries(component.methods)) state[key] = method.bind(state);
  return { state, component };
}

const report = (project) => ({ version: 1, project, candidates: {}, benchmarks: [] });
const response = (project) => ({ ok: true, json: async () => report(project) });

test('emulator permalinks retain their workload when comparison controls change', async () => {
  const { state } = view();
  state.report = {
    ...report('buildkit'),
    benchmarks: ['', 'Native'].map(suffix => ({
      name: `BenchmarkBuild/BenchmarkBuildEmulator${suffix}`,
      metrics: { duration: { unit: 's', chart: 'boxplot' } },
      samples: [{ buildkit: 'master', values: { duration: 1 } }],
    })),
  };
  for (const benchmark of state.report.benchmarks) {
    state.$route.query = { benchmark: benchmark.name };
    assert.equal(state.activeEntry.kind, 'emulator');
    assert.equal(state.activeBenchmark.name, benchmark.name);
    await state.navigateSelection({ metric: 'duration', baseline: 'master', chart: 'medians' });
    assert.equal(state.activeBenchmark.name, benchmark.name);
    assert.equal(state.$route.query.variant, benchmark.name);
  }
  state.variantName = state.report.benchmarks[0].name;
  assert.equal(state.activeBenchmark.name, state.report.benchmarks[0].name);
  state.search = 'native';
  assert.equal(state.filteredBenchmarks.length, 1);
});

function permalinkReport(project = 'buildx') {
  return {
    ...report(project),
    benchmarks: ['Simple', 'Other'].map(name => ({
      name: `BenchmarkBuild/BenchmarkBuild${name}`,
      metrics: {
        duration: { unit: 's', chart: 'boxplot' },
        alloc: { unit: 'bytes', chart: 'boxplot' },
      },
      samples: [{ buildkit: 'master', buildx: 'master', values: { duration: 2, alloc: 100 } }],
    })),
  };
}

test('a shared query restores project, benchmark, and metric after loading the report', async () => {
  const { state } = view(async url => {
    assert.equal(url, '/buildkit-bench/result/both/buildx.json');
    return { ok: true, json: async () => permalinkReport() };
  });
  state.$route.query = {
    project: 'buildx', benchmark: 'BenchmarkBuild/BenchmarkBuildOther', metric: 'alloc',
    chart: 'distribution', baseline: 'master', view: 'history', reference: 'master',
  };
  await state.loadReport();
  assert.equal(state.selectedProject, 'buildx');
  assert.equal(state.activeBenchmark.name, 'BenchmarkBuild/BenchmarkBuildOther');
  assert.equal(state.selectedMetric, 'alloc');
  assert.equal(state.chartMode, 'distribution');
  assert.equal(state.baselineRow.ref, 'master');
  assert.equal(state.view, 'history');
  assert.equal(state.historyRef, 'master');
});

test('selection navigation and router Back/Forward restore the displayed view', async () => {
  const { state } = view();
  const router = createRouter({ history: createMemoryHistory(), routes: [
    { name: 'ResultView', path: '/result/:result', component: {} },
  ] });
  await router.push('/result/both');
  state.$router = router;
  Object.defineProperty(state, '$route', { get: () => router.currentRoute.value });
  state.report = permalinkReport('buildkit');
  await state.selectBenchmark(state.navigation.find(b => b.name.endsWith('Other')));
  assert.equal(router.currentRoute.value.query.project, 'buildkit');
  assert.equal(router.currentRoute.value.query.benchmark, 'BenchmarkBuild/BenchmarkBuildOther');
  await state.navigateSelection({ metric: 'alloc', baseline: 'feature/a+b & c' });
  const shared = router.currentRoute.value.fullPath;
  assert.equal(state.selectedMetric, 'alloc');
  const traverse = direction => new Promise(resolve => {
    const stop = router.afterEach(() => { stop(); resolve(); });
    router.go(direction);
  });
  await traverse(-1);
  assert.equal(state.selectedMetric, 'duration');
  await traverse(1);
  assert.equal(state.selectedMetric, 'alloc');
  await router.push('/result/both');
  assert.equal(state.activeBenchmark.name, 'BenchmarkBuild/BenchmarkBuildSimple');
  await router.push(shared);
  assert.equal(state.activeBenchmark.name, 'BenchmarkBuild/BenchmarkBuildOther');
  assert.equal(state.baselineRef, 'feature/a+b & c');
});

test('unknown and repeated query values fall back to available selections', () => {
  const { state } = view();
  state.report = permalinkReport('buildkit');
  state.$route.query = { project: 'missing', benchmark: 'missing', metric: 'missing', chart: 'missing', view: 'missing' };
  assert.equal(state.selectedProject, 'buildkit');
  assert.equal(state.activeBenchmark.name, 'BenchmarkBuild/BenchmarkBuildSimple');
  assert.equal(state.selectedMetric, 'duration');
  assert.equal(state.chartMode, 'medians');
  assert.equal(state.view, 'benchmarks');
  state.$route.query = { project: ['buildx', 'buildkit'], benchmark: null, metric: ['alloc', 'duration'] };
  assert.equal(state.selectedProject, 'buildkit');
  assert.equal(state.selectedMetric, 'duration');
});

test('overview resolves the latest run and keeps its route when switching projects', async () => {
  const { state } = view();
  state.$route = { name: 'Overview', params: {}, query: {} };
  assert.equal(state.currentRun, 'both');
  state.report = permalinkReport('buildkit');
  assert.equal(state.view, 'overview');
  await state.navigateSelection({ project: 'buildx' });
  assert.equal(state.view, 'overview');
  await state.selectBenchmark(state.navigation[0]);
  assert.equal(state.view, 'benchmarks');
  state.$route.query = { benchmark: state.navigation[0].name };
  assert.equal(state.view, 'benchmarks');
});

test('overview evidence links focus the selected pair and can restore the full chart', async () => {
  const { state } = view();
  state.report = permalinkReport('buildkit');
  state.report.benchmarks[0].samples = Array.from({ length: 8 }, (_, i) => ({ buildkit: `ref-${i}`, values: { duration: i + 1 } }));
  state.$route.query = { baseline: 'ref-3', candidate: 'ref-4', metric: 'duration', view: 'benchmarks' };
  assert.equal(state.focusedRow.ref, 'ref-4');
  assert.equal(state.chartOption.dataZoom[0].startValue, 2);
  assert.equal(state.chartOption.dataZoom[0].endValue, 5);
  await state.navigateSelection({ candidate: undefined });
  assert.equal(state.focusedRow, undefined);
  assert.equal(state.chartOption.dataZoom[0].startValue, undefined);
});

test('overview routes distinguish latest and historical runs and preserve benchmark navigation', async () => {
  const { state } = view();
  const router = createRouter({ history: createMemoryHistory(), routes: [
    { name: 'Overview', path: '/overview/:result?', component: {} },
    { name: 'ResultView', path: '/result/:result', component: {} },
  ] });
  await router.push('/overview?project=buildx');
  state.$router = router;
  Object.defineProperty(state, '$route', { get: () => router.currentRoute.value });
  state.report = permalinkReport();
  assert.equal(state.currentRun, 'both');
  assert.equal(state.reportUrl, '/buildkit-bench/result/both/buildx.json');
  await state.navigateSelection({ project: 'buildkit' });
  assert.equal(router.currentRoute.value.path, '/overview');
  await state.selectBenchmark(state.navigation[0]);
  assert.equal(router.currentRoute.value.path, '/result/both');
  assert.equal(state.view, 'benchmarks');
  await state.navigateSelection({ view: 'overview' });
  assert.equal(router.currentRoute.value.path, '/overview');
  assert.equal(router.currentRoute.value.query.benchmark, undefined);
  await router.push('/overview/kit');
  assert.equal(state.currentRun, 'kit');
  assert.equal(state.reportUrl, '/buildkit-bench/result/kit/buildkit.json');
  await state.selectBenchmark(state.navigation[0]);
  assert.equal(router.currentRoute.value.path, '/result/kit');
  await state.navigateSelection({ view: 'overview' });
  assert.equal(router.currentRoute.value.path, '/overview/kit');
});

test('project selection falls back for single-project/legacy runs and preserves Pages paths', () => {
  const { state } = view();
  state.activeTab = 'buildx';
  assert.equal(state.reportUrl, '/buildkit-bench/result/both/buildx.json');
  state.$route.params.result = 'kit';
  assert.equal(state.selectedProject, 'buildkit');
  state.$route.params.result = 'legacy';
  assert.equal(state.reportUrl, '');
  assert.equal(state.originalUrl, '/buildkit-bench/result/legacy/index.html');
});

test('JSON-only runs load without an HTML report', async () => {
  const { state } = view(() => Promise.resolve(response('buildkit')));
  state.$route.params.result = 'kit';
  assert.equal(state.originalUrl, '');
  assert.equal(state.reportUrl, '/buildkit-bench/result/kit/buildkit.json');
  await state.loadReport();
  assert.equal(state.error, '');
  assert.equal(state.report.project, 'buildkit');
});

test('switching projects cancels the old request and ignores late results', async () => {
  const pending = [];
  const { state } = view(
    (url, { signal }) => new Promise((resolve) => pending.push({ url, signal, resolve }))
  );
  const first = state.loadReport();
  assert.equal(state.loading, true);
  state.activeTab = 'buildx';
  const second = state.loadReport();
  assert.equal(pending[0].signal.aborted, true);
  pending[1].resolve(response('buildx'));
  await second;
  pending[0].resolve({ ok: true, json: async () => ({ version: 999 }) });
  await first;
  assert.equal(state.error, '');
  assert.equal(state.loading, false);
});

test('HTTP failures and invalid schemas are visible and retry can recover', async () => {
  const responses = [
    { ok: false, status: 404 },
    { ok: true, json: async () => ({ ...report('buildkit'), version: 999 }) },
    response('buildkit'),
  ];
  const { state } = view(async () => responses.shift());
  await state.loadReport();
  assert.match(state.error, /404/);
  assert.equal(state.loading, false);
  await state.loadReport();
  assert.match(state.error, /Unsupported report version/);
  await state.loadReport();
  assert.equal(state.error, '');
});

test('legacy navigation clears charts and unmount cancels in-flight loading', async () => {
  const { state, component } = view(async () => response('buildkit'));
  await state.loadReport();
  const request = state.request;
  assert.ok(state.report);
  state.$route.params.result = 'legacy';
  await state.loadReport();
  assert.equal(request.signal.aborted, true);
  assert.equal(state.report, null);
  assert.equal(state.loading, false);
  component.beforeUnmount.call(state);
  assert.equal(state.request.signal.aborted, true);
});

test('navigating X to Y to X does not append metadata links twice', async () => {
  const pending = [];
  const { state, component } = view(undefined, (path) =>
    path.endsWith('env.txt')
      ? Promise.resolve({ default: 'GITHUB_REPOSITORY=moby/buildkit-bench\nGITHUB_SHA=123456789\n' })
      : new Promise((resolve) => pending.push(resolve))
  );
  state.$route.params.result = 'legacy';
  const first = component.methods.updateMetadata.call(state);
  state.$route.params.result = 'legacy2';
  const second = component.methods.updateMetadata.call(state);
  state.$route.params.result = 'legacy';
  const third = component.methods.updateMetadata.call(state);
  pending[2]({});
  await third;
  assert.equal(state.metadataLinks.length, 1);
  pending[0]({});
  pending[1]({});
  await Promise.all([first, second]);
  assert.equal(state.metadataLinks.length, 1);
});

test('baseline selection survives a benchmark where that candidate is absent', () => {
  const { state } = view();
  state.report = {
    version: 1,
    project: 'buildkit',
    candidates: { buildkit: [{ name: 'a' }, { name: 'b' }] },
    benchmarks: [
      {
        name: 'one',
        metrics: { duration: { chart: 'boxplot', description: 'Time', unit: 's' } },
        samples: [
          { buildkit: 'a', values: { duration: 4 } },
          { buildkit: 'b', values: { duration: 2 } },
        ],
      },
      {
        name: 'two',
        metrics: { duration: { chart: 'boxplot', description: 'Time', unit: 's' } },
        samples: [{ buildkit: 'a', values: { duration: 3 } }],
      },
    ],
  };
  state.baselineRef = 'b';
  assert.equal(state.baselineRow.ref, 'b');
  state.benchmarkName = 'two';
  assert.equal(state.baselineRow.ref, 'a');
  assert.equal(state.baselineRef, 'b');
  state.benchmarkName = 'one';
  assert.equal(state.baselineRow.ref, 'b');
});

test('explorer searches, selects one metric, and colors only known lower-is-better deltas', () => {
  const { state, component } = view();
  const sample = { buildkit: 'master', values: { duration: 2, alloc: 1024, custom: 7 } };
  state.report = {
    ...report('buildkit'),
    benchmarks: [
      {
        name: 'BenchmarkBuild/BenchmarkBuildSimple',
        description: 'Simple build',
        metrics: {
          custom: { chart: 'bar', description: 'Custom', unit: 'count' },
          alloc: { chart: 'boxplot', description: 'Memory', unit: 'bytes' },
          duration: { chart: 'boxplot', description: 'Time', unit: 's' },
        },
        samples: [sample],
      },
    ],
  };
  state.search = ' SIMPLE ';
  assert.equal(state.filteredBenchmarks.length, 1);
  assert.equal(state.activeBenchmark.description, 'Simple build');
  assert.deepEqual(Array.from(state.metrics), ['duration', 'alloc', 'custom']);
  assert.equal(state.chartOption.series.length, 1);
  assert.equal(state.chartOption.series[0].data[0].value, 2);
  assert.equal(component.methods.deltaText.call(state, { ref: 'other', median: 1 }), '-50.0%');
  assert.equal(component.methods.deltaClass.call(state, { ref: 'other', median: 1 }), 'improved');
  state.metricName = 'custom';
  assert.equal(component.methods.deltaClass.call(state, { ref: 'other', median: 1 }), '');
  sample.values.custom = 0;
  assert.equal(component.methods.deltaText.call(state, { ref: 'other', median: 1 }), 'N/A');
});

test('JSON headers use recorded URLs and legacy metadata retains the CI attempt', async () => {
  let calls = 0;
  const { state, component } = view(undefined, async (path) => {
    calls++;
    return path.endsWith('env.txt')
      ? {
          default:
            'GITHUB_REPOSITORY=moby/buildkit-bench\nGITHUB_SHA=abcdef\nGITHUB_RUN_ID=123\nGITHUB_RUN_ATTEMPT=2\n',
        }
      : {};
  });
  await component.methods.updateMetadata.call(state);
  assert.equal(calls, 0);
  state.report = {
    ...report('buildkit'),
    metadata: {
      commit: 'abcdef',
      commit_url: 'https://example.test/commit',
      run_url: 'https://example.test/run/attempts/2',
    },
  };
  assert.equal(state.headerLinks[1].url, 'https://example.test/run/attempts/2');
  state.$route.params.result = 'legacy';
  state.report = null;
  await component.methods.updateMetadata.call(state);
  assert.equal(
    state.headerLinks[1].url,
    'https://github.com/moby/buildkit-bench/actions/runs/123/attempts/2'
  );
});

test('router defaults and unknown hashes redirect without replacing explicit run links', () => {
  const script = readFileSync(new URL('../src/router/index.js', import.meta.url), 'utf8')
    .replace(/^import .*;$/gm, '')
    .replace('export default', 'router =');
  const context = {
    createRouter: (options) => options,
    createWebHashHistory: () => ({}),
    ResultView: {},
    process: { env: { BASE_URL: '/buildkit-bench/' } },
    resultsData: { results: ['20260315', '20261013'] },
  };
  vm.runInNewContext(script, context);
  const routes = context.router.routes;
  assert.equal(routes.find((r) => r.path === '/').redirect, '/overview');
  assert.equal(routes.find((r) => r.name === 'Overview').path, '/overview/:result?');
  assert.equal(routes.find((r) => r.path.includes('pathMatch')).redirect, '/');
  assert.equal(routes.find((r) => r.path === '/result/:result').redirect, undefined);
});

test('logs are linked only when an archive is listed for the run', () => {
  const { state } = view();
  assert.equal(
    state.headerLinks.find((link) => link.text === 'Logs').url,
    '/buildkit-bench/result/both/logs.tar.gz'
  );
  state.$route.params.result = 'kit';
  assert.equal(
    state.headerLinks.some((link) => link.text === 'Logs'),
    false
  );
});

test('grouped navigation keeps per-level distributions and searchable variants', () => {
  const { state, component } = view();
  const variants = [128, 16, 32, 64].map((n) => ({
    name: `BenchmarkBuild/BenchmarkBuildHighParallelization${n}x`,
    metrics: { duration: { unit: 's', chart: 'boxplot', description: 'Time' } },
    samples: [{ buildkit: 'master', values: { duration: n } }],
  }));
  state.report = { ...report('buildkit'), benchmarks: variants };
  assert.equal(state.navigation.length, 1);
  state.search = '128x';
  assert.equal(state.filteredBenchmarks.length, 1);
  component.methods.selectBenchmark.call(state, state.navigation[0]);
  assert.equal(state.activeBenchmark.name, variants[1].name);
  assert.equal(state.chartMode, 'distribution');
  state.variantName = variants[0].name;
  assert.equal(state.activeBenchmark.name, variants[0].name);
  const restored = view().state;
  restored.$route.query = { ...state.$route.query };
  restored.report = state.report;
  assert.equal(restored.activeBenchmark.name, variants[0].name);
  assert.equal(restored.chartMode, 'distribution');
  assert.equal(state.chartOption.series[0].type, 'boxplot');
  assert.equal(state.rows[0].median, 128);
});

test('failed and legacy reports expose diagnostics without adding measurement samples', () => {
  const { state } = view();
  state.report = report('buildkit');
  assert.equal(state.failures.length, 0);
  state.report.failures = [{ name: 'BenchmarkBuild/Simple/ref=old/run=1', output: '<script>failure</script>' }];
  assert.equal(state.failures.length, 1);
  assert.equal(state.sampleCount, 0);
  assert.equal(state.activeBenchmark, undefined);
  state.report.benchmarks = permalinkReport('buildkit').benchmarks;
  assert.equal(state.failures.length, 1);
  assert.equal(state.sampleCount, 1);
});

test('failure panel renders incomplete runs and escapes diagnostic output', async () => {
  const template = readFileSync(new URL('../src/components/ResultView.vue', import.meta.url), 'utf8')
    .match(/<section v-if="!loading && !error && failures.length"[\s\S]*?<\/section>/)[0];
  const failures = [
    { name: 'BenchmarkBuild', package: 'bench', output: '<script>unsafe()</script>', source_id: 'source' },
    { name: '', package: 'setup', output: '' },
  ];
  const html = await renderToString(createSSRApp({
    template,
    data: () => ({
      loading: false, error: '', failures,
      report: { failures, sources: { source: { name: 'raw.json' } }, missing_benchmarks: ['BenchmarkBuild/Simple'] },
    }),
  }));
  assert.match(html, /2 failures recorded/);
  assert.match(html, /BenchmarkBuild\/Simple/);
  assert.match(html, /raw.json/);
  assert.match(html, /&lt;script&gt;unsafe\(\)&lt;\/script&gt;/);
  assert.doesNotMatch(html, /<script>/);
  assert.match(html, /No diagnostic output was captured/);
});
