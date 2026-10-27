<template>
  <section class="chart-card parallelization-chart">
    <div class="chart-card-header">
      <div>
        <h2>{{ emulated ? 'Native and emulated builds' : 'Build high parallelization' }}</h2>
        <p>{{ emulated ? 'Median measurements for the same workload.' : 'Median measurements across concurrent build counts.' }}</p>
      </div>
      <label class="baseline-picker"
        >Baseline
        <select
          :value="baseline"
          aria-label="Comparison baseline"
          @change="$emit('update-baseline', $event.target.value)"
        >
          <option v-for="row in rows" :key="row.ref" :value="row.ref">{{ row.ref }}</option>
        </select>
      </label>
    </div>
    <div class="parallelism-controls">
      <div class="metric-tabs" role="group" aria-label="Comparison metric">
        <button
          v-for="key in metricKeys"
          :key="key"
          :class="{ active: metric === key }"
          :aria-pressed="metric === key"
          @click="$emit('update-metric', key)"
        >
          {{ metrics[key].description }}
        </button>
      </div>
      <details class="candidate-selector">
        <summary>
          Compare candidates <span class="count-badge">{{ selected.length }}</span>
        </summary>
        <div class="candidate-options">
          <label v-for="row in rows" :key="row.ref"
            ><input
              type="checkbox"
              :checked="selected.includes(row.ref)"
              :disabled="row.ref === baseline"
              @change="toggleCandidate(row.ref)"
            />{{ row.ref }}<span v-if="row.ref === baseline" class="baseline-tag">BASELINE</span></label
          >
        </div>
      </details>
    </div>
    <benchmark-chart v-if="rows.length" :option="option" :theme="theme" />
    <p v-else class="notice">No measurements for this metric.</p>
    <div class="chart-caption">
      <span v-for="series in option.series" :key="series.name"
        ><span class="legend-line" :style="{ background: series.lineStyle.color }"></span>{{ series.name
        }}{{ series.name === baseline ? ' (baseline)' : '' }}</span
      >
    </div>
    <p v-if="emulated" class="parallelism-note">
      Native and emulated builds run the same compression, archive and checksum workload.
      Host architecture and emulator configuration affect results. Missing measurements leave gaps.
    </p>
    <p v-else class="parallelism-note">
      Each level launches that many concurrent builds, increasing total work. These are recorded measurements,
      not normalized speedup. Missing measurements leave gaps.
    </p>
  </section>
</template>
<script>
import BenchmarkChart from './BenchmarkChart.vue';
import { variantRows } from '../reports/benchmarks.mjs';
import { chartPalette } from '../charts/dashboard.mjs';
import { formatValue } from '../reports/format.mjs';
export default {
  components: { BenchmarkChart },
  props: {
    report: Object,
    variants: Array,
    metric: String,
    metrics: Object,
    theme: String,
    baselineRef: String,
    kind: String,
  },
  emits: ['update-baseline', 'update-metric'],
  data: () => ({ selection: null }),
  computed: {
    emulated() {
      return this.kind === 'emulator';
    },
    metricKeys() {
      const order = { duration: 0, alloc: 1 };
      return Object.keys(this.metrics).sort(
        (a, b) => (order[a] ?? 2) - (order[b] ?? 2) || a.localeCompare(b)
      );
    },
    unit() {
      return this.metrics[this.metric]?.unit;
    },
    rows() {
      return variantRows(this.report, this.variants, this.metric, this.unit);
    },
    baseline() {
      return this.rows.find((row) => row.ref === this.baselineRef)?.ref || this.rows[0]?.ref;
    },
    selected() {
      const secondary =
        this.rows.find((row) => row.ref === 'master' && row.ref !== this.baseline) ||
        this.rows.findLast((row) => row.ref !== this.baseline);
      const refs = this.selection || (secondary ? [secondary.ref] : []);
      return this.rows
        .filter((row) => row.ref === this.baseline || refs.includes(row.ref))
        .map((row) => row.ref);
    },
    option() {
      const palette = chartPalette(this.theme);
      const colors = [palette.accent, '#55bfa1', '#e6a456', '#d17eb0', '#9a93f5', '#61b5cf'];
      return {
        backgroundColor: 'transparent',
        animationDuration: 250,
        animationDurationUpdate: 250,
        textStyle: { color: palette.text, fontFamily: 'Inter, Segoe UI, sans-serif' },
        grid: { top: 35, left: 85, right: 35, bottom: 65 },
        tooltip: {
          trigger: 'axis',
          backgroundColor: palette.background,
          borderColor: palette.border,
          textStyle: { color: palette.text },
          valueFormatter: (value) => formatValue(value, this.unit),
        },
        xAxis: {
          type: 'category',
          name: this.emulated ? 'Execution' : 'Concurrent builds',
          nameLocation: 'middle',
          nameGap: 35,
          data: this.variants.map((v) => v.label),
          axisLabel: { color: palette.text },
          axisTick: { show: false },
          axisLine: { lineStyle: { color: palette.grid } },
        },
        yAxis: {
          type: 'value',
          axisLabel: { color: palette.text, formatter: (value) => formatValue(value, this.unit) },
          splitLine: { lineStyle: { color: palette.grid, type: 'dashed' } },
        },
        series: this.rows
          .filter((row) => this.selected.includes(row.ref))
          .map((row) => {
            const color =
              row.ref === this.baseline ? '#a88aef' : colors[this.rows.indexOf(row) % colors.length];
            return {
              name: row.ref,
              type: this.emulated ? 'bar' : 'line',
              connectNulls: false,
              smooth: false,
              symbol: 'circle',
              symbolSize: 8,
              itemStyle: { color },
              lineStyle: {
                color,
                width: row.ref === this.baseline ? 3 : 2,
                type: row.ref === this.baseline ? 'dashed' : 'solid',
              },
              data: row.points.map((point) => point?.median ?? null),
            };
          }),
      };
    },
  },
  methods: {
    toggleCandidate(ref) {
      this.selection = this.selected.includes(ref)
        ? this.selected.filter((candidate) => candidate !== ref)
        : [...this.selected, ref];
    },
  },
};
</script>

<style scoped>
.parallelism-controls {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 20px;
  padding: 0 24px 16px;
  flex-wrap: wrap;
}
.candidate-selector {
  max-width: 100%;
}
.candidate-selector > summary {
  padding: 10px 0;
}
.candidate-options {
  display: grid;
  grid-template-columns: repeat(2, minmax(110px, 1fr));
  gap: 12px;
  max-height: 230px;
  overflow: auto;
  padding: 14px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
}
.candidate-options label {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 12px;
}
.candidate-options input {
  accent-color: var(--accent);
}
.parallelism-note {
  padding: 0 24px 20px;
  color: var(--muted);
  font-size: 11px;
  line-height: 1.7;
}
</style>
