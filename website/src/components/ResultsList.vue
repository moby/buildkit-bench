<template>
  <label class="run-picker">
    <span>RUN</span>
    <select
      aria-label="Benchmark run"
      :value="run"
      @change="loadResult($event.target.value)"
    >
      <option v-for="result in results" :key="result" :value="result">{{ runLabel(result) }}</option>
    </select>
  </label>
</template>

<script>
import resultsData from '../generated/results.json';
import { runLabel } from '../reports/format.mjs';

export default {
  props: { run: String },
  data: () => ({ results: [...resultsData.results].sort().reverse() }),
  methods: {
    runLabel,
    loadResult(result) {
      this.$router.push({
        name: this.$route.name === 'Overview' ? 'Overview' : 'ResultView',
        params: { result },
        query: this.$route.query,
      });
    },
  },
};
</script>

<style scoped>
.run-picker {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
}
.run-picker > span {
  font-size: 9px;
  color: var(--muted);
  letter-spacing: 1px;
}
@media (max-width: 560px) {
  .run-picker {
    gap: 7px;
  }
  .run-picker select {
    font-size: 11px;
  }
}
</style>
