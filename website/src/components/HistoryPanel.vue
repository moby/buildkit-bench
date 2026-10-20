<template>
  <section class="history-panel">
    <div class="comparison-toolbar">
      <div>
        <h3>Track a candidate</h3>
        <p class="muted">
          Up to six JSON runs ending at the selected run. Older HTML-only runs are excluded.
        </p>
      </div>
      <label class="baseline-picker"
        >Reference
        <select v-model="reference" aria-label="History reference">
          <option v-for="row in candidates" :key="row.ref" :value="row.ref">{{ row.ref }}</option>
        </select></label
      >
    </div>
    <div class="metric-tabs history-metrics" role="group" aria-label="History metric">
      <button
        v-for="key in metrics"
        :key="key"
        :class="{ active: selectedMetric === key }"
        :aria-pressed="selectedMetric === key"
        @click="metricName = key"
      >
        {{ benchmark.metrics[key].description }}
      </button>
    </div>
    <div v-if="loading" class="empty-state" role="status">Loading available history...</div>
    <template v-else>
      <section class="table-card environment-review">
        <div class="table-heading">
          <div>
            <h3><ui-icon name="cpu" /> Compare execution context</h3>
            <p>
              A matching reference can point to different commits. Review the recorded environments before
              opening the trend.
            </p>
          </div>
        </div>
        <div class="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Run</th>
                <th>Candidate revision</th>
                <th>Runner CPU</th>
                <th>Go runtime</th>
                <th>Compared with selected run</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="point in points" :key="point.name">
                <td>
                  {{ runLabel(point.name)
                  }}<span v-if="point.name === run" class="baseline-tag">SELECTED</span>
                </td>
                <td>
                  <code>{{ point.row?.sha?.slice(0, 8) || 'Not recorded' }}</code>
                </td>
                <td class="wrap-cell">{{ point.context.CPU?.join(', ') || 'Not recorded' }}</td>
                <td>{{ point.context['Go runtime']?.join(', ') || 'Not recorded' }}</td>
                <td>
                  <span
                    :class="[
                      'context-tag',
                      {
                        uncertain:
                          point.problem || point.comparison.missing.length || point.comparison.changed.length,
                      },
                    ]"
                    >{{ point.problem || point.comparison.label }}</span
                  >
                  <details v-if="!point.problem" class="context-differences">
                    <summary>Inspect recorded fields</summary>
                    <p v-if="point.comparison.changed.length">
                      Changed: {{ point.comparison.changed.join(', ') }}.
                    </p>
                    <p v-if="point.comparison.missing.length">
                      Incomplete: {{ point.comparison.missing.join(', ') }}.
                    </p>
                    <dl class="key-values comparison-values">
                      <div v-for="key in point.comparison.changed" :key="key">
                        <dt>{{ key }}</dt>
                        <dd>
                          Selected: {{ currentContext[key].join('; ') }}<br />This run:
                          {{ point.context[key].join('; ') }}
                        </dd>
                      </div>
                    </dl>
                    <dl class="key-values">
                      <div v-for="(values, key) in point.context" :key="key">
                        <dt>{{ key }}</dt>
                        <dd>{{ values.join('; ') }}</dd>
                      </div>
                    </dl>
                  </details>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="history-action">
          <p>
            Environment differences and missing metadata can limit interpretation. Matching recorded fields do
            not establish identical conditions.
          </p>
          <button v-if="!reviewed" class="primary-button" :disabled="!validPoints" @click="reviewed = true">
            Show trend <ui-icon name="arrow" /></button
          ><button v-if="points.some((p) => p.error)" class="subtle-button" @click="loadHistory">
            Retry failed requests
          </button>
        </div>
      </section>
      <section v-if="reviewed && validPoints" class="chart-card">
        <div class="chart-card-header">
          <div>
            <h3>{{ reference }} over time</h3>
            <p>
              {{ benchmark.metrics[selectedMetric]?.description }} · Median per run ·
              {{ validPoints }} recorded point{{ validPoints === 1 ? '' : 's' }}
            </p>
          </div>
          <span class="count-badge">{{ reference }}</span>
        </div>
        <benchmark-chart :option="option" :theme="theme" />
        <div class="chart-caption">
          <span><span class="legend-line"></span>Median of recorded samples</span
          ><span class="caption-right">{{
            validPoints < 2
              ? 'Another matching run is needed for a trend.'
              : 'Gaps indicate unavailable measurements.'
          }}</span>
        </div>
        <div class="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Run</th>
                <th class="numeric">Median</th>
                <th class="numeric">Samples</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="point in points" :key="point.name">
                <td>{{ runLabel(point.name) }}</td>
                <td class="numeric">
                  {{
                    point.row
                      ? formatValue(point.row.median, benchmark.metrics[selectedMetric]?.unit)
                      : 'Not available'
                  }}
                </td>
                <td class="numeric">{{ point.row?.count || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
      <p v-if="!validPoints" class="notice">
        No matching measurements are available for this benchmark and reference.
      </p>
    </template>
  </section>
</template>
<script>
import { markRaw } from 'vue';
import resultsData from '../generated/results.json';
import BenchmarkChart from './BenchmarkChart.vue';
import UiIcon from './UiIcon.vue';
import { reportCharts } from '../charts/report.mjs';
import { candidateRows, comparisonContext, compareContexts } from '../reports/benchmarks.mjs';
import { chartPalette } from '../charts/dashboard.mjs';
import { runLabel, runTimestamp, formatValue } from '../reports/format.mjs';

export default {
  components: { BenchmarkChart, UiIcon },
  props: { report: Object, benchmark: Object, metric: String, theme: String, run: String },
  data: () => ({
    entries: [],
    loading: false,
    request: null,
    reference: '',
    metricName: '',
    reviewed: false,
  }),
  computed: {
    metrics() {
      return Object.keys(this.benchmark.metrics).filter((key) =>
        this.benchmark.samples.some((sample) => Number.isFinite(sample.values[key]))
      );
    },
    selectedMetric() {
      return this.metrics.includes(this.metricName)
        ? this.metricName
        : this.metrics.includes(this.metric)
        ? this.metric
        : this.metrics[0];
    },
    candidates() {
      return candidateRows(this.report, this.benchmark, this.selectedMetric);
    },
    selectedReference() {
      return this.candidates.some((row) => row.ref === this.reference)
        ? this.reference
        : this.candidates.find((row) => row.ref === 'master')?.ref || this.candidates[0]?.ref;
    },
    currentContext() {
      return comparisonContext(this.report, this.benchmark, this.selectedReference);
    },
    points() {
      const current = this.currentContext;
      return this.entries.map((entry) => {
        const benchmark = entry.report?.benchmarks.find((b) => b.name === this.benchmark.name);
        const compatible =
          benchmark?.metrics[this.selectedMetric]?.unit === this.benchmark.metrics[this.selectedMetric]?.unit;
        const row = compatible
          ? candidateRows(entry.report, benchmark, this.selectedMetric).find(
              (r) => r.ref === this.selectedReference
            )
          : undefined;
        const context = entry.report
          ? comparisonContext(entry.report, benchmark, this.selectedReference)
          : {};
        return {
          ...entry,
          row,
          context,
          comparison: compareContexts(current, context),
          problem: entry.error
            ? 'Request failed'
            : !benchmark
            ? 'Benchmark absent'
            : !compatible
            ? 'Metric unavailable or unit changed'
            : !row
            ? 'Reference has no samples'
            : '',
        };
      });
    },
    validPoints() {
      return this.points.filter((point) => point.row).length;
    },
    option() {
      const palette = chartPalette(this.theme);
      return {
        backgroundColor: 'transparent',
        color: [palette.accent],
        textStyle: { color: palette.text, fontFamily: 'Inter, Segoe UI, sans-serif' },
        grid: { top: 30, bottom: 70, left: 75, right: 35 },
        tooltip: {
          trigger: 'axis',
          valueFormatter: (value) => formatValue(value, this.benchmark.metrics[this.selectedMetric]?.unit),
          backgroundColor: palette.background,
          borderColor: palette.border,
          textStyle: { color: palette.text },
        },
        useUTC: true,
        xAxis: {
          type: 'time',
          axisTick: { show: false },
          axisLine: { lineStyle: { color: palette.grid } },
          axisLabel: { color: palette.text },
        },
        yAxis: {
          type: 'value',
          axisLabel: {
            color: palette.text,
            formatter: (value) => formatValue(value, this.benchmark.metrics[this.selectedMetric]?.unit),
          },
          splitLine: { lineStyle: { color: palette.grid, type: 'dashed' } },
        },
        series: [
          {
            name: this.benchmark.metrics[this.selectedMetric]?.description,
            type: 'line',
            smooth: false,
            connectNulls: false,
            symbol: 'circle',
            symbolSize: 9,
            lineStyle: { width: 2 },
            areaStyle: { opacity: 0.06 },
            data: this.points.map((p) => [runTimestamp(p.name), p.row?.median ?? null]),
          },
        ],
      };
    },
  },
  watch: {
    run: { immediate: true, handler: 'loadHistory' },
    selectedReference: {
      immediate: true,
      handler(value) {
        this.reference = value;
        this.reviewed = false;
      },
    },
    selectedMetric() {
      this.reviewed = false;
    },
    'benchmark.name'() {
      this.reviewed = false;
    },
  },
  beforeUnmount() {
    this.request?.abort();
  },
  methods: {
    runLabel,
    formatValue,
    async loadHistory() {
      this.request?.abort();
      const request = markRaw(new AbortController());
      this.request = request;
      this.loading = true;
      this.reviewed = false;
      this.entries = [];
      const project = this.report.project;
      const names = resultsData.results
        .filter((name) => name <= this.run && resultsData.jsonReports[name]?.includes(project))
        .sort()
        .slice(-6);
      const entries = await Promise.all(
        names.map(async (name) => {
          if (name === this.run) return { name, report: this.report };
          try {
            const response = await fetch(`${process.env.BASE_URL}result/${name}/${project}.json`, {
              signal: request.signal,
            });
            if (!response.ok) throw new Error(`HTTP ${response.status}`);
            const report = await response.json();
            if (report.project !== project) throw new Error('Project mismatch');
            reportCharts(report);
            return { name, report };
          } catch (error) {
            return { name, error: error.message };
          }
        })
      );
      if (request.signal.aborted) return;
      this.entries = markRaw(entries);
      this.loading = false;
    },
  },
};
</script>

<style scoped>
.history-panel > .comparison-toolbar p {
  font-size: 11px;
  margin-top: 8px;
  line-height: 1.6;
}
.history-metrics {
  margin-bottom: 24px;
}
.environment-review .table-heading {
  align-items: flex-start;
}
.environment-review h3 svg {
  color: var(--accent);
  margin-right: 8px;
}
.context-tag {
  font-size: 10px;
  color: var(--green);
}
.context-tag.uncertain {
  color: var(--purple);
}
.context-differences {
  margin-top: 9px;
  max-width: 380px;
}
.context-differences p {
  white-space: normal;
  font-size: 11px;
  color: var(--muted);
  line-height: 1.6;
  margin-top: 10px;
}
.history-action {
  padding: 20px 24px;
  display: flex;
  align-items: center;
  gap: 20px;
  border-top: 1px solid var(--border);
}
.history-action p {
  color: var(--muted);
  font-size: 11px;
  line-height: 1.7;
  flex: 1;
}
@media (max-width: 560px) {
  .history-action {
    flex-wrap: wrap;
    padding: 18px;
  }
  .history-action p {
    flex-basis: 100%;
  }
}
</style>
