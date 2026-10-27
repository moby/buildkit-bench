<template>
  <section class="overview">
    <div class="overview-intro">
      <div>
        <h2>Where to look first</h2>
        <p v-if="analysis.baseline && analysis.target">
          <strong>master</strong> against <strong>{{ analysis.baseline.name }}</strong
          >, the latest stable release in this report.
        </p>
        <p v-else>
          Both master and a stable release are needed for regression comparisons. Data quality checks are
          still available.
        </p>
      </div>
      <label class="baseline-picker"
        >Noise multiplier
        <select v-model.number="noiseMultiplier" aria-label="Noise multiplier">
          <option :value="1">1×</option>
          <option :value="2">2×</option>
          <option :value="3">3×</option>
        </select>
      </label>
    </div>
    <div class="overview-stats">
      <div>
        <span>Increases to investigate</span
        ><strong :class="{ attention: analysis.findings.length }">{{ analysis.findings.length }}</strong
        ><small>{{ priorityCount }} with separated ranges and repeated samples</small>
      </div>
      <div>
        <span>Stable → master comparisons</span><strong>{{ analysis.comparisons }}</strong
        ><small>{{ report.benchmarks.length }} benchmarks · {{ analysis.candidates }} candidates</small>
      </div>
      <div>
        <span>Measurement coverage</span><strong>{{ coverage }}</strong
        ><small>{{ analysis.covered }} of {{ analysis.expected }} candidate / metric groups recorded</small>
      </div>
    </div>
    <div class="overview-method">
      <ui-icon name="info" />
      <p>
        These are investigation leads, not confirmed regressions. A median increase must exceed
        {{ noiseMultiplier }}× the larger observed min–max spread of stable or master, with at least 3 samples
        each. This is a noise heuristic, not a significance test. Increases use medians for time, memory, and
        size. Comparing stable with master does not identify the commit that caused a change. A revision
        comparison shows endpoints to investigate, not a verified causal commit range.
      </p>
    </div>
    <p v-if="analysis.insufficient" class="overview-note">
      Insufficient evidence for {{ analysis.insufficient }} comparison(s): fewer than 3 samples on either
      side. See Data quality for details.
    </p>
    <div class="overview-controls">
      <div class="segmented" role="group" aria-label="Overview findings">
        <button :class="{ active: tab === 'increases' }" @click="tab = 'increases'">
          Increases <span>{{ analysis.findings.length }}</span>
        </button>
        <button :class="{ active: tab === 'quality' }" @click="tab = 'quality'">
          Data quality <span>{{ analysis.issues.length }}</span>
        </button>
      </div>
      <label class="overview-search"
        ><ui-icon name="search" /><input
          v-model="search"
          placeholder="Filter benchmarks or candidates..."
          aria-label="Filter overview findings"
      /></label>
    </div>
    <template v-if="tab === 'increases'">
      <p class="overview-note">
        Sorted by evidence, then median increase. Range separation is a triage heuristic, not a statistical
        significance test.
      </p>
      <article
        v-for="finding in visibleFindings"
        :key="`${finding.benchmark.name}/${finding.metric}/${finding.after.ref}`"
        class="finding-card"
      >
        <header>
          <div>
            <span class="finding-label">{{ finding.config.description || finding.metric }}</span>
            <h3>{{ finding.benchmark.description || finding.benchmark.name }}</h3>
          </div>
          <strong class="finding-delta"
            >{{ finding.delta === null ? 'From zero' : `+${finding.delta.toFixed(1)}%`
            }}<small>median increase</small></strong
          >
        </header>
        <div class="finding-evidence">
          <div class="finding-samples">
            <div v-for="(row, index) in [finding.before, finding.after]" :key="index" class="sample-row">
              <span class="sample-ref"><i :class="{ after: index === 1 }"></i>{{ row.ref }}</span>
              <strong>{{ formatValue(row.median, finding.config.unit) }}</strong>
              <span>{{ row.count }} samples</span>
              <svg
                viewBox="0 0 240 24"
                role="img"
                :aria-label="`${row.ref}: ${formatValue(row.min, finding.config.unit)} to ${formatValue(
                  row.max,
                  finding.config.unit
                )}`"
                :class="{ after: index === 1 }"
              >
                <line :x1="position(finding, row.min)" y1="12" :x2="position(finding, row.max)" y2="12" />
                <line :x1="position(finding, row.min)" y1="7" :x2="position(finding, row.min)" y2="17" />
                <line :x1="position(finding, row.max)" y1="7" :x2="position(finding, row.max)" y2="17" />
                <circle :cx="position(finding, row.median)" cy="12" r="4" />
              </svg>
            </div>
            <p class="range-caption">Observed min–max ranges on a shared scale. Dots show medians.</p>
          </div>
          <p class="overview-note">
            Observed noise: {{ formatValue(finding.noise, finding.config.unit)
            }}<template v-if="finding.noisePercent !== null">
              ({{ finding.noisePercent.toFixed(1) }}% of stable)</template
            >. Required increase: greater than {{ formatValue(finding.threshold, finding.config.unit)
            }}<template v-if="finding.thresholdPercent !== null">
              ({{ finding.thresholdPercent.toFixed(1) }}%)</template
            >, using {{ noiseMultiplier }}× noise.
          </p>
          <div class="finding-tags">
            <span :class="{ emphasized: finding.priority }">{{
              finding.separated ? 'Ranges do not overlap' : 'Ranges overlap'
            }}</span>
            <span v-if="!finding.repeated">Fewer than 3 samples</span>
            <span v-if="finding.sameRevision">Same recorded revision</span>
            <span>{{ finding.comparison.label }}</span>
          </div>
        </div>
        <footer>
          <router-link
            class="evidence-link"
            :to="
              evidenceLocation(
                report,
                run,
                finding.benchmark,
                finding.metric,
                finding.before.ref,
                finding.after.ref
              )
            "
            >Inspect measurements <ui-icon name="arrow"
          /></router-link>
          <a
            v-if="revisionComparison(report.project, finding.before.sha, finding.after.sha)"
            :href="revisionComparison(report.project, finding.before.sha, finding.after.sha)"
            target="_blank"
            rel="noopener"
            >Compare revisions <ui-icon name="external"
          /></a>
          <details>
            <summary>Evidence details</summary>
            <div class="evidence-details">
              <p v-for="row in [finding.before, finding.after]" :key="row.ref">
                <strong>{{ row.ref }}</strong
                >: median {{ formatValue(row.median, finding.config.unit) }}, range
                {{ formatValue(row.min, finding.config.unit) }}–{{
                  formatValue(row.max, finding.config.unit)
                }}, {{ row.count }} samples.<br />Candidate revision:
                <code>{{ row.sha || 'Not recorded' }}</code>
              </p>
              <p v-if="finding.comparison.changed.length">
                Changed execution fields: {{ finding.comparison.changed.join(', ') }}.
              </p>
              <p v-if="finding.comparison.missing.length">
                Missing execution fields: {{ finding.comparison.missing.join(', ') }}. Matching recorded
                fields do not establish identical conditions.
              </p>
              <p v-if="finding.sameRevision">
                Both candidates have the same recorded revision. Investigate measurement variability or
                execution differences before attributing this to code.
              </p>

              <p>
                Revision links use {{ report.project === 'buildkit' ? 'moby/buildkit' : 'docker/buildx' }}.
                Custom repository provenance and commit ancestry are not recorded in this report.
              </p>
            </div>
          </details>
        </footer>
      </article>
      <div v-if="!filteredFindings.length" class="overview-empty">
        <ui-icon name="chart" />
        <h3>
          {{
            !analysis.baseline
              ? 'No stable baseline available'
              : !analysis.target
              ? 'Master is unavailable'
              : !analysis.comparisons
              ? 'No baseline comparisons available'
              : search
              ? 'No matching increases'
              : analysis.insufficient === analysis.comparisons
              ? 'Insufficient evidence to assess increases'
              : `No median increases exceed ${noiseMultiplier}× observed noise`
          }}
        </h3>
        <p>
          Check data quality and individual charts too. Small shifts, noisy samples, and missing measurements
          can hide changes.
        </p>
      </div>
      <button v-if="filteredFindings.length > limit" class="subtle-button" @click="limit += 20">
        Show more increases ({{ filteredFindings.length - limit }} remaining)
      </button>
    </template>
    <template v-else>
      <p class="overview-note">
        Coverage checks apply to benchmarks and candidates present in this report. They cannot detect an
        entirely absent benchmark or replace CI test failures.
      </p>
      <div class="quality-list">
        <article v-for="(issue, index) in filteredIssues.slice(0, limit)" :key="index">
          <ui-icon name="info" />
          <div>
            <h3>{{ issue.benchmark?.description || issue.benchmark?.name || 'Candidate coverage' }}</h3>
            <p>{{ issue.ref }}{{ issue.metric ? ` · ${issue.metric}` : '' }}</p>
            <p>{{ issue.message }}</p>
          </div>
          <router-link
            v-if="issue.benchmark"
            :to="evidenceLocation(report, run, issue.benchmark, issue.metric, undefined, issue.ref)"
            >Inspect <ui-icon name="arrow"
          /></router-link>
        </article>
      </div>
      <div v-if="!filteredIssues.length" class="overview-empty">
        <h3>No matching data quality flags</h3>
        <p>This is a limited consistency check, not a guarantee that the run is healthy.</p>
      </div>
      <button v-if="filteredIssues.length > limit" class="subtle-button" @click="limit += 20">
        Show more flags ({{ filteredIssues.length - limit }} remaining)
      </button>
    </template>
  </section>
</template>

<script>
import UiIcon from './UiIcon.vue';
import { reportOverview, evidenceLocation, revisionComparison } from '../reports/overview.mjs';
import { formatValue } from '../reports/format.mjs';

export default {
  components: { UiIcon },
  props: { report: Object, run: String },
  data: () => ({ noiseMultiplier: 2, tab: 'increases', search: '', limit: 20 }),
  computed: {
    analysis() {
      return reportOverview(this.report, this.noiseMultiplier);
    },
    priorityCount() {
      return this.analysis.findings.filter((f) => f.priority).length;
    },
    coverage() {
      return this.analysis.expected
        ? `${Math.round((100 * this.analysis.covered) / this.analysis.expected)}%`
        : 'N/A';
    },
    filteredFindings() {
      return this.analysis.findings.filter((f) =>
        this.matches(
          `${f.benchmark.description} ${f.benchmark.name} ${f.before.ref} ${f.after.ref} ${f.metric}`
        )
      );
    },
    visibleFindings() {
      return this.filteredFindings.slice(0, this.limit);
    },
    filteredIssues() {
      return this.analysis.issues.filter((f) =>
        this.matches(`${f.benchmark?.description} ${f.benchmark?.name} ${f.ref} ${f.metric} ${f.message}`)
      );
    },
  },
  methods: {
    formatValue,
    evidenceLocation,
    revisionComparison,
    matches(text) {
      return text.toLowerCase().includes(this.search.trim().toLowerCase());
    },
    position(finding, value) {
      const min = Math.min(finding.before.min, finding.after.min);
      const max = Math.max(finding.before.max, finding.after.max);
      return max === min ? 120 : 8 + (224 * (value - min)) / (max - min);
    },
  },
};
</script>

<style scoped>
.overview-intro,
.overview-controls,
.finding-card header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
}
.overview-intro {
  margin-bottom: 24px;
}
.overview-intro p,
.overview-note {
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
  margin-top: 8px;
}
.overview-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}
.overview-stats > div {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 22px;
}
.overview-stats span,
.overview-stats small {
  display: block;
  color: var(--muted);
  font-size: 11px;
  line-height: 1.6;
}
.overview-stats strong {
  display: block;
  font-size: 32px;
  font-weight: 600;
  margin: 12px 0;
}
.attention,
.finding-delta {
  color: var(--red);
}
.overview-method {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 18px 0 24px;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
}
.overview-method svg {
  margin-top: 3px;
}
.overview-controls {
  flex-wrap: wrap;
  margin-bottom: 12px;
}
.segmented span {
  margin-left: 6px;
  color: var(--accent);
}
.overview-search {
  display: flex;
  gap: 9px;
  align-items: center;
  border: 1px solid var(--border);
  border-radius: 7px;
  padding: 10px 12px;
  color: var(--muted);
}
.overview-search input {
  background: transparent;
  border: 0;
  min-width: 220px;
  outline: none;
}
.overview-note {
  margin: 0 0 18px;
}
.finding-card {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 10px;
  margin-bottom: 16px;
  overflow: hidden;
}
.finding-card header {
  padding: 20px 24px 14px;
}
.finding-label {
  color: var(--muted);
  font-size: 11px;
  display: block;
  margin-bottom: 8px;
}
.finding-card h3 {
  font-size: 16px;
}
.finding-delta {
  font-size: 25px;
  font-weight: 600;
  white-space: nowrap;
  text-align: right;
}
.finding-delta small {
  font-size: 10px;
  color: var(--muted);
  display: block;
  font-weight: 400;
  margin-top: 5px;
}
.finding-evidence {
  padding: 8px 24px 18px;
}
.sample-row {
  display: grid;
  grid-template-columns: minmax(110px, 1fr) minmax(110px, 1fr) 80px minmax(100px, 1.3fr);
  gap: 14px;
  align-items: center;
  font-size: 12px;
  margin: 10px 0;
}
.sample-ref {
  overflow-wrap: anywhere;
}
.sample-ref i {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--purple);
  margin-right: 9px;
}
.sample-ref i.after {
  background: var(--accent);
}
.sample-row > span:nth-child(3) {
  color: var(--muted);
  font-size: 11px;
}
.sample-row svg {
  width: 100%;
  height: 24px;
  color: var(--purple);
}
.sample-row svg.after {
  color: var(--accent);
}
.sample-row svg line {
  stroke: currentColor;
  stroke-width: 2;
}
.sample-row svg circle {
  fill: currentColor;
}
.range-caption {
  color: var(--muted);
  font-size: 10px;
  text-align: right;
  margin-top: 12px;
}
.finding-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 18px;
}
.finding-tags span {
  font-size: 10px;
  border: 1px solid var(--border);
  color: var(--muted);
  border-radius: 5px;
  padding: 5px 8px;
}
.finding-tags .emphasized {
  color: var(--red);
  background: var(--red-bg);
  border-color: transparent;
}
.finding-card footer {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 16px 24px;
  border-top: 1px solid var(--border);
  padding: 16px 24px;
  font-size: 11px;
}
.finding-card a {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.finding-card a:hover,
.quality-list a:hover {
  text-decoration: underline;
}
.finding-card a svg {
  width: 14px;
  height: 14px;
}
.evidence-link {
  color: var(--accent);
}
.finding-card details {
  flex: 1;
  min-width: 160px;
  color: var(--muted);
}
.evidence-details {
  padding-top: 14px;
  line-height: 1.7;
  overflow-wrap: anywhere;
}
.evidence-details p + p {
  margin-top: 10px;
}
.overview-empty {
  text-align: center;
  border: 1px dashed var(--border);
  border-radius: 10px;
  padding: 38px 24px;
}
.overview-empty h3 {
  margin: 12px 0;
}
.overview-empty p {
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
}
.quality-list article {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 20px;
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 8px;
  margin-bottom: 10px;
}
.quality-list article > svg {
  color: var(--purple);
}
.quality-list article > div {
  flex: 1;
}
.quality-list p {
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
  margin-top: 6px;
}
.quality-list a {
  color: var(--accent);
  font-size: 12px;
  white-space: nowrap;
}
@media (max-width: 1100px) {
  .overview-stats {
    gap: 10px;
  }
  .overview-stats > div {
    padding: 16px;
  }
  .sample-row {
    grid-template-columns: 1fr 1fr 75px;
  }
  .sample-row svg {
    grid-column: 1 / -1;
  }
}
@media (max-width: 600px) {
  .overview-intro {
    align-items: flex-start;
    flex-direction: column;
  }
  .overview-stats {
    grid-template-columns: 1fr;
  }
  .overview-stats strong {
    margin: 6px 0;
  }
  .overview-search {
    width: 100%;
  }
  .overview-search input {
    min-width: 0;
    width: 100%;
  }
  .finding-card header,
  .finding-evidence,
  .finding-card footer {
    padding: 16px;
  }
  .finding-card h3 {
    font-size: 14px;
  }
  .sample-row {
    gap: 8px;
    grid-template-columns: 1fr 1fr;
  }
  .finding-delta {
    font-size: 21px;
  }
  .sample-row svg {
    grid-column: 2;
  }
  .quality-list article {
    flex-wrap: wrap;
  }
}
</style>
