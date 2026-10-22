<template>
  <div class="dashboard">
    <aside class="sidebar">
      <a class="brand" href="#/" aria-label="BuildKit Bench home">
        <span class="brand-logo" :style="{ maskImage: `url(${logoUrl})` }" aria-hidden="true"></span>
        <span>BuildKit Bench</span>
      </a>
      <div class="workspace-label"><span class="status-dot"></span> PUBLIC BENCHMARKS</div>
      <nav class="section-nav" aria-label="Dashboard views">
        <button :class="{ selected: view === 'benchmarks' }" @click="view = 'benchmarks'">
          <ui-icon name="chart" /> Benchmarks
          <span class="nav-count">{{ navigation.length }}</span>
        </button>
        <button :class="{ selected: view === 'history' }" @click="view = 'history'">
          <ui-icon name="history" /> Run history
        </button>
      </nav>
      <div class="nav-heading">
        BENCHMARKS <span>{{ filteredBenchmarks.length }}</span>
      </div>
      <label class="search-box"
        ><ui-icon name="search" /><input
          v-model="search"
          placeholder="Find a benchmark..."
          aria-label="Search benchmarks"
        /><button v-if="search" class="clear-search" @click="search = ''" aria-label="Clear search">
          ×
        </button></label
      >
      <nav class="benchmark-nav" aria-label="Benchmarks">
        <button
          v-for="benchmark in filteredBenchmarks"
          :key="benchmark.name"
          :class="{ active: activeEntry?.name === benchmark.name }"
          :aria-current="activeEntry?.name === benchmark.name ? 'true' : undefined"
          @click="selectBenchmark(benchmark)"
        >
          <span class="benchmark-dot"></span><span>{{ benchmark.description || benchmark.name }}</span
          ><ui-icon name="chevron" />
        </button>
        <p v-if="report && !filteredBenchmarks.length" class="nav-empty">No benchmarks match your search.</p>
        <p v-else-if="!report && !loading" class="nav-empty">Select a JSON run to explore benchmarks.</p>
      </nav>
      <a class="sidebar-footer" href="https://github.com/moby/buildkit-bench" target="_blank" rel="noopener"
        ><ui-icon name="git" /><span>moby / buildkit-bench<small>Open source benchmark suite</small></span
        ><ui-icon name="external"
      /></a>
    </aside>

    <div class="main-panel">
      <header class="topbar">
        <div class="breadcrumb">
          Performance <ui-icon name="chevron" /><strong>{{
            selectedProject === 'buildx' ? 'Buildx' : 'BuildKit'
          }}</strong>
        </div>
        <div class="topbar-actions">
          <results-list /><button
            class="icon-button"
            @click="$emit('toggle-theme')"
            :aria-label="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
          >
            <ui-icon :name="theme === 'dark' ? 'sun' : 'moon'" />
          </button>
        </div>
      </header>
      <main class="workspace" :aria-busy="loading">
        <div class="page-heading">
          <div>
            <div class="eyebrow">BUILD PERFORMANCE, IN FOCUS</div>
            <h1>{{ view === 'history' ? 'Run history' : 'Benchmark explorer' }}</h1>
            <p class="page-description">
              {{
                view === 'history'
                  ? 'Follow a benchmark over time. Keep the execution context in view.'
                  : 'Explore the measurements. Understand what changed.'
              }}
            </p>
          </div>
          <div class="segmented project-tabs" role="group" aria-label="Benchmark project">
            <button
              v-for="project in projects"
              :key="project"
              :class="{ active: selectedProject === project }"
              :aria-pressed="selectedProject === project"
              @click="activeTab = project"
            >
              {{ project === 'buildkit' ? 'BuildKit' : 'Buildx' }}
            </button>
          </div>
        </div>
        <div class="run-context">
          <span class="run-badge"><span class="status-dot"></span>{{ runLabel($route.params.result) }}</span
          ><a v-for="link in headerLinks" :key="link.url" :href="link.url" target="_blank" rel="noopener"
            ><ui-icon :name="link.kind || 'external'" />{{ link.text }}</a
          >
        </div>

        <div v-if="loading" class="empty-state" role="status">
          <span class="loading-ring"></span>
          <h2>Loading measurements</h2>
          <p>Preparing this run's benchmark reports.</p>
        </div>
        <div v-else-if="error" class="empty-state" role="alert">
          <ui-icon name="info" />
          <h2>Couldn't load this report</h2>
          <p>{{ error }}</p>
          <button class="primary-button" @click="loadReport">Try again</button>
        </div>
        <template v-else-if="!reportUrl && originalUrl"
          ><div class="notice">
            <ui-icon name="info" /><span
              >This archived run uses the original HTML report. Native charts are available for runs with JSON
              data.</span
            >
          </div>
          <iframe
            class="legacy-report"
            :src="originalUrl"
            :title="`${selectedProject} benchmark report`"
          ></iframe
        ></template>
        <div v-else-if="!activeBenchmark" class="empty-state">
          <ui-icon name="chart" />
          <h2>No measurements available</h2>
          <p>Choose another run from the selector above.</p>
        </div>
        <template v-else>
          <parallelization-chart
            v-if="activeEntry?.variants && view === 'benchmarks'"
            :key="reportUrl"
            :report="report"
            :variants="activeEntry.variants"
            :metric="selectedMetric"
            :metrics="activeBenchmark.metrics"
            :theme="theme"
            :baseline-ref="baselineRef"
            @update-baseline="baselineRef = $event"
            @update-metric="metricName = $event"
          />
          <div v-if="activeEntry?.variants" class="comparison-toolbar variant-toolbar">
            <div>
              <h3>
                {{ view === 'history' ? 'History by parallelism level' : 'Inspect a parallelism level' }}
              </h3>
              <p class="muted">Distributions and execution context remain separate for each workload.</p>
            </div>
            <label class="baseline-picker"
              >Concurrent builds<select
                :value="activeBenchmark.name"
                aria-label="Parallelism level"
                @change="variantName = $event.target.value"
              >
                <option
                  v-for="variant in activeEntry.variants"
                  :key="variant.benchmark.name"
                  :value="variant.benchmark.name"
                >
                  {{ variant.parallelism }}x
                </option>
              </select></label
            >
          </div>
          <div class="benchmark-heading">
            <div>
              <h2>{{ activeBenchmark.description || activeBenchmark.name }}</h2>
              <code>{{ activeBenchmark.name }}</code>
            </div>
            <button class="subtle-button" :aria-expanded="detailsOpen" @click="detailsOpen = !detailsOpen">
              <ui-icon name="cpu" /> Run details <span>{{ detailsOpen ? '−' : '+' }}</span>
            </button>
          </div>
          <run-details v-if="detailsOpen" :report="report" :benchmark="activeBenchmark" />
          <history-panel
            v-if="view === 'history'"
            :key="reportUrl"
            :report="report"
            :benchmark="activeBenchmark"
            :metric="selectedMetric"
            :theme="theme"
            :run="$route.params.result"
            :reference="historyRef"
            @update-reference="historyRef = $event"
            @update-metric="metricName = $event"
          />
          <template v-else>
            <div class="comparison-toolbar">
              <div class="metric-tabs" role="group" aria-label="Metric">
                <button
                  v-for="metric in metrics"
                  :key="metric"
                  :class="{ active: selectedMetric === metric }"
                  :aria-pressed="selectedMetric === metric"
                  @click="metricName = metric"
                >
                  {{ activeBenchmark.metrics[metric].description }}
                </button>
              </div>
              <label class="baseline-picker"
                >Baseline
                <select
                  :value="baselineRow?.ref"
                  @change="baselineRef = $event.target.value"
                  aria-label="Baseline candidate"
                >
                  <option v-for="row in rows" :key="row.ref" :value="row.ref">{{ row.ref }}</option>
                </select></label
              >
            </div>
            <div class="stat-grid">
              <div class="stat-card">
                <span class="stat-label">Baseline median <span class="tiny-dot purple"></span></span
                ><strong>{{ formatValue(baselineRow?.median, metricConfig?.unit) }}</strong
                ><small>{{ baselineRow?.ref || 'No baseline selected' }}</small>
              </div>
              <div class="stat-card">
                <span class="stat-label">Lowest median <span class="tiny-dot mint"></span></span
                ><strong>{{ formatValue(lowestRow?.median, metricConfig?.unit) }}</strong
                ><small
                  >{{ lowestRow?.ref || 'No data' }}
                  <span v-if="lowestRow && baselineRow" class="delta-chip" :class="deltaClass(lowestRow)">{{
                    deltaText(lowestRow)
                  }}</span></small
                >
              </div>
              <div class="stat-card">
                <span class="stat-label">Candidates compared</span
                ><strong
                  >{{ rows.length.toString().padStart(2, '0') }}<span class="stat-unit">refs</span></strong
                ><small>{{
                  report.candidates[selectedProject]?.length
                    ? 'Candidate commit order'
                    : 'Reference name order'
                }}</small>
              </div>
              <div class="stat-card">
                <span class="stat-label">Recorded samples</span
                ><strong>{{ sampleCount.toLocaleString() }}</strong
                ><small>Individual measurements, no averaging</small>
              </div>
            </div>
            <section class="chart-card">
              <div class="chart-card-header">
                <div>
                  <h3>{{ chartMode === 'medians' ? 'Median comparison' : 'Measurement distribution' }}</h3>
                  <p>
                    {{ metricConfig?.description }} <span class="quiet-separator">/</span>
                    {{ rows.length }} candidates
                  </p>
                </div>
                <div class="segmented">
                  <button
                    :class="{ active: chartMode === 'medians' }"
                    :aria-pressed="chartMode === 'medians'"
                    @click="chartMode = 'medians'"
                  >
                    Medians</button
                  ><button
                    :class="{ active: chartMode === 'distribution' }"
                    :aria-pressed="chartMode === 'distribution'"
                    @click="chartMode = 'distribution'"
                  >
                    {{ metricConfig?.chart === 'bar' ? 'Bars' : 'Box plots' }}
                  </button>
                </div>
              </div>
              <benchmark-chart v-if="chartOption" :option="chartOption" :theme="theme" />
              <div v-else class="empty-state">No measurements for this metric.</div>
              <div class="chart-caption">
                <span
                  ><span class="legend-line"></span
                  >{{
                    chartMode === 'medians'
                      ? 'Median per candidate'
                      : metricConfig?.chart === 'bar'
                      ? 'Rounded median per candidate'
                      : 'Min · Q1 · Median · Q3 · Max'
                  }}</span
                ><span v-if="baselineRow"
                  ><span class="legend-line baseline"></span>Baseline: {{ baselineRow.ref }}</span
                ><span class="caption-right">{{
                  selectedMetric === 'duration' || selectedMetric === 'alloc'
                    ? 'Lower is better'
                    : 'Compare values in the recorded unit'
                }}</span>
              </div>
            </section>
            <section class="table-card">
              <div class="table-heading">
                <div>
                  <h3>Candidate breakdown</h3>
                  <p>
                    Medians and observed range. Deltas are relative to
                    {{ baselineRow?.ref || 'the selected baseline' }}.
                  </p>
                </div>
                <span class="count-badge">{{ rows.length }} candidates</span>
              </div>
              <div class="table-scroll">
                <table>
                  <thead>
                    <tr>
                      <th>Candidate</th>
                      <th>Revision</th>
                      <th class="numeric">Median</th>
                      <th class="numeric">vs. baseline</th>
                      <th class="numeric">Min / Max</th>
                      <th class="numeric">Samples</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="row in rows"
                      :key="row.ref"
                      :class="{ 'baseline-row': row.ref === baselineRow?.ref }"
                    >
                      <td>
                        <span class="candidate-name">{{ row.ref }}</span
                        ><span v-if="row.ref === baselineRow?.ref" class="baseline-tag">BASELINE</span>
                      </td>
                      <td>
                        <code :title="row.sha">{{ row.sha ? row.sha.slice(0, 8) : 'Not recorded' }}</code>
                      </td>
                      <td class="numeric strong">{{ formatValue(row.median, metricConfig?.unit) }}</td>
                      <td class="numeric">
                        <span class="delta-chip" :class="deltaClass(row)">{{ deltaText(row) }}</span>
                      </td>
                      <td class="numeric muted">
                        {{ formatValue(row.min, metricConfig?.unit) }} /
                        {{ formatValue(row.max, metricConfig?.unit) }}
                      </td>
                      <td class="numeric muted">{{ row.count }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>
          </template>
          <footer class="workspace-footer">
            <span
              >BuildKit Benchmarks <span class="quiet-separator">/</span>
              {{ report.benchmarks.length }} benchmarks in this run</span
            ><span>Measurements, with context.</span>
          </footer>
        </template>
      </main>
    </div>
  </div>
</template>

<script>
import resultsData from '../generated/results.json';
import { markRaw } from 'vue';
import BenchmarkChart from './BenchmarkChart.vue';
import ResultsList from './ResultsList.vue';
import RunDetails from './RunDetails.vue';
import HistoryPanel from './HistoryPanel.vue';
import UiIcon from './UiIcon.vue';
import ParallelizationChart from './ParallelizationChart.vue';
import { reportCharts } from '../charts/report.mjs';
import { candidateRows, percentChange, benchmarkNavigation } from '../reports/benchmarks.mjs';
import { dashboardChart } from '../charts/dashboard.mjs';
import { formatValue, runLabel } from '../reports/format.mjs';

// Read selections directly from the route so reload and Back/Forward share the
// same state as clicks. Ignore repeated/empty query values in hand-edited links.
function querySelection(key, fallback) {
  return {
    get() {
      const value = this.$route.query[key];
      return typeof value === 'string' && value ? value : fallback;
    },
    set(value) {
      this.navigateSelection({ [key]: value });
    },
  };
}

export default {
  name: 'ResultView',
  components: { BenchmarkChart, ResultsList, RunDetails, HistoryPanel, UiIcon, ParallelizationChart },
  props: { theme: { type: String, default: 'light' } },
  emits: ['toggle-theme'],
  data: () => ({
    logoUrl: `${process.env.BASE_URL}img/moby.svg`,
    metadataLinks: [],
    report: null,
    loading: false,
    error: '',
    request: null,
    search: '',
    detailsOpen: false,
  }),
  computed: {
    activeTab: querySelection('project', 'buildkit'),
    benchmarkName: querySelection('benchmark', ''),
    variantName: querySelection('variant', ''),
    metricName: querySelection('metric', 'duration'),
    baselineRef: querySelection('baseline', ''),
    historyRef: querySelection('reference', ''),
    chartMode: {
      get() {
        const mode = this.$route.query.chart;
        return ['medians', 'distribution'].includes(mode)
          ? mode
          : this.activeEntry?.variants ? 'distribution' : 'medians';
      },
      set(chart) { this.navigateSelection({ chart }); },
    },
    view: {
      get() { return this.$route.query.view === 'history' ? 'history' : 'benchmarks'; },
      set(view) { this.navigateSelection({ view }); },
    },
    projects() {
      const result = this.$route.params.result;
      return ['buildkit', 'buildx'].filter(
        (project) =>
          resultsData.jsonReports[result]?.includes(project) || resultsData.htmlReports[result]?.[project]
      );
    },
    selectedProject() {
      return this.projects.includes(this.activeTab) ? this.activeTab : this.projects[0];
    },
    reportUrl() {
      const result = this.$route.params.result;
      return resultsData.jsonReports[result]?.includes(this.selectedProject)
        ? `${process.env.BASE_URL}result/${result}/${this.selectedProject}.json`
        : '';
    },
    originalUrl() {
      const result = this.$route.params.result;
      const file = resultsData.htmlReports[result]?.[this.selectedProject];
      return file ? `${process.env.BASE_URL}result/${result}/${file}` : '';
    },
    headerLinks() {
      const meta = this.report?.metadata || {};
      const links = this.report
        ? [
            meta.commit_url && {
              text: meta.commit?.slice(0, 8) || 'Suite commit',
              url: meta.commit_url,
              kind: 'git',
            },
            meta.run_url && { text: 'View CI run', url: meta.run_url, kind: 'external' },
          ].filter(Boolean)
        : [...this.metadataLinks];
      if (resultsData.logArchives.includes(this.$route.params.result)) {
        links.push({
          text: 'Logs',
          url: `${process.env.BASE_URL}result/${this.$route.params.result}/logs.tar.gz`,
          kind: 'download',
        });
      }
      return links;
    },
    navigation() {
      return benchmarkNavigation(this.report?.benchmarks || []);
    },
    filteredBenchmarks() {
      const query = this.search.toLowerCase().trim();
      return this.navigation.filter((b) =>
        `${b.description} ${b.name} ${(b.variants || [])
          .map((v) => `${v.parallelism}x ${v.benchmark.name}`)
          .join(' ')}`
          .toLowerCase()
          .includes(query)
      );
    },
    activeEntry() {
      return (
        this.navigation.find((b) => b.name === this.benchmarkName) ||
        this.navigation.find((b) => b.name.endsWith('/BenchmarkBuildSimple')) ||
        this.navigation[0]
      );
    },
    activeBenchmark() {
      return this.activeEntry?.variants
        ? (
            this.activeEntry.variants.find((v) => v.benchmark.name === this.variantName) ||
            this.activeEntry.variants[0]
          ).benchmark
        : this.activeEntry;
    },
    metrics() {
      if (!this.activeBenchmark) return [];
      const order = { duration: 0, alloc: 1 };
      return Object.keys(this.activeBenchmark.metrics)
        .filter((key) => this.activeBenchmark.samples.some((s) => Number.isFinite(s.values[key])))
        .sort((a, b) => (order[a] ?? 2) - (order[b] ?? 2) || a.localeCompare(b));
    },
    selectedMetric() {
      return this.metrics.includes(this.metricName) ? this.metricName : this.metrics[0];
    },
    metricConfig() {
      return this.activeBenchmark?.metrics[this.selectedMetric];
    },
    rows() {
      return this.report ? candidateRows(this.report, this.activeBenchmark, this.selectedMetric) : [];
    },
    baselineRow() {
      return this.rows.find((row) => row.ref === this.baselineRef) || this.rows[0];
    },
    lowestRow() {
      return this.rows.reduce((lowest, row) => (!lowest || row.median < lowest.median ? row : lowest), null);
    },
    sampleCount() {
      return this.rows.reduce((count, row) => count + row.count, 0);
    },
    chartOption() {
      if (!this.activeBenchmark || !this.selectedMetric) return null;
      const benchmark = {
        ...this.activeBenchmark,
        samples: this.activeBenchmark.samples.map((sample) => ({
          ...sample,
          values: Object.fromEntries(
            Object.entries(sample.values).filter(([unit]) => unit === this.selectedMetric)
          ),
        })),
      };
      const option = reportCharts({ ...this.report, benchmarks: [benchmark] })[0]?.option;
      return option
        ? dashboardChart(
            option,
            this.theme,
            this.baselineRow,
            this.chartMode,
            this.rows,
            this.metricConfig.unit
          )
        : null;
    },
  },
  watch: {
    '$route.params.result': 'updateMetadata',
    reportUrl: { immediate: true, handler: 'loadReport' },
    selectedProject: 'updateMetadata',
  },
  created() {
    this.updateMetadata();
  },
  beforeUnmount() {
    this.request?.abort();
  },
  methods: {
    formatValue,
    runLabel,
    navigateSelection(changes) {
      return this.$router.push({
        name: 'ResultView',
        params: this.$route.params,
        query: {
          ...this.$route.query,
          project: this.selectedProject,
          benchmark: this.activeEntry?.name,
          metric: this.selectedMetric || this.metricName,
          ...changes,
        },
      });
    },
    selectBenchmark(benchmark) {
      return this.navigateSelection({
        benchmark: benchmark.name,
        variant: undefined,
        chart: benchmark.variants ? 'distribution' : 'medians',
      });
    },
    deltaText(row) {
      if (row.ref === this.baselineRow?.ref) return 'Baseline';
      const delta = percentChange(row.median, this.baselineRow?.median);
      return delta === null ? 'N/A' : `${delta > 0 ? '+' : ''}${delta.toFixed(1)}%`;
    },
    deltaClass(row) {
      if (!['duration', 'alloc'].includes(this.selectedMetric) || row.ref === this.baselineRow?.ref)
        return '';
      const delta = percentChange(row.median, this.baselineRow?.median);
      return delta === null ? '' : delta < 0 ? 'improved' : delta > 0 ? 'regressed' : '';
    },
    async loadReport() {
      this.request?.abort();
      const request = markRaw(new AbortController());
      this.request = request;
      this.report = null;
      this.error = '';
      this.loading = !!this.reportUrl;
      if (!this.reportUrl) return;
      const project = this.selectedProject;
      try {
        const response = await fetch(this.reportUrl, { signal: request.signal });
        if (!response.ok) throw new Error(`Report request failed (${response.status})`);
        const report = await response.json();
        if (request.signal.aborted) return;
        if (report.project !== project) throw new Error('Report project does not match the selection');
        reportCharts(report);
        this.report = markRaw(report);
      } catch (error) {
        if (!request.signal.aborted) this.error = error.message;
      } finally {
        if (!request.signal.aborted) this.loading = false;
      }
    },
    async updateMetadata() {
      this.metadataLinks = [];
      const resultName = this.$route.params.result;
      if (this.reportUrl) return;
      try {
        const event = await import(`../../public/result/${resultName}/gha-event.json`);
        const env = await import(`../../public/result/${resultName}/env.txt`);
        if (resultName !== this.$route.params.result) return;
        const envs = Object.fromEntries(
          env.default
            .split('\n')
            .filter(Boolean)
            .map((line) => {
              const split = line.indexOf('=');
              return [line.slice(0, split), line.slice(split + 1).trim()];
            })
        );
        const repo = event.repository?.full_name || envs.GITHUB_REPOSITORY;
        if (!repo) return;
        const links = [];
        if (envs.GITHUB_SHA)
          links.push({
            text: envs.GITHUB_SHA.slice(0, 8),
            url: `https://github.com/${repo}/commit/${envs.GITHUB_SHA}`,
            kind: 'git',
          });
        if (envs.GITHUB_RUN_ID)
          links.push({
            text: 'View CI run',
            url: `https://github.com/${repo}/actions/runs/${envs.GITHUB_RUN_ID}${
              envs.GITHUB_RUN_ATTEMPT ? `/attempts/${envs.GITHUB_RUN_ATTEMPT}` : ''
            }`,
            kind: 'external',
          });
        this.metadataLinks = links;
      } catch {
        /* Older reports may not contain workflow metadata. */
      }
    },
  },
};
</script>

<style scoped>
.dashboard {
  display: flex;
  min-height: 100vh;
}
.sidebar {
  width: 320px;
  flex-shrink: 0;
  background: var(--sidebar);
  border-right: 1px solid var(--border);
  position: sticky;
  top: 0;
  height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 30px 16px 18px;
}
.brand {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 0 12px;
  font-size: 23px;
  font-weight: 700;
  letter-spacing: -0.6px;
  white-space: nowrap;
}
.brand-logo {
  width: 42px;
  height: 42px;
  flex-shrink: 0;
  background: var(--accent);
  mask-size: contain;
  mask-repeat: no-repeat;
  mask-position: center;
}
.workspace-label {
  font-size: 9px;
  letter-spacing: 1.4px;
  color: var(--muted);
  margin: 38px 12px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.status-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--green);
  flex-shrink: 0;
}
.section-nav {
  display: grid;
  gap: 6px;
  padding-bottom: 25px;
  border-bottom: 1px solid var(--border);
}
.section-nav button {
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 12px;
  text-align: left;
  border-radius: 7px;
  color: var(--muted);
  font-size: 13px;
}
.section-nav button.selected {
  color: var(--accent);
  background: var(--selected);
  font-weight: 600;
}
.section-nav button:hover,
.benchmark-nav button:hover {
  background: var(--hover);
}
.nav-count {
  margin-left: auto;
  font-size: 10px;
}
.nav-heading {
  display: flex;
  justify-content: space-between;
  padding: 25px 12px 15px;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 1.3px;
  color: var(--muted);
}
.search-box {
  display: flex;
  gap: 8px;
  align-items: center;
  border: 1px solid var(--border);
  border-radius: 6px;
  margin: 0 4px 13px;
  padding: 9px;
  color: var(--muted);
}
.search-box svg {
  width: 14px;
}
.search-box input {
  width: 100%;
  min-width: 0;
  background: transparent;
  border: 0;
  font-size: 11px;
}
.search-box input::placeholder {
  color: var(--muted);
}
.clear-search {
  padding: 0;
}
.benchmark-nav {
  overflow-y: auto;
  scrollbar-width: thin;
  flex: 1;
  min-height: 0;
}
.benchmark-nav button {
  display: flex;
  width: 100%;
  text-align: left;
  align-items: center;
  gap: 9px;
  padding: 11px 12px;
  border-radius: 6px;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.5;
}
.benchmark-nav button.active {
  background: var(--selected);
  color: var(--accent);
}
.benchmark-nav button svg {
  width: 12px;
  margin-left: auto;
  opacity: 0;
}
.benchmark-nav button.active svg {
  opacity: 1;
}
.benchmark-dot {
  width: 5px;
  height: 5px;
  flex-shrink: 0;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.6;
}
.nav-empty {
  font-size: 12px;
  padding: 12px;
  color: var(--muted);
  line-height: 1.6;
}
.sidebar-footer {
  display: flex;
  align-items: center;
  gap: 9px;
  border-top: 1px solid var(--border);
  margin-top: 15px;
  padding: 20px 8px 0;
  font-size: 11px;
}
.sidebar-footer small {
  display: block;
  color: var(--muted);
  font-size: 9px;
  margin-top: 5px;
}
.sidebar-footer > svg:last-child {
  width: 12px;
  margin-left: auto;
  color: var(--muted);
}
.main-panel {
  flex: 1;
  min-width: 0;
}
.topbar {
  height: 77px;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 36px;
  background: var(--panel);
  gap: 16px;
}
.breadcrumb {
  display: flex;
  gap: 15px;
  align-items: center;
  color: var(--muted);
  font-size: 12px;
}
.breadcrumb svg {
  width: 11px;
}
.breadcrumb strong {
  font-weight: 500;
  color: var(--text);
}
.topbar-actions {
  display: flex;
  align-items: center;
  gap: 18px;
}
.icon-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--muted);
}
.icon-button:hover {
  background: var(--hover);
}
.workspace {
  max-width: 1580px;
  margin: auto;
  padding: 35px 36px 20px;
}
.page-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}
.eyebrow {
  font-size: 9px;
  color: var(--accent);
  letter-spacing: 1.8px;
  font-weight: 600;
  margin-bottom: 10px;
}
.page-description {
  font-size: 13px;
  color: var(--muted);
  margin-top: 10px;
  line-height: 1.6;
}
.project-tabs button {
  font-size: 12px;
  padding: 9px 20px;
}
.run-context {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 18px;
  margin: 22px 0 30px;
  padding-bottom: 25px;
  border-bottom: 1px solid var(--border);
  color: var(--muted);
  font-size: 11px;
}
.run-context a {
  display: flex;
  align-items: center;
  gap: 6px;
}
.run-context a:hover {
  color: var(--accent);
}
.run-context svg {
  width: 13px;
  height: 13px;
}
.run-badge {
  display: flex;
  align-items: center;
  gap: 8px;
}
.benchmark-heading {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 15px;
  margin-bottom: 25px;
}
.benchmark-heading code {
  display: block;
  font-size: 10px;
  color: var(--muted);
  margin-top: 8px;
  overflow-wrap: anywhere;
}
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  margin-bottom: 22px;
}
.stat-card {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 9px;
  padding: 19px 20px;
  box-shadow: var(--shadow);
}
.stat-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 11px;
  color: var(--muted);
  margin-bottom: 14px;
}
.stat-card > strong {
  font-size: 26px;
  font-weight: 600;
  letter-spacing: -0.8px;
  display: block;
  font-variant-numeric: tabular-nums;
}
.stat-card small {
  display: block;
  font-size: 10px;
  color: var(--muted);
  margin-top: 12px;
  line-height: 1.5;
}
.stat-unit {
  font-size: 11px;
  font-weight: 400;
  color: var(--muted);
  letter-spacing: 0;
  margin-left: 7px;
}
.tiny-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
.purple {
  background: var(--purple);
}
.mint {
  background: var(--green);
}
.delta-chip {
  font-size: 10px;
  color: var(--muted);
  padding: 4px 6px;
  border-radius: 4px;
  white-space: nowrap;
}
.delta-chip.improved {
  color: var(--green);
  background: var(--green-bg);
}
.delta-chip.regressed {
  color: var(--red);
  background: var(--red-bg);
}
.stat-card .delta-chip {
  margin-left: 5px;
  font-size: 9px;
}
.baseline-row {
  background: var(--selected);
}
.candidate-name {
  font-weight: 600;
}
.workspace-footer {
  display: flex;
  justify-content: space-between;
  gap: 15px;
  color: var(--muted);
  font-size: 10px;
  padding: 8px 0 15px;
}
.legacy-report {
  width: 100%;
  height: 75vh;
  background: white;
  border: 0;
}
.loading-ring {
  display: inline-block;
  width: 25px;
  height: 25px;
  border: 2px solid var(--border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 20px;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
@media (min-width: 1600px) {
  .workspace {
    padding: 40px 50px;
  }
}
@media (max-width: 1150px) {
  .sidebar {
    padding-left: 10px;
    padding-right: 10px;
  }
  .workspace {
    padding: 28px 24px;
  }
  .topbar {
    padding: 0 24px;
  }
  .stat-card {
    padding: 16px;
  }
  .stat-card > strong {
    font-size: 22px;
  }
}
@media (max-width: 850px) {
  .dashboard {
    display: block;
  }
  .sidebar {
    position: static;
    width: 100%;
    height: auto;
    padding: 18px 20px;
    border-right: 0;
    border-bottom: 1px solid var(--border);
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }
  .brand {
    padding: 0;
    font-size: 21px;
  }
  .workspace-label,
  .nav-heading,
  .sidebar-footer {
    display: none;
  }
  .section-nav {
    display: flex;
    border: 0;
    padding: 0;
    justify-content: flex-end;
    gap: 4px;
  }
  .section-nav button {
    font-size: 11px;
    padding: 8px;
    gap: 6px;
  }
  .nav-count {
    display: none;
  }
  .search-box {
    margin: 0;
  }
  .benchmark-nav {
    max-height: 95px;
  }
  .benchmark-nav button {
    padding: 7px 10px;
  }
  .topbar {
    height: 65px;
  }
  .stat-grid {
    gap: 10px;
  }
  .benchmark-heading {
    margin-bottom: 20px;
  }
}
@media (max-width: 560px) {
  .workspace {
    padding: 24px 16px;
  }
  .topbar {
    padding: 0 16px;
  }
  .breadcrumb {
    display: none;
  }
  .topbar-actions {
    width: 100%;
    justify-content: space-between;
    gap: 10px;
  }
  .sidebar {
    padding: 16px;
    gap: 12px;
    grid-template-columns: 1fr;
  }
  .brand {
    grid-row: 1;
  }
  .section-nav {
    grid-row: 1;
    justify-self: end;
  }
  .section-nav button {
    font-size: 0;
  }
  .section-nav svg {
    width: 20px;
    height: 20px;
  }
  .search-box {
    grid-row: 2;
    grid-column: 1 / -1;
  }
  .benchmark-nav {
    grid-row: 3;
    grid-column: 1 / -1;
    display: flex;
    max-height: 75px;
    overflow: auto;
    gap: 6px;
  }
  .benchmark-nav button {
    min-width: 160px;
    max-width: 200px;
    font-size: 11px;
  }
  .benchmark-nav button svg {
    display: none;
  }
  .page-heading {
    flex-wrap: wrap;
    gap: 16px;
  }
  .eyebrow {
    font-size: 8px;
  }
  .page-description {
    font-size: 12px;
  }
  .project-tabs button {
    padding: 7px 18px;
  }
  .run-context {
    gap: 14px;
    margin: 20px 0;
    padding-bottom: 20px;
    font-size: 10px;
  }
  .benchmark-heading {
    flex-wrap: wrap;
  }
  .stat-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .stat-card > strong {
    font-size: 24px;
  }
  .stat-card small {
    font-size: 9px;
  }
  .stat-label {
    font-size: 10px;
  }
  .workspace-footer {
    flex-wrap: wrap;
    font-size: 9px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .loading-ring {
    animation: none;
  }
}
.variant-toolbar {
  padding-top: 12px;
}
.variant-toolbar p {
  font-size: 11px;
  margin-top: 8px;
  line-height: 1.6;
}
</style>
