import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import vm from 'node:vm';
import { markRaw } from 'vue';
import { reportCharts } from '../src/charts/report.mjs';

// Exercise the component's actual loading methods/computed properties without
// requiring a DOM or mounting ECharts in Node.
const source = readFileSync(new URL('../src/components/ResultView.vue', import.meta.url), 'utf8')
  .match(/<script>([\s\S]*?)<\/script>/)[1]
  .replace(/^import .*;$/gm, '').replace(/\bimport\(/g, 'loadMetadata(')
  .replace('export default', 'component =');

function view(fetch, loadMetadata) {
  const context = {
    fetch, loadMetadata, markRaw, reportCharts, AbortController, BenchmarkChart: {},
    require: path => path,
    process: { env: { BASE_URL: '/buildkit-bench/' } },
    resultsData: {
      jsonReports: { both: ['buildkit', 'buildx'], kit: ['buildkit'], legacy: [] },
      htmlReports: { both: { buildkit: 'buildkit.html', buildx: 'buildx.html' }, legacy: { buildkit: 'index.html' } },
    },
  };
  vm.runInNewContext(source, context);
  const component = context.component;
  const state = { ...component.data(), $route: { params: { result: 'both' } } };
  for (const [key, getter] of Object.entries(component.computed)) {
    Object.defineProperty(state, key, { get: () => getter.call(state) });
  }
  state.loadReport = component.methods.loadReport.bind(state);
  return { state, component };
}

const report = project => ({ version: 1, project, candidates: {}, benchmarks: [] });
const response = project => ({ ok: true, json: async () => report(project) });

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

test('switching projects cancels the old request and ignores late results', async () => {
  const pending = [];
  const { state } = view((url, { signal }) => new Promise(resolve => pending.push({ url, signal, resolve })));
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
  state.charts = [{}];
  state.$route.params.result = 'legacy';
  await state.loadReport();
  assert.equal(request.signal.aborted, true);
  assert.equal(state.charts.length, 0);
  assert.equal(state.loading, false);
  component.beforeUnmount.call(state);
  assert.equal(state.request.signal.aborted, true);
});

test('navigating X to Y to X does not append metadata links twice', async () => {
  const pending = [];
  const { state, component } = view(undefined, path => path.endsWith('env.txt')
    ? Promise.resolve({ default: 'GITHUB_REPOSITORY=moby/buildkit-bench\nGITHUB_SHA=123456789\n' })
    : new Promise(resolve => pending.push(resolve)));
  const first = component.methods.updateMetadata.call(state);
  state.$route.params.result = 'kit';
  const second = component.methods.updateMetadata.call(state);
  state.$route.params.result = 'both';
  const third = component.methods.updateMetadata.call(state);
  pending[2]({});
  await third;
  assert.equal(state.metadataLinks.length, 2);
  pending[0]({});
  pending[1]({});
  await Promise.all([first, second]);
  assert.equal(state.metadataLinks.length, 2);
});

test('initial navigation to an older run is not replaced by the newest result', async () => {
  const script = readFileSync(new URL('../src/components/ResultsList.vue', import.meta.url), 'utf8')
    .match(/<script>([\s\S]*?)<\/script>/)[1]
    .replace(/\bimport\(/g, 'loadResults(').replace('export default', 'component =');
  const context = { loadResults: async () => ({ results: ['20260315', '20261013'] }) };
  vm.runInNewContext(script, context);
  let ready;
  const initialNavigation = new Promise(resolve => { ready = resolve; });
  const redirects = [];
  const state = {
    ...context.component.data(),
    $route: { params: {} },
    $router: {
      isReady: () => initialNavigation,
      push: path => { redirects.push(path); return Promise.resolve(); },
    },
  };
  state.updateSelectedResult = context.component.methods.updateSelectedResult.bind(state);
  const created = context.component.created.call(state);
  await new Promise(resolve => setImmediate(resolve));
  assert.deepEqual(redirects, []);
  state.$route.params.result = '20260315';
  ready();
  await created;
  assert.equal(state.selectedResult, '20260315');
  assert.deepEqual(redirects, []);
});
